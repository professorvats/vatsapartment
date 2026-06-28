package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"vatsapartment-go/db"
	"net/url"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// ─── Dashboard ────────────────────────────────────────────────

func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	stats := getDashboardStats()
	renderPrivate(w, "admin_dashboard.html", map[string]interface{}{"Stats": stats, "Active": "dashboard", "Title": "Dashboard"})
}

type DashboardStats struct {
	TotalRooms      int     `json:"totalRooms"`
	OccupiedRooms   int     `json:"occupiedRooms"`
	TotalTenants    int     `json:"totalTenants"`
	ActiveTenants   int     `json:"activeTenants"`
	MonthlyRevenue  float64 `json:"monthlyRevenue"`
	CollectionRate  float64 `json:"collectionRate"`
	PendingPayments int     `json:"pendingPayments"`
}

func getDashboardStats() DashboardStats {
	var s DashboardStats
	var total, completed float64
	var wg sync.WaitGroup
	wg.Add(8)

	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COUNT(*) FROM rooms").Scan(&s.TotalRooms) }()
	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COUNT(*) FROM bookings WHERE status = 'active'").Scan(&s.OccupiedRooms) }()
	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COUNT(*) FROM tenants").Scan(&s.TotalTenants) }()
	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COUNT(*) FROM tenants WHERE status = 'active'").Scan(&s.ActiveTenants) }()
	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed' AND payment_date >= date_trunc('month', CURRENT_DATE)").Scan(&s.MonthlyRevenue) }()
	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COUNT(*) FROM payments WHERE status = 'pending'").Scan(&s.PendingPayments) }()
	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE payment_date >= date_trunc('month', CURRENT_DATE)").Scan(&total) }()
	go func() { defer wg.Done(); db.DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed' AND payment_date >= date_trunc('month', CURRENT_DATE)").Scan(&completed) }()

	wg.Wait()
	if total > 0 {
		s.CollectionRate = (completed / total) * 100
	}
	return s
}

// ─── Rooms ─────────────────────────────────────────────────────

func handleAdminRooms(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	rows, err := db.DB.Query(`SELECT r.id, r.room_number, r.floor, r.type, r.price,
		COALESCE(b.status, 'available') as status
		FROM rooms r
		LEFT JOIN bookings b ON r.id = b.room_id AND b.status = 'active'
		ORDER BY r.floor, r.id`)
	if err != nil {
		log.Printf("ERROR loading rooms: %v", err)
		renderAdminError(w, "Failed to load rooms")
		return
	}
	defer rows.Close()

	type RoomAdmin struct {
		ID, RoomNumber, RoomType, Status string
		Floor                            int
		Price                            float64
	}
	var rooms []RoomAdmin
	for rows.Next() {
		var r RoomAdmin
		if err := rows.Scan(&r.ID, &r.RoomNumber, &r.Floor, &r.RoomType, &r.Price, &r.Status); err != nil {
			log.Printf("ERROR scanning room row: %v", err)
			continue
		}
		rooms = append(rooms, r)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating room rows: %v", err)
		renderAdminError(w, "Data error loading rooms")
		return
	}
	renderPrivate(w, "admin_rooms.html", map[string]interface{}{"Rooms": rooms, "Active": "rooms", "Title": "Manage Rooms"})
}

func handleAdminRoomsSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	action := r.FormValue("action")
	id := r.FormValue("id")

	if action == "edit" {
		roomNum := r.FormValue("room_number")
		roomType := r.FormValue("type")
		floor, _ := strconv.Atoi(r.FormValue("floor"))
		price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
		db.DB.Exec(`UPDATE rooms SET room_number=$1, type=$2, floor=$3, price=$4, updated_at=NOW() WHERE id=$5`,
			roomNum, roomType, floor, price, id)
		http.Redirect(w, r, "/admin/rooms?msg=Room+updated", http.StatusSeeOther)
		return
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	status := r.FormValue("status")

	if price > 0 {
		db.DB.Exec("UPDATE rooms SET price = $1, updated_at = NOW() WHERE id = $2", price, id)
	}
	if status == "available" {
		db.DB.Exec("UPDATE bookings SET status = 'ended', updated_at = NOW() WHERE room_id = $1 AND status = 'active'", id)
	} else if status == "occupied" {
		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM bookings WHERE room_id = $1 AND status = 'active'", id).Scan(&count)
		if count == 0 {
			db.DB.Exec(`INSERT INTO bookings (id, room_id, rent_amount, check_in_date, status)
				VALUES ($1, $2, (SELECT price FROM rooms WHERE id = $2), $3, 'active')`,
				fmt.Sprintf("BK%d", time.Now().UnixNano()), id, time.Now().Format("2006-01-02"))
		}
	}
	http.Redirect(w, r, "/admin/rooms", http.StatusSeeOther)
}

func handleAdminRoomAdd(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	roomNum := r.FormValue("room_number")
	roomType := r.FormValue("type")
	floor, _ := strconv.Atoi(r.FormValue("floor"))
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	id := roomNum

	db.DB.Exec(`INSERT INTO rooms (id, room_number, floor, type, price) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET room_number=$2, floor=$3, type=$4, price=$5, updated_at=NOW()`,
		id, roomNum, floor, roomType, price)
	http.Redirect(w, r, "/admin/rooms", http.StatusSeeOther)
}

// ─── Tenants ───────────────────────────────────────────────────

func handleAdminTenants(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	rows, err := db.DB.Query(`SELECT t.id, t.name, t.email, t.phone, t.status, COALESCE(t.check_in_date::text, '') as check_in_date,
		COALESCE(t.security_deposit, 0) as security_deposit, COALESCE(t.security_lock_in_period, 0) as security_lock_in_period,
		COALESCE(ra.room_id, '') as room_id, COALESCE(r.room_number, '') as room_number,
		COALESCE(ra.rent_amount, 0) as rent_amount, COALESCE(ra.start_date::text, '') as start_date,
		COALESCE(t.password_hash, '') as password_hash,
		COALESCE(tv.status, 'not_submitted') as ver_status
		FROM tenants t
		LEFT JOIN room_assignments ra ON t.id = ra.tenant_id AND ra.is_active
		LEFT JOIN rooms r ON ra.room_id = r.id
		LEFT JOIN tenant_verifications tv ON t.id = tv.tenant_id
		ORDER BY t.name`)
	if err != nil {
		log.Printf("ERROR loading tenants: %v", err); renderAdminError(w, "Failed to load tenants")
		return
	}
	defer rows.Close()

	type TenantAdmin struct {
		ID, Name, Email, Phone, Status, CheckInDate, RoomID, RoomNumber string
		SecurityDeposit                                                  float64
		LockInPeriod                                                     int
		RentAmount                                                       float64
		StartDate                                                        string
		HasPassword                                                      bool
		VerificationStatus                                                string
	}
	var tenants []TenantAdmin
	for rows.Next() {
		var t TenantAdmin
		var pwHash, verStatus string
		if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Phone, &t.Status, &t.CheckInDate,
			&t.SecurityDeposit, &t.LockInPeriod, &t.RoomID, &t.RoomNumber, &t.RentAmount, &t.StartDate,
			&pwHash, &verStatus); err != nil {
			log.Printf("ERROR scanning tenant row: %v", err)
			continue
		}
		t.HasPassword = pwHash != ""
		t.VerificationStatus = verStatus
		tenants = append(tenants, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating tenant rows: %v", err)
		renderAdminError(w, "Data error loading tenants")
		return
	}
	if tenants == nil {
		tenants = []TenantAdmin{}
	}
	// Rooms for assign dropdown
	type RoomOpt struct{ ID, Number string }
	var roomOpts []RoomOpt
	roomRows, err := db.DB.Query("SELECT id, room_number FROM rooms ORDER BY room_number")
	if err != nil {
		log.Printf("ERROR loading room options: %v", err)
	} else {
		defer roomRows.Close()
		for roomRows.Next() {
			var r RoomOpt
			if err := roomRows.Scan(&r.ID, &r.Number); err != nil {
				log.Printf("ERROR scanning room option: %v", err)
				continue
			}
			roomOpts = append(roomOpts, r)
		}
	}

	renderPrivate(w, "admin_tenants.html", map[string]interface{}{
		"Tenants": tenants,
		"Rooms":   roomOpts,
		"Active":  "tenants",
		"Title":   "Manage Tenants",
	})
}

func handleAdminTenantsSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	id := r.FormValue("id")
	action := r.FormValue("action")

		if action == "add" {
			name := r.FormValue("name")
			phone := r.FormValue("phone")
			email := r.FormValue("email")
			status := r.FormValue("status")
			if status == "" {
				status = "active"
			}
			checkInDate := r.FormValue("check_in_date")
			secDeposit, _ := strconv.ParseFloat(r.FormValue("security_deposit"), 64)
			lockIn, _ := strconv.Atoi(r.FormValue("lock_in_period"))
			newPass := r.FormValue("password")

			// Handle empty email as NULL to avoid unique constraint violation
			var emailPtr interface{}
			if email == "" {
				emailPtr = nil
			} else {
				emailPtr = email
			}

			id := fmt.Sprintf("T%03d", time.Now().UnixNano()%100000)

			if checkInDate != "" {
				_, err := db.DB.Exec(`INSERT INTO tenants (id, name, phone, email, status, check_in_date, security_deposit, security_lock_in_period)
					VALUES ($1, $2, $3, $4, $5, $6::timestamp, $7, $8)`,
					id, name, phone, emailPtr, status, checkInDate, secDeposit, lockIn)
				if err != nil {
					log.Printf("ERROR adding tenant: %v", err)
					renderAdminError(w, "Failed to add tenant: "+friendlyDBError(err))
					return
				}
			} else {
				_, err := db.DB.Exec(`INSERT INTO tenants (id, name, phone, email, status, security_deposit, security_lock_in_period)
					VALUES ($1, $2, $3, $4, $5, $6, $7)`,
					id, name, phone, emailPtr, status, secDeposit, lockIn)
				if err != nil {
					log.Printf("ERROR adding tenant: %v", err)
					renderAdminError(w, "Failed to add tenant: "+friendlyDBError(err))
					return
				}
			}

			if newPass != "" {
				hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
				if err == nil {
					db.DB.Exec("UPDATE tenants SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hash), id)
				}
			}

			// Room assignment
			roomID := r.FormValue("room_id")
			rent, _ := strconv.ParseFloat(r.FormValue("rent"), 64)
			startDate := r.FormValue("start_date")
			if roomID != "" {
				assignID := fmt.Sprintf("RA%d", time.Now().UnixNano())
				db.DB.Exec(`INSERT INTO room_assignments (id, tenant_id, room_id, rent_amount, start_date, is_active)
					VALUES ($1, $2, $3, $4, $5::timestamp, true)`, assignID, id, roomID, rent, startDate)
			}

			http.Redirect(w, r, "/admin/tenants?msg=Tenant+added", http.StatusSeeOther)
			return
		} else 	if action == "update" {
		name := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		status := r.FormValue("status")
		// Handle empty email as NULL to avoid unique constraint violations
		var emailPtr interface{}
		if email == "" {
			emailPtr = nil
		} else {
			emailPtr = email
		}
		db.DB.Exec(`UPDATE tenants SET name=$1, phone=$2, email=$3, status=$4, updated_at=NOW() WHERE id=$5`,
			name, phone, emailPtr, status, id)
	} else if action == "assign" {
		roomID := r.FormValue("room_id")
		rent, _ := strconv.ParseFloat(r.FormValue("rent"), 64)
		startDate := r.FormValue("start_date")
		if roomID != "" {
			// First end any active assignments
			db.DB.Exec("UPDATE room_assignments SET is_active=false, updated_at=NOW() WHERE tenant_id=$1 AND is_active", id)
			assignID := fmt.Sprintf("RA%d", time.Now().UnixNano())
			db.DB.Exec(`INSERT INTO room_assignments (id, tenant_id, room_id, rent_amount, start_date, is_active)
				VALUES ($1, $2, $3, $4, $5, true)`, assignID, id, roomID, rent, startDate)
		}
	} else if action == "set_password" {
		newPass := r.FormValue("password")
		if newPass != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
			if err == nil {
				db.DB.Exec("UPDATE tenants SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hash), id)
			}
		}
	} else if action == "save_all" {
		name := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		status := r.FormValue("status")
		secDeposit, _ := strconv.ParseFloat(r.FormValue("security_deposit"), 64)
		lockIn, _ := strconv.Atoi(r.FormValue("lock_in_period"))
		newPass := r.FormValue("password")
		// Handle empty email as NULL to avoid unique constraint violations
		var emailPtr interface{}
		if email == "" {
			emailPtr = nil
		} else {
			emailPtr = email
		}
		db.DB.Exec(`UPDATE tenants SET name=$1, phone=$2, email=$3, status=$4,
			security_deposit=$5, security_lock_in_period=$6, updated_at=NOW() WHERE id=$7`,
			name, phone, emailPtr, status, secDeposit, lockIn, id)
		if newPass != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
			if err == nil {
				db.DB.Exec("UPDATE tenants SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hash), id)
			}
		}
		roomID := r.FormValue("room_id")
		rent, _ := strconv.ParseFloat(r.FormValue("rent"), 64)
		startDate := r.FormValue("start_date")
		if roomID != "" {
			db.DB.Exec("UPDATE room_assignments SET is_active=false, updated_at=NOW() WHERE tenant_id=$1 AND is_active", id)
			var existing int
			db.DB.QueryRow("SELECT COUNT(*) FROM room_assignments WHERE tenant_id=$1 AND room_id=$2 AND is_active", id, roomID).Scan(&existing)
			if existing == 0 {
				assignID := fmt.Sprintf("RA%d", time.Now().UnixNano())
				db.DB.Exec(`INSERT INTO room_assignments (id, tenant_id, room_id, rent_amount, start_date, is_active)
					VALUES ($1, $2, $3, $4, $5, true)`, assignID, id, roomID, rent, startDate)
			}
		}
		} else if action == "edit" {
			name := r.FormValue("name")
			phone := r.FormValue("phone")
			email := r.FormValue("email")
			status := r.FormValue("status")
			checkInDate := r.FormValue("check_in_date")
			secDeposit, _ := strconv.ParseFloat(r.FormValue("security_deposit"), 64)
			lockIn, _ := strconv.Atoi(r.FormValue("lock_in_period"))
			newPass := r.FormValue("password")

			// Handle empty email as NULL to avoid unique constraint violations
			var emailPtr interface{}
			if email == "" {
				emailPtr = nil
			} else {
				emailPtr = email
			}

			var updateErr error
			if checkInDate != "" {
				_, updateErr = db.DB.Exec(`UPDATE tenants SET name=$1, phone=$2, email=$3, status=$4,
					check_in_date=$5::timestamp, security_deposit=$6, security_lock_in_period=$7,
					updated_at=NOW() WHERE id=$8`,
					name, phone, emailPtr, status, checkInDate, secDeposit, lockIn, id)
			} else {
				_, updateErr = db.DB.Exec(`UPDATE tenants SET name=$1, phone=$2, email=$3, status=$4,
					security_deposit=$5, security_lock_in_period=$6, updated_at=NOW() WHERE id=$7`,
					name, phone, emailPtr, status, secDeposit, lockIn, id)
			}

			if updateErr != nil {
				log.Printf("ERROR updating tenant: %v", updateErr)
				http.Redirect(w, r, "/admin/tenants?error="+url.QueryEscape(friendlyDBError(updateErr)), http.StatusSeeOther)
				return
			}

			if newPass != "" {
				hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
				if err == nil {
					db.DB.Exec("UPDATE tenants SET password_hash=$1, updated_at=NOW() WHERE id=$2", string(hash), id)
				}
			}

			roomID := r.FormValue("room_id")
			rent, _ := strconv.ParseFloat(r.FormValue("rent"), 64)
			startDate := r.FormValue("start_date")
			if roomID != "" {
				db.DB.Exec("UPDATE room_assignments SET is_active=false, updated_at=NOW() WHERE tenant_id=$1 AND is_active", id)
				assignID := fmt.Sprintf("RA%d", time.Now().UnixNano())
				db.DB.Exec(`INSERT INTO room_assignments (id, tenant_id, room_id, rent_amount, start_date, is_active)
					VALUES ($1, $2, $3, $4, $5, true)`, assignID, id, roomID, rent, startDate)
			}

			http.Redirect(w, r, "/admin/tenants?msg=Tenant+updated", http.StatusSeeOther)
			return
		} else if action == "generate_pass" || action == "reset_pass" {
			// Get tenant phone for auto password generation
			var tenantPhone, tenantHash string
			db.DB.QueryRow(`SELECT t.phone, COALESCE(t.password_hash, '') FROM tenants t WHERE t.id = $1`, id).Scan(&tenantPhone, &tenantHash)

			autoPass := ""
			if tenantHash == "" {
				phone := strings.ReplaceAll(tenantPhone, " ", "")
				if len(phone) >= 4 {
					autoPass = "vats" + phone[len(phone)-4:]
				} else {
					autoPass = "vats" + phone
				}
			} else {
				// Reset: generate new random password
				phone := strings.ReplaceAll(tenantPhone, " ", "")
				if len(phone) >= 4 {
					autoPass = "vats" + phone[len(phone)-4:]
				} else {
					autoPass = "vats" + phone
				}
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(autoPass), bcrypt.DefaultCost)
			if err == nil {
				db.DB.Exec("UPDATE tenants SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hash), id)
			}

			http.Redirect(w, r, "/admin/tenants?msg="+url.QueryEscape("Login+created+|+Password:+ "+autoPass), http.StatusSeeOther)
			return
		} else if action == "delete" {
			db.DB.Exec("DELETE FROM room_assignments WHERE tenant_id = $1", id)
			db.DB.Exec("DELETE FROM tenant_verifications WHERE tenant_id = $1", id)
			db.DB.Exec("DELETE FROM tenant_passes WHERE tenant_id = $1", id)
			db.DB.Exec("DELETE FROM payments WHERE tenant_id = $1", id)
			_, err := db.DB.Exec("DELETE FROM tenants WHERE id = $1", id)
			if err != nil {
				log.Printf("ERROR deleting tenant %s: %v", id, err)
				http.Redirect(w, r, "/admin/tenants?error=Failed+to+delete+tenant", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/admin/tenants?msg=Tenant+deleted", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/admin/tenants", http.StatusSeeOther)
	}


// ─── Payments ──────────────────────────────────────────────────

func handleAdminPayments(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	rows, err := db.DB.Query(`SELECT p.id, p.tenant_id, COALESCE(t.name, 'N/A') as tenant_name,
		p.amount, p.payment_date, p.payment_method, p.status, p.month_covered, p.late_fee, p.notes
		FROM payments p
		LEFT JOIN tenants t ON p.tenant_id = t.id
		ORDER BY p.created_at DESC LIMIT 100`)
	if err != nil {
		log.Printf("ERROR loading payments: %v", err)
		renderAdminError(w, "Failed to load payments")
		return
	}
	defer rows.Close()

	type PaymentAdmin struct {
		ID, TenantID, TenantName, Date, Method, Status, MonthCovered, Notes string
		Amount, LateFee                                                      float64
	}
	var payments []PaymentAdmin
	for rows.Next() {
		var p PaymentAdmin
		if err := rows.Scan(&p.ID, &p.TenantID, &p.TenantName, &p.Amount, &p.Date, &p.Method, &p.Status, &p.MonthCovered, &p.LateFee, &p.Notes); err != nil {
			log.Printf("ERROR scanning payment row: %v", err)
			continue
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating payment rows: %v", err)
		renderAdminError(w, "Data error loading payments")
		return
	}
	if payments == nil {
		payments = []PaymentAdmin{}
	}

	// Tenant list for dropdown
	type TenantOpt struct{ ID, Name string }
	var tenantOpts []TenantOpt
	tenantRows, err := db.DB.Query("SELECT id, name FROM tenants WHERE status = 'active' ORDER BY name")
	if err != nil {
		log.Printf("ERROR loading tenant options: %v", err)
	} else {
		defer tenantRows.Close()
		for tenantRows.Next() {
			var t TenantOpt
			if err := tenantRows.Scan(&t.ID, &t.Name); err != nil {
				log.Printf("ERROR scanning tenant option: %v", err)
				continue
			}
			tenantOpts = append(tenantOpts, t)
		}
	}

	// Month options for dropdown (12 months back, 6 months forward from current month)
	type MonthOpt struct{ Value, Label string }
	var monthOpts []MonthOpt
	now := time.Now()
	for i := -12; i <= 6; i++ {
		t := now.AddDate(0, i, 0)
		monthOpts = append(monthOpts, MonthOpt{
			Value: t.Format("2006-01"),
			Label: t.Format("January 2006"),
		})
	}

	renderPrivate(w, "admin_payments.html", map[string]interface{}{
		"Payments": payments,
		"Tenants":  tenantOpts,
		"Months":   monthOpts,
		"Active":   "payments",
		"Title":    "Payments",
	})
}

func handleAdminPaymentsSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	action := r.FormValue("action")

	if action == "add" {
		tenantID := r.FormValue("tenant_id")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
		date := r.FormValue("date")
		method := r.FormValue("method")
		month := r.FormValue("month_covered")
		lateFee, _ := strconv.ParseFloat(r.FormValue("late_fee"), 64)
		notes := r.FormValue("notes")
		id := fmt.Sprintf("PAY%d", time.Now().UnixNano())
		db.DB.Exec(`INSERT INTO payments (id, tenant_id, amount, payment_date, payment_method, status, month_covered, late_fee, notes)
			VALUES ($1, $2, $3, $4, $5, 'completed', $6, $7, $8)`,
			id, tenantID, amount, date, method, month, lateFee, notes)
	} else if action == "update_status" {
		id := r.FormValue("id")
		status := r.FormValue("status")
		db.DB.Exec("UPDATE payments SET status=$1, updated_at=NOW() WHERE id=$2", status, id)
	} else if action == "delete" {
		id := r.FormValue("id")
		db.DB.Exec("DELETE FROM payments WHERE id=$1", id)
	}
	http.Redirect(w, r, "/admin/payments", http.StatusSeeOther)
}

// ─── Meters (Electricity) ──────────────────────────────────────

func handleAdminMeters(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	rows, err := db.DB.Query(`SELECT m.id, m.room_id, r.room_number, m.meter_type, m.meter_number,
		m.current_reading, m.initial_reading, m.is_active,
		COALESCE(b.status, '') as booking_status
		FROM meters m
		LEFT JOIN rooms r ON m.room_id = r.id
		LEFT JOIN bookings b ON m.room_id = b.room_id AND b.status = 'active'
		ORDER BY r.room_number, m.meter_type`)
	if err != nil {
		log.Printf("ERROR loading meters: %v", err)
		renderAdminError(w, "Failed to load meters")
		return
	}
	defer rows.Close()

	type MeterAdmin struct {
		ID, RoomID, RoomNumber, MeterType, MeterNumber, BookingStatus string
		CurrentReading, InitialReading                                int
		IsActive                                                      bool
	}
	var meters []MeterAdmin
	for rows.Next() {
		var m MeterAdmin
		var active bool
		if err := rows.Scan(&m.ID, &m.RoomID, &m.RoomNumber, &m.MeterType, &m.MeterNumber, &m.CurrentReading, &m.InitialReading, &active, &m.BookingStatus); err != nil {
			log.Printf("ERROR scanning meter row: %v", err)
			continue
		}
		m.IsActive = active
		meters = append(meters, m)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating meter rows: %v", err)
		renderAdminError(w, "Data error loading meters")
		return
	}
	if meters == nil {
		meters = []MeterAdmin{}
	}

	// Group meters by room
	type RoomMeterGroup struct {
		RoomID, RoomNumber string
		Occupied           bool
		Meters             []MeterAdmin
	}
	var roomGroups []RoomMeterGroup
	seen := map[string]int{}
	for _, m := range meters {
		idx, ok := seen[m.RoomNumber]
		if !ok {
			idx = len(roomGroups)
			roomGroups = append(roomGroups, RoomMeterGroup{
				RoomID:     m.RoomID,
				RoomNumber: m.RoomNumber,
				Occupied:   m.BookingStatus == "active",
				Meters:     []MeterAdmin{}})
			seen[m.RoomNumber] = idx
		}
		roomGroups[idx].Meters = append(roomGroups[idx].Meters, m)
	}

	var ratePerUnit float64 = 12
	var unitRate string
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'electricity_rate'").Scan(&unitRate)
	if unitRate != "" {
		if r, err := strconv.ParseFloat(unitRate, 64); err == nil {
			ratePerUnit = r
		}
	}

	renderPrivate(w, "admin_meters.html", map[string]interface{}{
		"RoomMeters":  roomGroups,
		"Meters":      meters,
		"RatePerUnit": ratePerUnit,
		"Active":      "meters",
		"Title":       "Electricity Meters",
	})
}

func handleAdminMetersSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	action := r.FormValue("action")

	if action == "add_meter" {
		roomID := r.FormValue("room_id")
		meterType := r.FormValue("meter_type")
		meterNumber := r.FormValue("meter_number")
		initial, _ := strconv.Atoi(r.FormValue("initial_reading"))
		id := fmt.Sprintf("MTR%d", time.Now().UnixNano())
		db.DB.Exec(`INSERT INTO meters (id, room_id, meter_type, meter_number, initial_reading, current_reading, is_active)
			VALUES ($1, $2, $3, $4, $5, $5, true)`, id, roomID, meterType, meterNumber, initial)
	} else if action == "add_reading" {
		meterID := r.FormValue("meter_id")
		reading, _ := strconv.Atoi(r.FormValue("reading"))
		readingDate := r.FormValue("reading_date")
		billingPeriod := r.FormValue("billing_period")
		id := fmt.Sprintf("MR%d", time.Now().UnixNano())
		db.DB.Exec(`INSERT INTO meter_readings (id, meter_id, reading, reading_date, billing_period)
			VALUES ($1, $2, $3, $4, $5)`, id, meterID, reading, readingDate, billingPeriod)
		db.DB.Exec("UPDATE meters SET current_reading = $1 WHERE id = $2", reading, meterID)
	} else if action == "update_rate" {
		rate := r.FormValue("rate")
		db.DB.Exec(`INSERT INTO settings (key, value, description) VALUES ('electricity_rate', $1, 'Electricity rate per unit')
			ON CONFLICT (key) DO UPDATE SET value=$1, updated_at=NOW()`, rate)
	}
	http.Redirect(w, r, "/admin/meters", http.StatusSeeOther)
}

func handleAdminMeterReadings(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	meterID := r.URL.Query().Get("meter_id")
	if meterID == "" {
		http.Error(w, "meter_id required", 400)
		return
	}
	rows, err := db.DB.Query(`SELECT id, meter_id, reading, reading_date, billing_period, created_at
		FROM meter_readings WHERE meter_id = $1 ORDER BY reading_date DESC LIMIT 50`, meterID)
	if err != nil {
		log.Printf("ERROR loading readings: %v", err)
		renderAdminError(w, "Failed to load readings")
		return
	}
	defer rows.Close()

	type Reading struct {
		ID, MeterID, Date, Period, CreatedAt string
		Value                                 int
	}
	var readings []Reading
	for rows.Next() {
		var rd Reading
		if err := rows.Scan(&rd.ID, &rd.MeterID, &rd.Value, &rd.Date, &rd.Period, &rd.CreatedAt); err != nil {
			log.Printf("ERROR scanning reading row: %v", err)
			continue
		}
		readings = append(readings, rd)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating reading rows: %v", err)
	}
	jsonContent(w)
	json.NewEncoder(w).Encode(map[string]interface{}{"readings": readings})
}

// ─── Passes ─────────────────────────────────────────────────────

func handleAdminPasses(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	rows, err := db.DB.Query(`
		SELECT t.id, t.name, t.phone, COALESCE(r.room_number,''),
			COALESCE(tp.id,'') as pass_id, COALESCE(tp.pass_number,'') as pass_number,
			COALESCE(tp.is_active,0) as pass_active
		FROM tenants t
		LEFT JOIN room_assignments ra ON t.id = ra.tenant_id AND ra.is_active
		LEFT JOIN rooms r ON ra.room_id = r.id
		LEFT JOIN tenant_passes tp ON t.id = tp.tenant_id AND tp.is_active = 1
		WHERE t.status = 'active'
		ORDER BY t.name`)
	if err != nil {
		log.Printf("ERROR loading passes: %v", err); renderAdminError(w, "Failed to load passes")
		return
	}
	defer rows.Close()

	type PassTenant struct {
		ID, Name, Phone, RoomNumber, PassID, PassNumber string
		PassActive                                       bool
	}
	var tenants []PassTenant
	for rows.Next() {
		var t PassTenant
		var pa int
		if err := rows.Scan(&t.ID, &t.Name, &t.Phone, &t.RoomNumber, &t.PassID, &t.PassNumber, &pa); err != nil {
			log.Printf("ERROR scanning pass row: %v", err)
			continue
		}
		t.PassActive = pa == 1
		tenants = append(tenants, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating pass rows: %v", err)
		renderAdminError(w, "Data error loading passes")
		return
	}
	if tenants == nil {
		tenants = []PassTenant{}
	}

	renderPrivate(w, "admin_passes.html", map[string]interface{}{
		"Tenants": tenants,
		"Active":  "passes",
		"Title":   "Digital Passes",
	})
}

func handleAdminPassGenerate(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	tenantID := r.FormValue("tenant_id")
	if tenantID == "" {
		http.Redirect(w, r, "/admin/passes", http.StatusSeeOther)
		return
	}

	var existingID, existingPass string
	db.DB.QueryRow("SELECT id, pass_number FROM tenant_passes WHERE tenant_id = $1 AND is_active = 1", tenantID).Scan(&existingID, &existingPass)
	if existingPass != "" {
		http.Redirect(w, r, "/admin/passes?msg=Pass+already+exists:+ "+existingPass, http.StatusSeeOther)
		return
	}

	var roomNum, tenantName string
	db.DB.QueryRow(`
		SELECT t.name, COALESCE(r.room_number,'')
		FROM tenants t
		LEFT JOIN room_assignments ra ON t.id = ra.tenant_id AND ra.is_active
		LEFT JOIN rooms r ON ra.room_id = r.id
		WHERE t.id = $1`, tenantID).Scan(&tenantName, &roomNum)

	passNum := fmt.Sprintf("VATS%s%s", roomNum, time.Now().Format("20060102"))
	id := fmt.Sprintf("PASS%d", time.Now().UnixNano())
	db.DB.Exec(`INSERT INTO tenant_passes (id, tenant_id, pass_number, issued_by, issued_at, is_active)
		VALUES ($1, $2, $3, 'admin', NOW(), 1)`, id, tenantID, passNum)

	http.Redirect(w, r, "/admin/passes?msg=Pass+"+passNum+"+generated+for+"+tenantName, http.StatusSeeOther)
}

func handleAdminPassRevoke(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	passID := r.FormValue("pass_id")
	if passID != "" {
		db.DB.Exec("UPDATE tenant_passes SET is_active = 0, updated_at = NOW() WHERE id = $1", passID)
	}
	http.Redirect(w, r, "/admin/passes", http.StatusSeeOther)
}

// ─── Verifications ─────────────────────────────────────────────

func handleAdminVerifications(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	rows, err := db.DB.Query(`
		SELECT tv.id, tv.tenant_id, t.name, t.phone, COALESCE(r.room_number,''),
			COALESCE(tv.lpu_id_photo,''), COALESCE(tv.aadhar_photo,''),
			COALESCE(tv.status,'not_submitted'),
			COALESCE(TO_CHAR(tv.submitted_at, 'Mon DD, YYYY HH24:MI'), ''),
			COALESCE(tv.notes,'')
		FROM tenant_verifications tv
		JOIN tenants t ON tv.tenant_id = t.id
		LEFT JOIN room_assignments ra ON t.id = ra.tenant_id AND ra.is_active
		LEFT JOIN rooms r ON ra.room_id = r.id
		ORDER BY tv.submitted_at DESC NULLS LAST, tv.created_at DESC`)
	if err != nil {
		log.Printf("ERROR loading verifications: %v", err)
		renderAdminError(w, "Failed to load verifications")
		return
	}
	defer rows.Close()

	type VerificationAdmin struct {
		ID, TenantID, Name, Phone, Room, LpuPhoto, AadharPhoto, Status, SubmittedAt, Notes string
	}
	var verifications []VerificationAdmin
	for rows.Next() {
		var v VerificationAdmin
		if err := rows.Scan(&v.ID, &v.TenantID, &v.Name, &v.Phone, &v.Room,
			&v.LpuPhoto, &v.AadharPhoto, &v.Status, &v.SubmittedAt, &v.Notes); err != nil {
			log.Printf("ERROR scanning verification row: %v", err)
			continue
		}
		verifications = append(verifications, v)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating verification rows: %v", err)
		renderAdminError(w, "Data error loading verifications")
		return
	}
	if verifications == nil {
		verifications = []VerificationAdmin{}
	}

	renderPrivate(w, "admin_verifications.html", map[string]interface{}{
		"Verifications": verifications,
		"Active":        "verifications",
		"Title":         "Verification Requests",
	})
}

func handleAdminVerificationAction(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	verID := r.FormValue("id")
	action := r.FormValue("action")
	notes := r.FormValue("notes")

	if action == "verify" {
		db.DB.Exec(`UPDATE tenant_verifications SET status = 'verified', verified_at = NOW(), notes = $1, updated_at = NOW() WHERE id = $2`, notes, verID)
	} else if action == "reject" {
		db.DB.Exec(`UPDATE tenant_verifications SET status = 'rejected', notes = $1, updated_at = NOW() WHERE id = $2`, notes, verID)
	}
	http.Redirect(w, r, "/admin/verifications", http.StatusSeeOther)
}

// ─── Helpers ───────────────────────────────────────────────────

// friendlyDBError converts PostgreSQL unique violations into user-friendly messages
func friendlyDBError(err error) string {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		constraint := pqErr.Constraint
		switch constraint {
		case "tenants_phone_unique", "tenants_phone_key":
			return "A tenant with this phone number already exists"
		case "tenants_email_unique", "tenants_email_key":
			return "A tenant with this email address already exists"
		default:
			return "This record already exists (duplicate: " + constraint + ")"
		}
	}
	return err.Error()
}

func renderAdminError(w http.ResponseWriter, message string) {
	renderPrivate(w, "admin_error.html", map[string]interface{}{
		"ErrorMessage": message,
		"Active":       "dashboard",
		"Title":        "Error",
	})
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("₹%.0f", amount)
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
