package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
		DiscountAmount                                                    float64
		DiscountNote                                                      string
		Status                                                             string
	}
	var currentBill BillInfo
	db.DB.QueryRow(`
		SELECT COALESCE(rent_amount, 0), COALESCE(maintenance_amount, 0),
			COALESCE(electricity_amount, 0), COALESCE(water_amount, 0),
			COALESCE(discount_amount, 0), COALESCE(discount_note, ''), COALESCE(total_amount, 0), COALESCE(status, 'not_generated')
		FROM bills WHERE tenant_id = $1 AND billing_month = $2`,
		tenantID, currentMonth,
	).Scan(&currentBill.RentAmount, &currentBill.MaintenanceAmt, &currentBill.ElectricityAmt,
		&currentBill.WaterAmt, &currentBill.DiscountAmount, &currentBill.DiscountNote,
		&currentBill.TotalAmount, &currentBill.Status)

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

	// Electricity rate
	var ratePerUnit float64 = 12
	var unitRate string
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'electricity_rate'").Scan(&unitRate)
	if unitRate != "" {
		if r, err := strconv.ParseFloat(unitRate, 64); err == nil {
			ratePerUnit = r
		}
	}

	// Meter readings for tenant'''s room (current month)
	type MeterReading struct {
		MeterType, MeterNumber         string
		InitialReading, CurrentReading int
		Units                          int
		RatePerUnit                    float64
		Amount                         float64
	}
	var meterReadings []MeterReading
	if ti.RoomID != "" {
		roomRows, mrErr := db.DB.Query(`
			SELECT m.meter_type, m.meter_number,
				COALESCE(mr.initial_reading, 0), COALESCE(mr.current_reading, 0),
				COALESCE(mr.current_reading, 0) - COALESCE(mr.initial_reading, 0) as units
			FROM meters m
			JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
			WHERE m.room_id = $2 AND m.is_active = true AND m.room_id != 'BUILDING'
			ORDER BY m.meter_type`, currentMonth, ti.RoomID)
		if mrErr == nil {
			defer roomRows.Close()
			for roomRows.Next() {
				var mr MeterReading
				roomRows.Scan(&mr.MeterType, &mr.MeterNumber, &mr.InitialReading, &mr.CurrentReading, &mr.Units)
				mr.RatePerUnit = ratePerUnit
				mr.Amount = float64(mr.Units) * ratePerUnit
				meterReadings = append(meterReadings, mr)
			}
		}
	}
	if meterReadings == nil {
		meterReadings = []MeterReading{}
	}

	// Water per-room amount
	var waterPerRoom float64
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'water_per_room_' || $1", currentMonth).Scan(&waterPerRoom)

	// Recent payments
	type PaymentCard struct {
		ID, Date, Method, Status, MonthCovered, Notes string
		Amount, LateFee                               float64
	}
	var recentPayments []PaymentCard
	payRows, payErr := db.DB.Query(`
		SELECT id, amount, COALESCE(payment_date::text,''), COALESCE(payment_method,''), COALESCE(status,''), COALESCE(month_covered,''), COALESCE(late_fee,0), COALESCE(notes,'')
		FROM payments WHERE tenant_id = $1 ORDER BY payment_date DESC LIMIT 5`, tenantID)
	if payErr == nil {
		defer payRows.Close()
		for payRows.Next() {
			var p PaymentCard
			payRows.Scan(&p.ID, &p.Amount, &p.Date, &p.Method, &p.Status, &p.MonthCovered, &p.LateFee, &p.Notes)
			recentPayments = append(recentPayments, p)
		}
	}
	if recentPayments == nil {
		recentPayments = []PaymentCard{}
	}

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
		"MeterReadings":      meterReadings,
		"RatePerUnit":        ratePerUnit,
		"WaterPerRoom":       waterPerRoom,
		"RecentPayments":     recentPayments,
		"Msg":                msg,
		"Error":              errMsg,
		"Active":             "dashboard",
	})
}

