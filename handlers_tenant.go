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
	}
	var ti TenantInfo
	err := db.DB.QueryRow(`
		SELECT t.id, t.name, t.phone, COALESCE(t.email,''),
			COALESCE(ra.room_id,''), COALESCE(r.room_number,''),
			COALESCE(t.check_in_date,''),
			COALESCE(ra.rent_amount,0)
		FROM tenants t
		LEFT JOIN room_assignments ra ON t.id = ra.tenant_id AND ra.is_active IS TRUE
		LEFT JOIN rooms r ON ra.room_id = r.id
		WHERE t.id = $1`, tenantID,
	).Scan(&ti.ID, &ti.Name, &ti.Phone, &ti.Email, &ti.RoomID, &ti.RoomNumber, &ti.CheckInDate, &ti.RentAmount)
	if err != nil {
		http.Error(w, "Tenant not found", 404)
		return
	}

	var verificationStatus, lpuPhoto, aadharPhoto string
	db.DB.QueryRow(`
		SELECT COALESCE(status,'not_submitted'), COALESCE(lpu_id_photo,''), COALESCE(aadhar_photo,'')
		FROM tenant_verifications WHERE tenant_id = $1`, tenantID,
	).Scan(&verificationStatus, &lpuPhoto, &aadharPhoto)
	if verificationStatus == "" {
		verificationStatus = "not_submitted"
	}

	var passID, passNumber string
	db.DB.QueryRow(`
		SELECT COALESCE(id,''), COALESCE(pass_number,'')
		FROM tenant_passes WHERE tenant_id = $1 AND is_active = 1`, tenantID,
	).Scan(&passID, &passNumber)

	var prevMonthDue float64
	prevMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount),0) FROM payments
		WHERE tenant_id = $1 AND month_covered = $2 AND status = 'pending'`,
		tenantID, prevMonth,
	).Scan(&prevMonthDue)

	var totalPending float64
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount),0) FROM payments
		WHERE tenant_id = $1 AND status = 'pending'`,
		tenantID,
	).Scan(&totalPending)

	render(w, "tenant_dashboard.html", map[string]interface{}{
		"Tenant":             ti,
		"VerificationStatus": verificationStatus,
		"LpuPhoto":           lpuPhoto,
		"AadharPhoto":        aadharPhoto,
		"PassID":             passID,
		"PassNumber":         passNumber,
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

	var existingID string
	db.DB.QueryRow("SELECT id FROM tenant_verifications WHERE tenant_id = $1", tenantID).Scan(&existingID)

	if existingID != "" {
		if lpuPath != "" {
			db.DB.Exec("UPDATE tenant_verifications SET lpu_id_photo = $1, updated_at = NOW() WHERE tenant_id = $2", lpuPath, tenantID)
		}
		if aadharPath != "" {
			db.DB.Exec("UPDATE tenant_verifications SET aadhar_photo = $1, updated_at = NOW() WHERE tenant_id = $2", aadharPath, tenantID)
		}
		db.DB.Exec(`UPDATE tenant_verifications SET status = 'pending', submitted_at = NOW(), updated_at = NOW() 
			WHERE tenant_id = $1 AND (lpu_id_photo IS NOT NULL OR aadhar_photo IS NOT NULL)`, tenantID)
	} else {
		if lpuPath == "" {
			lpuPath = ""
		}
		if aadharPath == "" {
			aadharPath = ""
		}
		id := fmt.Sprintf("VER%d", time.Now().UnixNano())
		status := "pending"
		if lpuPath == "" && aadharPath == "" {
			status = "not_submitted"
		}
		db.DB.Exec(`INSERT INTO tenant_verifications (id, tenant_id, lpu_id_photo, aadhar_photo, status, submitted_at)
			VALUES ($1, $2, $3, $4, $5, NOW())`, id, tenantID, lpuPath, aadharPath, status)
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

	render(w, "tenant_payments.html", map[string]interface{}{
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
