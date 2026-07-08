package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vatsapartment-go/db"
	"golang.org/x/crypto/bcrypt"
)

func requireTenant(w http.ResponseWriter, r *http.Request) (string, bool) {
	cookie, err := r.Cookie("tenant_session")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login?mode=tenant", http.StatusSeeOther)
		return "", false
	}
	parts := strings.SplitN(cookie.Value, ":", 2)
	if len(parts) != 2 || parts[0] != "tenant" {
		http.Redirect(w, r, "/login?mode=tenant", http.StatusSeeOther)
		return "", false
	}
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM tenants WHERE id = $1 AND status = 'active'", parts[1]).Scan(&count)
	if count == 0 {
		http.Redirect(w, r, "/login?mode=tenant", http.StatusSeeOther)
		return "", false
	}
	return parts[1], true
}

func handleTenantDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	msg := r.URL.Query().Get("msg")
	errMsg := r.URL.Query().Get("error")

	type TenantInfo struct {
		ID, Name, Phone, Email, RoomID, RoomNumber, CheckInDate string
		RentAmount                                               float64
		PassNumber                                               string
		VerificationStatus                                       string
		LpuPhoto, AadharPhoto                                   string
	}
	var ti TenantInfo
	err := db.DB.QueryRow(`
		SELECT t.id, t.name, t.phone, COALESCE(t.email,''),
			COALESCE(t.room_id,''), COALESCE(r.room_number,''),
			COALESCE(t.check_in_date::text,''),
			COALESCE(t.rent_amount,0),
			COALESCE(t.pass_number,''),
			COALESCE(t.verification_status,'not_submitted'),
			COALESCE(t.lpu_id_photo,''), COALESCE(t.aadhar_photo,'')
		FROM tenants t
		LEFT JOIN rooms r ON t.room_id = r.id
		WHERE t.id = $1`, tenantID,
	).Scan(&ti.ID, &ti.Name, &ti.Phone, &ti.Email, &ti.RoomID, &ti.RoomNumber, &ti.CheckInDate, &ti.RentAmount,
		&ti.PassNumber, &ti.VerificationStatus, &ti.LpuPhoto, &ti.AadharPhoto)
	if err != nil {
		http.Error(w, "Tenant not found", 404)
		return
	}

	verificationStatus := ti.VerificationStatus
	lpuPhoto := ti.LpuPhoto
	aadharPhoto := ti.AadharPhoto
	if verificationStatus == "" {
		verificationStatus = "not_submitted"
	}

	// Pass data from tenant query above
	passNumber := ti.PassNumber

	currentMonth := time.Now().Format("2006-01")
	type BillInfo struct {
		RentAmount, MaintenanceAmt, ElectricityAmt, WaterAmt, TotalAmount float64
		Status                                                             string
	}
	var currentBill BillInfo
	db.DB.QueryRow(`
		SELECT COALESCE(rent_amount, 0), COALESCE(maintenance_amount, 0),
			COALESCE(electricity_amount, 0), COALESCE(water_amount, 0),
			COALESCE(total_amount, 0), COALESCE(status, 'not_generated')
		FROM bills WHERE tenant_id = $1 AND billing_month = $2`,
		tenantID, currentMonth,
	).Scan(&currentBill.RentAmount, &currentBill.MaintenanceAmt, &currentBill.ElectricityAmt,
		&currentBill.WaterAmt, &currentBill.TotalAmount, &currentBill.Status)

	var prevMonthDue float64
	prevMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(total_amount),0) FROM bills
		WHERE tenant_id = $1 AND billing_month = $2 AND status = 'pending'`,
		tenantID, prevMonth,
	).Scan(&prevMonthDue)

	var totalPending float64
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(total_amount),0) FROM bills
		WHERE tenant_id = $1 AND status = 'pending'`,
		tenantID,
	).Scan(&totalPending)

	renderPrivate(w, "tenant_dashboard.html", map[string]interface{}{
		"Tenant":             ti,
		"VerificationStatus": verificationStatus,
		"LpuPhoto":           lpuPhoto,
		"AadharPhoto":        aadharPhoto,
		"PassNumber":         passNumber,
		"CurrentBill":        currentBill,
		"CurrentMonth":       currentMonth,
		"PrevMonthDue":       prevMonthDue,
		"TotalPending":       totalPending,
		"PrevMonth":          prevMonth,
		"Msg":                msg,
		"Error":              errMsg,
	})
}