func handleTenantVerificationUpload(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Redirect(w, r, "/tenant/dashboard?error=Upload+failed:+file+too+large+(max+10MB)", http.StatusSeeOther)
		return
	}
	lpuFile, lpuHeader, lpuErr := r.FormFile("lpu_id_photo")
	aadharFile, aadharHeader, aadharErr := r.FormFile("aadhar_photo")

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
		if _, copyErr := io.Copy(dst, file); copyErr != nil {
			return ""
		}
		return "/uploads/" + filename
	}

	var lpuPath, aadharPath string
	if lpuFile != nil && lpuErr == nil {
		defer lpuFile.Close()
		contentType := ""
		if lpuHeader != nil {
			contentType = lpuHeader.Header.Get("Content-Type")
		}
		hdr := make(http.Header)
		hdr.Set("Content-Type", contentType)
		lpuPath = saveFile(lpuFile, &hdr, "lpu")
	} else if lpuErr != nil {
		http.Redirect(w, r, "/tenant/dashboard?error=Upload+failed:+could+not+read+file", http.StatusSeeOther)
		return
	}
	if aadharFile != nil && aadharErr == nil {
		defer aadharFile.Close()
		contentType := ""
		if aadharHeader != nil {
			contentType = aadharHeader.Header.Get("Content-Type")
		}
		hdr := make(http.Header)
		hdr.Set("Content-Type", contentType)
		aadharPath = saveFile(aadharFile, &hdr, "aadhar")
	} else if aadharErr != nil {
		http.Redirect(w, r, "/tenant/dashboard?error=Upload+failed:+could+not+read+file", http.StatusSeeOther)
		return
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

	msg := r.URL.Query().Get("msg")
	errMsg := r.URL.Query().Get("error")

	// Current pending bills
	currentMonth := time.Now().Format("2006-01")
	type PendingBill struct {
		Month       string
		TotalAmount float64
		Status      string
	}
	var pendingBills []PendingBill
	billRows, billErr := db.DB.Query(`
		SELECT billing_month, total_amount, COALESCE(status,'pending')
		FROM bills WHERE tenant_id = $1 AND status != 'paid'
		ORDER BY billing_month DESC LIMIT 3`, tenantID)
	if billErr == nil {
		defer billRows.Close()
		for billRows.Next() {
			var pb PendingBill
			billRows.Scan(&pb.Month, &pb.TotalAmount, &pb.Status)
			pendingBills = append(pendingBills, pb)
		}
	}
	if pendingBills == nil {
		pendingBills = []PendingBill{}
	}

	// Payment history
	rows, err := db.DB.Query(`
		SELECT id, amount, COALESCE(payment_date::text,''), COALESCE(payment_method,''),
			COALESCE(status,'pending'), COALESCE(month_covered,''), COALESCE(late_fee,0), COALESCE(notes,''),
			COALESCE(screenshot,'')
		FROM payments WHERE tenant_id = $1 ORDER BY payment_date DESC LIMIT 50`, tenantID)
	if err != nil {
		http.Error(w, "Failed to load payments", 500)
		return
	}
	defer rows.Close()

	type TenantPayment struct {
		ID, Date, Method, Status, MonthCovered, Notes string
		Amount, LateFee                                float64
		Screenshot                                     string
	}
	var payments []TenantPayment
	for rows.Next() {
		var p TenantPayment
		rows.Scan(&p.ID, &p.Amount, &p.Date, &p.Method, &p.Status, &p.MonthCovered, &p.LateFee, &p.Notes, &p.Screenshot)
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []TenantPayment{}
	}

	// Check if any payment is currently verifying
	hasVerifying := false
	for _, p := range payments {
		if p.Status == "paid_by_user" {
			hasVerifying = true
			break
		}
	}

	renderPrivate(w, "tenant_payments.html", map[string]interface{}{
		"Payments":      payments,
		"PendingBills":  pendingBills,
		"CurrentMonth":  currentMonth,
		"HasVerifying":  hasVerifying,
		"Msg":           msg,
		"Error":         errMsg,
		"Active":        "payments",
	})
}



func handleTenantMarkPaid(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Redirect(w, r, "/tenant/payments?error=File+too+large+(max+10MB)", http.StatusSeeOther)
		return
	}

	monthCovered := r.FormValue("month_covered")
	amountStr := r.FormValue("amount")

	// Parse amount
	amount := 0.0
	if amountStr != "" {
		if a, err := strconv.ParseFloat(amountStr, 64); err == nil {
			amount = a
		}
	}

	// Handle screenshot upload
	screenshotFile, screenshotHeader, scrErr := r.FormFile("screenshot")
	var screenshotPath string
	if screenshotFile != nil && scrErr == nil {
		defer screenshotFile.Close()
		ext := ".jpg"
		if screenshotHeader != nil {
			ct := screenshotHeader.Header.Get("Content-Type")
			if strings.Contains(ct, "png") {
				ext = ".png"
			} else if strings.Contains(ct, "webp") {
				ext = ".webp"
			}
		}
		filename := fmt.Sprintf("pay_%s_%d%s", tenantID, time.Now().UnixNano(), ext)
		dst, err := os.Create(filepath.Join("uploads", filename))
		if err == nil {
			defer dst.Close()
			io.Copy(dst, screenshotFile)
			screenshotPath = "/uploads/" + filename
		}
	}

	// Create payment record
	paymentID := fmt.Sprintf("PAY%d", time.Now().UnixNano())
	if monthCovered == "" {
		monthCovered = time.Now().Format("2006-01")
	}

	_, err := db.DB.Exec(`
		INSERT INTO payments (id, tenant_id, amount, payment_date, payment_method, status, month_covered, screenshot)
		VALUES ($1, $2, $3, $4, 'UPI', 'paid_by_user', $5, $6)`,
		paymentID, tenantID, amount, time.Now().Format("2006-01-02"), monthCovered, screenshotPath)
	if err != nil {
		log.Printf("ERROR creating payment: %v", err)
		http.Redirect(w, r, "/tenant/payments?error=Failed+to+submit+payment", http.StatusSeeOther)
		return
	}

	// Update bill status if this payment covers a bill
	db.DB.Exec("UPDATE bills SET status = 'paid_by_user' WHERE tenant_id = $1 AND billing_month = $2 AND status = 'pending'", tenantID, monthCovered)

	http.Redirect(w, r, "/tenant/payments?msg=Payment+submitted!+Admin+will+verify+shortly", http.StatusSeeOther)
}

func handleTenantMeterDetails(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	var roomID string
	db.DB.QueryRow("SELECT COALESCE(room_id,'') FROM tenants WHERE id = $1", tenantID).Scan(&roomID)

	// Rate
	var ratePerUnit float64 = 12
	var unitRate string
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'electricity_rate'").Scan(&unitRate)
	if unitRate != "" {
		if r, err := strconv.ParseFloat(unitRate, 64); err == nil {
			ratePerUnit = r
		}
	}

	type MeterDetail struct {
		MeterType, MeterNumber          string
		InitialReading, CurrentReading  int
		Units                           int
		RatePerUnit                     float64
		Amount                          float64
	}
	var meters []MeterDetail
	rows, err := db.DB.Query(`
		SELECT m.meter_type, m.meter_number,
			COALESCE(mr.initial_reading, 0), COALESCE(mr.current_reading, 0),
			COALESCE(mr.current_reading, 0) - COALESCE(mr.initial_reading, 0) as units
		FROM meters m
		JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
		WHERE m.room_id = $2 AND m.is_active = true AND m.room_id != 'BUILDING'
		ORDER BY m.meter_type`, month, roomID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var md MeterDetail
			rows.Scan(&md.MeterType, &md.MeterNumber, &md.InitialReading, &md.CurrentReading, &md.Units)
			md.RatePerUnit = ratePerUnit
			md.Amount = float64(md.Units) * ratePerUnit
			meters = append(meters, md)
		}
	}
	if meters == nil {
		meters = []MeterDetail{}
	}

	// Water share
	var waterPerRoom float64
	var waterUnits int
	var occupiedCount int
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'water_per_room_' || $1", month).Scan(&waterPerRoom)
	db.DB.QueryRow("SELECT COUNT(*) FROM tenants WHERE status = 'active' AND room_id IS NOT NULL AND (end_date IS NULL OR end_date = '')").Scan(&occupiedCount)
	if waterPerRoom == 0 && occupiedCount > 0 {
		var buildingUnits int
		db.DB.QueryRow(`
			SELECT COALESCE(SUM(mr.current_reading - mr.initial_reading), 0)
			FROM meters m
			JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
			WHERE m.room_id = 'BUILDING' AND m.meter_type = 'Water' AND m.is_active = true`, month).Scan(&buildingUnits)
		waterUnits = buildingUnits
		if occupiedCount > 0 {
			waterPerRoom = (float64(buildingUnits) / float64(occupiedCount)) * ratePerUnit
		}
	} else if occupiedCount > 0 {
		waterUnits = int(waterPerRoom / ratePerUnit * float64(occupiedCount))
	}

	monthLabel := month
	if t, err := time.Parse("2006-01", month); err == nil {
		monthLabel = t.Format("January 2006")
	}

	renderPrivate(w, "tenant_meter_details.html", map[string]interface{}{
		"Meters":        meters,
		"Month":         month,
		"MonthLabel":    monthLabel,
		"RatePerUnit":   ratePerUnit,
		"WaterPerRoom":  waterPerRoom,
		"WaterUnits":    waterUnits,
		"OccupiedCount": occupiedCount,
		"Active":        "payments",
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
	if !strings.HasPrefix(filePath, "/uploads/") {
		http.NotFound(w, r)
		return
	}

	// Check admin session
	if adminCookie, err := r.Cookie("session"); err == nil && adminCookie.Value != "" {
		var role string
		db.DB.QueryRow("SELECT role FROM users WHERE username = $1", adminCookie.Value).Scan(&role)
		if role == "admin" {
			http.ServeFile(w, r, filePath[1:])
			return
		}
	}

	// Check tenant session — only serve if the file belongs to this tenant
	if tenantCookie, err := r.Cookie("tenant_session"); err == nil && tenantCookie.Value != "" {
		parts := strings.SplitN(tenantCookie.Value, ":", 2)
		if len(parts) == 2 && parts[0] == "tenant" {
			tenantID := parts[1]
			// Filename format: {prefix}_{tenantID}_{timestamp}.{ext}
			filename := filepath.Base(filePath)
			fileParts := strings.SplitN(filename, "_", 3)
			if len(fileParts) >= 2 && fileParts[1] == tenantID {
				http.ServeFile(w, r, filePath[1:])
				return
			}
		}
	}

	http.NotFound(w, r)
}