func handleTenantVerificationUpload(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	r.ParseMultipartForm(10 << 20)
	lpuFile, lpuHeader, _ := r.FormFile("lpu_id_photo")
	aadharFile, aadharHeader, _ := r.FormFile("aadhar_photo")

	saveFile := func(file io.Reader, header *http.Header, prefix string) string {
		if file == nil {
			return ""
		}
		ext := ".jpg"
		if header != nil {
			contentType := header.Get("Content-Type")
			if strings.Contains(contentType, "png") {
				ext = ".png"
			} else if strings.Contains(contentType, "webp") {
				ext = ".webp"
			}
		}
		filename := fmt.Sprintf("%s_%s_%d%s", prefix, tenantID, time.Now().UnixNano(), ext)
		dst, err := os.Create(filepath.Join("uploads", filename))
		if err != nil {
			return ""
		}
		defer dst.Close()
		io.Copy(dst, file)
		return "/uploads/" + filename
	}

	var lpuPath, aadharPath string
	if lpuFile != nil {
		defer lpuFile.Close()
		contentType := ""
		if lpuHeader != nil {
			contentType = lpuHeader.Header.Get("Content-Type")
		}
		hdr := make(http.Header)
		hdr.Set("Content-Type", contentType)
		lpuPath = saveFile(lpuFile, &hdr, "lpu")
	}
	if aadharFile != nil {
		defer aadharFile.Close()
		contentType := ""
		if aadharHeader != nil {
			contentType = aadharHeader.Header.Get("Content-Type")
		}
		hdr := make(http.Header)
		hdr.Set("Content-Type", contentType)
		aadharPath = saveFile(aadharFile, &hdr, "aadhar")
	}

	if lpuPath != "" {
		db.DB.Exec("UPDATE tenants SET lpu_id_photo=$1, verification_status='pending', verification_submitted_at=NOW() WHERE id=$2", lpuPath, tenantID)
	}
	if aadharPath != "" {
		db.DB.Exec("UPDATE tenants SET aadhar_photo=$1, verification_status='pending', verification_submitted_at=NOW() WHERE id=$2", aadharPath, tenantID)
	}

	http.Redirect(w, r, "/tenant/dashboard", http.StatusSeeOther)
}

func handleTenantLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "tenant_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handleTenantSetPassword(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	newPass := r.FormValue("new_password")
	if len(newPass) < 4 {
		http.Redirect(w, r, "/tenant/dashboard?error=Password+must+be+at+least+4+characters", http.StatusSeeOther)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		http.Redirect(w, r, "/tenant/dashboard?error=Failed+to+set+password", http.StatusSeeOther)
		return
	}

	db.DB.Exec("UPDATE tenants SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hash), tenantID)
	http.Redirect(w, r, "/tenant/dashboard?msg=Password+set+successfully", http.StatusSeeOther)
}

func handleTenantPayments(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, amount, payment_date, payment_method, status, month_covered, late_fee, notes
		FROM payments WHERE tenant_id = $1 ORDER BY payment_date DESC LIMIT 50`, tenantID)
	if err != nil {
		http.Error(w, "Failed to load payments", 500)
		return
	}
	defer rows.Close()

	type TenantPayment struct {
		ID, Date, Method, Status, MonthCovered, Notes string
		Amount, LateFee                                float64
	}
	var payments []TenantPayment
	for rows.Next() {
		var p TenantPayment
		rows.Scan(&p.ID, &p.Amount, &p.Date, &p.Method, &p.Status, &p.MonthCovered, &p.LateFee, &p.Notes)
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []TenantPayment{}
	}

	renderPrivate(w, "tenant_payments.html", map[string]interface{}{
		"Payments": payments,
		"Active":   "payments",
	})
}

func getTenantIDFromSession(r *http.Request) string {
	cookie, err := r.Cookie("tenant_session")
	if err != nil {
		return ""
	}
	parts := strings.SplitN(cookie.Value, ":", 2)
	if len(parts) == 2 && parts[0] == "tenant" {
		return parts[1]
	}
	return ""
}

func handleTenantUploadedFile(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path
	if strings.HasPrefix(filePath, "/uploads/") {
		http.ServeFile(w, r, filePath[1:])
		return
	}
	http.NotFound(w, r)
}
