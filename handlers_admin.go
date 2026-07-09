package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	db.DB.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM rooms),
			(SELECT COUNT(DISTINCT room_id) FROM tenants WHERE room_id IS NOT NULL AND (end_date IS NULL OR end_date = '')),
			(SELECT COUNT(*) FROM tenants),
			(SELECT COUNT(*) FROM tenants WHERE status = 'active')
	`).Scan(&s.TotalRooms, &s.OccupiedRooms, &s.TotalTenants, &s.ActiveTenants)

	currentMonth := time.Now().Format("2006-01")
	var total, completed float64
	db.DB.QueryRow(`
		SELECT
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'paid'), 0),
			COALESCE(SUM(total_amount), 0),
			COUNT(*) FILTER (WHERE status = 'pending')
		FROM bills WHERE billing_month = $1
	`, currentMonth).Scan(&completed, &total, &s.PendingPayments)

	s.MonthlyRevenue = completed
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
		CASE WHEN EXISTS (
			SELECT 1 FROM tenants t2 WHERE t2.room_id = r.id AND (t2.end_date IS NULL OR t2.end_date = '')
		) THEN 'occupied' ELSE 'vacant' END as status
		FROM rooms r WHERE r.id != 'BUILDING'
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
	rows, err := db.DB.Query(`
		SELECT t.id, t.name, COALESCE(t.email, '') as email, t.phone, t.status,
		COALESCE(t.check_in_date::text, '') as check_in_date,
		COALESCE(t.security_deposit, 0) as security_deposit, COALESCE(t.security_lock_in_period, 0) as security_lock_in_period,
		COALESCE(t.has_maintenance, false) as has_maintenance,
		COALESCE(t.room_id, '') as room_id, COALESCE(r.room_number, '') as room_number,
		COALESCE(t.rent_amount, 0) as rent_amount, COALESCE(t.check_in_date::text, '') as start_date,
		COALESCE(r.maintenance_amount, 500) as maintenance_amount, COALESCE(r.price, 0) as room_price,
		COALESCE(t.password_hash, '') as password_hash,
		COALESCE(t.pass_number, '') as pass_number,
		COALESCE(t.verification_status, 'not_submitted') as ver_status,
		COALESCE(t.lpu_id_photo, '') as lpu_photo, COALESCE(t.aadhar_photo, '') as aadhar_photo,
		COALESCE(TO_CHAR(t.verification_submitted_at, 'Mon DD, YYYY HH24:MI'), '') as ver_submitted_at,
		COALESCE(t.verification_notes, '') as ver_notes
		FROM tenants t
		LEFT JOIN rooms r ON t.room_id = r.id
		ORDER BY t.name`)
	if err != nil {
		log.Printf("ERROR loading tenants: %v", err); renderAdminError(w, "Failed to load tenants")
		return
	}
	defer rows.Close()

	type TenantAdmin struct {
		ID, Name, Email, Phone, Status, CheckInDate, RoomID, RoomNumber           string
		SecurityDeposit                                                            float64
		LockInPeriod                                                               int
		RentAmount                                                                 float64
		StartDate                                                                  string
		HasPassword                                                                bool
		HasMaintenance                                                             bool
		MaintenanceAmount                                                          float64
		RoomPrice                                                                  float64
		PassNumber                                                                 string
		VerificationStatus, LpuPhoto, AadharPhoto, VerificationSubmittedAt, VerificationNotes string
	}
	var tenants []TenantAdmin
	for rows.Next() {
		var t TenantAdmin
		var pwHash, passNum, lpuPhoto, aadharPhoto, verSubmittedAt, verNotes, verStatus string
		if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Phone, &t.Status, &t.CheckInDate,
			&t.SecurityDeposit, &t.LockInPeriod, &t.HasMaintenance, &t.RoomID, &t.RoomNumber, &t.RentAmount, &t.StartDate, &t.MaintenanceAmount, &t.RoomPrice,
			&pwHash,
			&passNum,
			&verStatus, &lpuPhoto, &aadharPhoto, &verSubmittedAt, &verNotes); err != nil {
			log.Printf("ERROR scanning tenant row: %v", err)
			continue
		}
		t.HasPassword = pwHash != ""
		t.PassNumber = passNum
		t.VerificationStatus = verStatus
		t.LpuPhoto = lpuPhoto
		t.AadharPhoto = aadharPhoto
		t.VerificationSubmittedAt = verSubmittedAt
		t.VerificationNotes = verNotes
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
		_ = id
			phone := r.FormValue("phone")
			email := r.FormValue("email")
			status := r.FormValue("status")
			if status == "" {
				status = "active"
			}
			secDeposit, _ := strconv.ParseFloat(r.FormValue("security_deposit"), 64)
			lockIn, _ := strconv.Atoi(r.FormValue("lock_in_period"))
			hasMaint := r.FormValue("has_maintenance") == "on"
			newPass := r.FormValue("password")

			var emailPtr interface{}
			if email == "" {
				emailPtr = nil
			} else {
				emailPtr = email
			}

			id := fmt.Sprintf("T%03d", time.Now().UnixNano()%100000)
			roomID := r.FormValue("room_id")
			rent, _ := strconv.ParseFloat(r.FormValue("rent"), 64)
			startDate := r.FormValue("start_date")
			maintAmt, _ := strconv.ParseFloat(r.FormValue("maintenance_amount"), 64)

			_, err := db.DB.Exec(`INSERT INTO tenants (id, name, phone, email, status, security_deposit, security_lock_in_period, has_maintenance, room_id, rent_amount, check_in_date)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
				id, name, phone, emailPtr, status, secDeposit, lockIn, hasMaint, roomID, rent, startDate)
			if err != nil {
				log.Printf("ERROR adding tenant: %v", err)
				renderAdminError(w, "Failed to add tenant: "+friendlyDBError(err))
			}

		if newPass != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
			if err == nil {
				db.DB.Exec("UPDATE tenants SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hash), id)
			}
		}

		if roomID != "" {
			db.DB.Exec("UPDATE rooms SET maintenance_amount = $1, updated_at = NOW() WHERE id = $2", maintAmt, roomID)
		}

		http.Redirect(w, r, "/admin/tenants?msg=Tenant+added", http.StatusSeeOther)
		return
	} else 	if action == "update" {
		name := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		status := r.FormValue("status")
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
			db.DB.Exec("UPDATE tenants SET room_id=$1, rent_amount=$2, check_in_date=$3, end_date=NULL, updated_at=NOW() WHERE id=$4",
				roomID, rent, startDate, id)
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
		hasMaint := r.FormValue("has_maintenance") == "on"
		newPass := r.FormValue("password")
		var emailPtr interface{}
		if email == "" {
			emailPtr = nil
		} else {
			emailPtr = email
		}
		db.DB.Exec(`UPDATE tenants SET name=$1, phone=$2, email=$3, status=$4,
			security_deposit=$5, security_lock_in_period=$6, has_maintenance=$7,
			updated_at=NOW() WHERE id=$8`,
			name, phone, emailPtr, status, secDeposit, lockIn, hasMaint, id)
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
			db.DB.Exec("UPDATE tenants SET room_id=$1, rent_amount=$2, check_in_date=$3, end_date=NULL, updated_at=NOW() WHERE id=$4",
				roomID, rent, startDate, id)
		}
	} else if action == "edit" {
		name := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		status := r.FormValue("status")
		secDeposit, _ := strconv.ParseFloat(r.FormValue("security_deposit"), 64)
		lockIn, _ := strconv.Atoi(r.FormValue("lock_in_period"))
		hasMaint := r.FormValue("has_maintenance") == "on"
		maintAmt, _ := strconv.ParseFloat(r.FormValue("maintenance_amount"), 64)
		newPass := r.FormValue("password")

		var emailPtr interface{}
		if email == "" {
			emailPtr = nil
		} else {
			emailPtr = email
		}

		_, err := db.DB.Exec(`UPDATE tenants SET name=$1, phone=$2, email=$3, status=$4,
			security_deposit=$5, security_lock_in_period=$6, has_maintenance=$7,
			updated_at=NOW() WHERE id=$8`,
			name, phone, emailPtr, status, secDeposit, lockIn, hasMaint, id)
		if err != nil {
			log.Printf("ERROR updating tenant: %v", err)
			http.Redirect(w, r, "/admin/tenants?error="+url.QueryEscape(friendlyDBError(err)), http.StatusSeeOther)
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
			db.DB.Exec("UPDATE tenants SET room_id=$1, rent_amount=$2, check_in_date=$3, end_date=NULL, updated_at=NOW() WHERE id=$4",
				roomID, rent, startDate, id)
			db.DB.Exec("UPDATE rooms SET maintenance_amount = $1, updated_at = NOW() WHERE id = $2", maintAmt, roomID)
		}

		http.Redirect(w, r, "/admin/tenants?msg=Tenant+updated", http.StatusSeeOther)
		return
	} else if action == "generate_pass" || action == "reset_pass" {
		var roomNum string
		db.DB.QueryRow(`SELECT COALESCE(r.room_number,'')
			FROM tenants t LEFT JOIN rooms r ON t.room_id = r.id
			WHERE t.id = $1`, id).Scan(&roomNum)
		passNum := fmt.Sprintf("VATS%s%s", roomNum, time.Now().Format("20060102"))
		passHash, _ := bcrypt.GenerateFromPassword([]byte(passNum), bcrypt.DefaultCost)
		db.DB.Exec("UPDATE tenants SET pass_number=$1, pass_issued_at=NOW(), password_hash=$2, updated_at=NOW() WHERE id=$3", passNum, string(passHash), id)
		http.Redirect(w, r, "/admin/tenants?msg="+url.QueryEscape("Pass+generated:+ "+passNum+"+(also+login+password)"), http.StatusSeeOther)
		return
	} else if action == "delete" {
		db.DB.Exec("DELETE FROM payments WHERE tenant_id = $1", id)
		_, err := db.DB.Exec("DELETE FROM tenants WHERE id = $1", id)
		if err != nil {
			log.Printf("ERROR deleting tenant %s: %v", id, err)
			http.Redirect(w, r, "/admin/tenants?error=Failed+to+delete+tenant", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/admin/tenants?msg=Tenant+deleted", http.StatusSeeOther)
		return
	} else if action == "revoke_pass" {
		db.DB.Exec("UPDATE tenants SET pass_number=NULL, pass_issued_at=NULL, password_hash=NULL WHERE id=$1", id)
		http.Redirect(w, r, "/admin/tenants?msg=Pass+revoked+(login+disabled)", http.StatusSeeOther)
		return
	} else if action == "verify_tenant" {
		notes := r.FormValue("notes")
		db.DB.Exec("UPDATE tenants SET verification_status='verified', verification_verified_at=NOW(), verification_notes=$1 WHERE id=$2", notes, id)
		http.Redirect(w, r, "/admin/tenants?msg=Tenant+verified", http.StatusSeeOther)
		return
	} else if action == "reject_tenant" {
		notes := r.FormValue("notes")
		db.DB.Exec("UPDATE tenants SET verification_status='rejected', verification_notes=$1 WHERE id=$2", notes, id)
		http.Redirect(w, r, "/admin/tenants?msg=Verification+rejected", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin/tenants", http.StatusSeeOther)
}

// ─── Payments ──────────────────────────────────────────────────

type RoomPaymentCard struct {
	RoomID, RoomNumber, TenantID, TenantName              string
	RentAmount, RoomPrice, DiscountAmount                  float64
	HasMaintenance                                         bool
	MaintenanceAmt                                         float64
	ElectricityAmt                                         float64
	WaterAmt                                               float64
	MonthlyDiscount                                        float64
	MonthlyDiscountNote                                    string
	TotalDue                                               float64
	HasPaid                                                bool
	BillID, BillStatus                                     string
	PaymentID, PaymentDate, PaymentMethod, PaymentNotes, PaidTo string
	PaymentAmount                                          float64
}

func handleAdminPayments(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	// Determine billing month (default: current month)
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	prevMonth := monthOffset(month, -1)
	nextMonth := monthOffset(month, 1)
	currentMonth := time.Now().Format("2006-01")
	if nextMonth > currentMonth {
		nextMonth = ""
	}

	rows, err := db.DB.Query(`
		SELECT
			r.id, r.room_number,
			COALESCE(t.id, '') as tenant_id,
			COALESCE(t.name, '') as tenant_name,
			COALESCE(t.rent_amount, 0) as rent_amount,
			COALESCE(r.price, 0) as room_price,
			COALESCE(t.has_maintenance, false) as has_maintenance,
			COALESCE(r.maintenance_amount, 500) as maintenance_amount,
			COALESCE(b.id, '') as bill_id,
			COALESCE(b.electricity_amount, 0) as bill_elec,
			COALESCE(b.water_amount, 0) as bill_water,
			COALESCE(b.discount_amount, 0) as bill_discount,
			COALESCE(b.discount_note, '') as bill_discount_note,
			COALESCE(b.total_amount, 0) as bill_total,
			COALESCE(b.status, '') as bill_status,
			COALESCE(p.id, '') as payment_id,
			COALESCE(p.amount, 0) as payment_amount,
			COALESCE(p.payment_date::text, '') as payment_date,
			COALESCE(p.payment_method, '') as payment_method,
			COALESCE(p.notes, '') as payment_notes,
			COALESCE(p.paid_to, '') as paid_to
		FROM rooms r
		JOIN tenants t ON t.room_id = r.id AND t.status = 'active' AND (t.end_date IS NULL OR t.end_date = '')
		LEFT JOIN bills b ON b.tenant_id = t.id AND b.billing_month = $1
		LEFT JOIN payments p ON p.bill_id = b.id AND p.status = 'completed'
		WHERE r.id != 'BUILDING' ORDER BY r.room_number`, month)
	if err != nil {
		log.Printf("ERROR loading payment cards: %v", err)
		renderAdminError(w, "Failed to load payments")
		return
	}
	defer rows.Close()

	var cards []RoomPaymentCard
	for rows.Next() {
		var c RoomPaymentCard
		var billID, billStatus, billDiscountNote string
		var billElec, billWater, billDiscount, billTotal float64
		if err := rows.Scan(&c.RoomID, &c.RoomNumber, &c.TenantID, &c.TenantName,
			&c.RentAmount, &c.RoomPrice, &c.HasMaintenance, &c.MaintenanceAmt,
			&billID, &billElec, &billWater, &billDiscount, &billDiscountNote, &billTotal, &billStatus,
			&c.PaymentID, &c.PaymentAmount, &c.PaymentDate, &c.PaymentMethod, &c.PaymentNotes, &c.PaidTo); err != nil {
			log.Printf("ERROR scanning payment card: %v", err)
			continue
		}
		c.BillID = billID
		c.BillStatus = billStatus
		c.DiscountAmount = c.RoomPrice - c.RentAmount
		if c.DiscountAmount < 0 {
			c.DiscountAmount = 0
		}
		c.MonthlyDiscount = billDiscount
		c.MonthlyDiscountNote = billDiscountNote
		c.HasPaid = c.PaymentID != ""

		if billID != "" {
			c.ElectricityAmt = billElec + billWater
			c.WaterAmt = 0
			c.TotalDue = billTotal - billDiscount
		} else {
			c.TotalDue = c.RentAmount
			if c.HasMaintenance {
				c.TotalDue += c.MaintenanceAmt
			}
		}
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR iterating payment cards: %v", err)
		renderAdminError(w, "Data error loading payments")
		return
	}
	if cards == nil {
		cards = []RoomPaymentCard{}
	}

	billsGenerated := false
	for _, c := range cards {
		if c.BillID != "" {
			billsGenerated = true
			break
		}
	}

	// If no bills yet, compute utility amounts from meters for display
	if !billsGenerated {
		var ratePerUnit float64 = 12
		var rateStr string
		db.DB.QueryRow("SELECT value FROM settings WHERE key = 'electricity_rate'").Scan(&rateStr)
		if rateStr != "" {
			if r, err := strconv.ParseFloat(rateStr, 64); err == nil {
				ratePerUnit = r
			}
		}

		// Per-room electricity from meters
		utilityMap := map[string]float64{}
		eRows, err := db.DB.Query(`
			SELECT m.room_id, COALESCE(SUM(mr.current_reading - mr.initial_reading), 0) as units
			FROM meters m
			JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
			WHERE m.is_active = true AND m.room_id != 'BUILDING'
			GROUP BY m.room_id`, month)
		if err == nil {
			defer eRows.Close()
			for eRows.Next() {
				var roomID string
				var units float64
				eRows.Scan(&roomID, &units)
				if units > 0 {
					utilityMap[roomID] = units * ratePerUnit
				}
			}
		}

		// Water share
		var waterPerRoom float64
		var occCount int
		db.DB.QueryRow("SELECT COUNT(DISTINCT room_id) FROM tenants WHERE room_id IS NOT NULL AND (end_date IS NULL OR end_date = '')").Scan(&occCount)
		if occCount > 0 {
			var waterUnits float64
			db.DB.QueryRow(`
				SELECT COALESCE(SUM(mr.current_reading - mr.initial_reading), 0)
				FROM meters m
				JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
				WHERE m.room_id = 'BUILDING' AND m.meter_type = 'Water' AND m.is_active = true`, month).Scan(&waterUnits)
			var savedPerRoom string
			db.DB.QueryRow("SELECT value FROM settings WHERE key = 'water_per_room_' || $1", month).Scan(&savedPerRoom)
			if savedPerRoom != "" {
				if v, _ := strconv.ParseFloat(savedPerRoom, 64); v > 0 {
					waterPerRoom = v
				}
			} else if waterUnits > 0 {
				waterPerRoom = (waterUnits / float64(occCount)) * ratePerUnit
			}
		}

		// Apply to cards
		for i := range cards {
			var utilityTotal float64
			if amt, ok := utilityMap[cards[i].RoomID]; ok {
				cards[i].ElectricityAmt = amt
				utilityTotal += amt
			}
			if waterPerRoom > 0 {
				cards[i].WaterAmt = waterPerRoom
				utilityTotal += waterPerRoom
			}
			if utilityTotal > 0 {
				cards[i].TotalDue += utilityTotal
			}
		}
	}

	renderPrivate(w, "admin_payments.html", map[string]interface{}{
		"Cards":           cards,
		"Month":           month,
		"MonthLabel":      formatMonthLabel(month),
		"PrevMonth":       prevMonth,
		"NextMonth":       nextMonth,
		"BillsGenerated":  billsGenerated,
		"Active":          "payments",
		"Title":           "Payments",
	})
}

func handleAdminPaymentsSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	action := r.FormValue("action")
	month := r.FormValue("month")

	if action == "add" {
		tenantID := r.FormValue("tenant_id")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
		date := r.FormValue("date")
		method := r.FormValue("method")
		monthCovered := r.FormValue("month_covered")
		lateFee, _ := strconv.ParseFloat(r.FormValue("late_fee"), 64)
		notes := r.FormValue("notes")
		id := fmt.Sprintf("PAY%d", time.Now().UnixNano())
		db.DB.Exec(`INSERT INTO payments (id, tenant_id, amount, payment_date, payment_method, status, month_covered, late_fee, notes)
			VALUES ($1, $2, $3, $4, $5, 'completed', $6, $7, $8)`,
			id, tenantID, amount, date, method, monthCovered, lateFee, notes)
		if monthCovered != "" {
			http.Redirect(w, r, "/admin/payments?month="+monthCovered, http.StatusSeeOther)
			return
		}
	} else if action == "undo" {
		paymentID := r.FormValue("payment_id")
		var billID string
		db.DB.QueryRow("SELECT bill_id FROM payments WHERE id = $1", paymentID).Scan(&billID)
		db.DB.Exec("DELETE FROM payments WHERE id = $1", paymentID)
		if billID != "" {
			db.DB.Exec("UPDATE bills SET status = 'pending' WHERE id = $1", billID)
		}
		http.Redirect(w, r, "/admin/payments?month="+month, http.StatusSeeOther)
		return
	} else if action == "complete" {
		tenantID := r.FormValue("tenant_id")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
		discount, _ := strconv.ParseFloat(r.FormValue("discount_amount"), 64)
		discountNote := r.FormValue("discount_note")
		paidAmount := amount - discount
		if paidAmount < 0 { paidAmount = 0 }
		date := r.FormValue("date")
		method := r.FormValue("method")
		monthCovered := r.FormValue("month_covered")
		notes := r.FormValue("notes")
		paidTo := r.FormValue("paid_to")
		maintAmt, _ := strconv.ParseFloat(r.FormValue("maintenance_amount"), 64)
		hasMaint := maintAmt > 0
		id := fmt.Sprintf("PAY%d", time.Now().UnixNano())

		var billID string
		db.DB.QueryRow(`SELECT id FROM bills WHERE tenant_id = $1 AND billing_month = $2`,
			tenantID, monthCovered).Scan(&billID)
		if billID == "" {
			billID = fmt.Sprintf("BILL%d", time.Now().UnixNano())
			var roomID string
			db.DB.QueryRow(`SELECT COALESCE(room_id,'') FROM tenants WHERE id = $1`, tenantID).Scan(&roomID)
			db.DB.Exec(`INSERT INTO bills (id, tenant_id, room_id, billing_month, rent_amount, total_amount, discount_amount, discount_note, status)
				VALUES ($1, $2, $3, $4, $5, $5, 'paid')`,
				billID, tenantID, roomID, monthCovered, amount, discount, discountNote)
		} else {
			db.DB.Exec("UPDATE bills SET discount_amount = $1, discount_note = $2, status = 'paid' WHERE id = $3", discount, discountNote, billID)
		}

		db.DB.Exec(`INSERT INTO payments (id, tenant_id, bill_id, amount, payment_date, payment_method, status, month_covered, notes, paid_to)
			VALUES ($1, $2, $3, $4, $5, $6, 'completed', $7, $8, $9)`,
			id, tenantID, billID, paidAmount, date, method, monthCovered, notes, paidTo)
		db.DB.Exec("UPDATE tenants SET has_maintenance = $1, updated_at = NOW() WHERE id = $2", hasMaint, tenantID)
		db.DB.Exec(`UPDATE rooms SET maintenance_amount = $1, updated_at = NOW()
			WHERE id = (SELECT COALESCE(room_id,'') FROM tenants WHERE id = $2)`, maintAmt, tenantID)
		http.Redirect(w, r, "/admin/payments?month="+monthCovered, http.StatusSeeOther)
		return
	} else if action == "toggle_maintenance" {
		tenantID := r.FormValue("tenant_id")
		currentVal := r.FormValue("has_maintenance") == "true"
		db.DB.Exec("UPDATE tenants SET has_maintenance = $1, updated_at = NOW() WHERE id = $2", !currentVal, tenantID)
		if month != "" {
			http.Redirect(w, r, "/admin/payments?month="+month, http.StatusSeeOther)
			return
		}
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

func handleAdminMeters(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	// Determine billing month (default: current month)
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	prevMonth := monthOffset(month, -1)
	nextMonth := monthOffset(month, 1)
	// Don't allow future months beyond current
	currentMonth := time.Now().Format("2006-01")
	if nextMonth > currentMonth {
		nextMonth = ""
	}

	// Ensure monthly readings exist for this month (auto-create from previous month)
	autoCreateMonthlyReadings(month)

	// Load per-room meters with monthly readings
	rows, err := db.DB.Query(`SELECT m.id, m.room_id, r.room_number, m.meter_type, m.meter_number,
		COALESCE(mr.initial_reading, 0), COALESCE(mr.current_reading, 0), m.is_active,
		CASE WHEN EXISTS (SELECT 1 FROM tenants t3 WHERE t3.room_id = m.room_id AND (t3.end_date IS NULL OR t3.end_date = '')) THEN 'active' ELSE '' END as booking_status,
		COALESCE(mr.id, '') as reading_id
		FROM meters m
		LEFT JOIN rooms r ON m.room_id = r.id
		LEFT JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
		WHERE m.room_id != 'BUILDING'
		ORDER BY r.room_number, m.meter_type`, month)
	if err != nil {
		log.Printf("ERROR loading meters: %v", err)
		renderAdminError(w, "Failed to load meters")
		return
	}
	defer rows.Close()

	type MeterAdmin struct {
		ID, RoomID, RoomNumber, MeterType, MeterNumber, BookingStatus, ReadingID string
		CurrentReading, InitialReading                                           int
		IsActive                                                                 bool
	}
	var meters []MeterAdmin
	for rows.Next() {
		var m MeterAdmin
		var active bool
		if err := rows.Scan(&m.ID, &m.RoomID, &m.RoomNumber, &m.MeterType, &m.MeterNumber,
			&m.InitialReading, &m.CurrentReading, &active, &m.BookingStatus, &m.ReadingID); err != nil {
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
		RoomTotal          float64
		GrandTotal         float64
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

	// Rate settings
	var ratePerUnit float64 = 12
	var unitRate string
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'electricity_rate'").Scan(&unitRate)
	if unitRate != "" {
		if r, err := strconv.ParseFloat(unitRate, 64); err == nil {
			ratePerUnit = r
		}
	}

	// Compute room totals
	for i := range roomGroups {
		var total float64
		for _, m := range roomGroups[i].Meters {
			units := m.CurrentReading - m.InitialReading
			if units > 0 {
				total += float64(units) * ratePerUnit
			}
		}
		roomGroups[i].RoomTotal = total
	}

	// Water meter with monthly reading
	var waterMeter MeterAdmin
	var waterActive bool
	db.DB.QueryRow(`SELECT m.id, m.room_id, 'Building', m.meter_type, m.meter_number,
		COALESCE(mr.initial_reading, 0), COALESCE(mr.current_reading, 0), m.is_active,
		COALESCE(mr.id, '') as reading_id
		FROM meters m
		LEFT JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
		WHERE m.room_id = 'BUILDING' AND m.meter_type = 'Water' AND m.is_active = true`, month).Scan(
		&waterMeter.ID, &waterMeter.RoomID, &waterMeter.RoomNumber, &waterMeter.MeterType, &waterMeter.MeterNumber,
		&waterMeter.InitialReading, &waterMeter.CurrentReading, &waterActive, &waterMeter.ReadingID)
	waterMeter.IsActive = waterActive
	hasWaterMeter := waterMeter.ID != ""

	// Count occupied rooms
	var occupiedCount int
	db.DB.QueryRow("SELECT COUNT(DISTINCT room_id) FROM tenants WHERE room_id IS NOT NULL AND (end_date IS NULL OR end_date = '')").Scan(&occupiedCount)

	// Water calculations
	waterUnits := waterMeter.CurrentReading - waterMeter.InitialReading
	var waterUnitsPerRoom, waterPerRoom float64
	if occupiedCount > 0 && waterUnits > 0 {
		waterUnitsPerRoom = float64(waterUnits) / float64(occupiedCount)
		waterPerRoom = waterUnitsPerRoom * ratePerUnit
	}
	// Load any manual water overrides for this month from settings
	waterBill := waterPerRoom * float64(occupiedCount)
	waterPerRoomOverride := waterPerRoom
	var savedUnits, savedAmount, savedPerRoom string
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'water_units_' || $1", month).Scan(&savedUnits)
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'water_amount_' || $1", month).Scan(&savedAmount)
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'water_per_room_' || $1", month).Scan(&savedPerRoom)
	if savedUnits != "" {
		if v, err := strconv.Atoi(savedUnits); err == nil {
			waterUnits = v
			waterUnitsPerRoom = float64(v) / float64(max(occupiedCount, 1))
		}
	}
	if savedAmount != "" {
		if v, err := strconv.ParseFloat(savedAmount, 64); err == nil {
			waterBill = v
		}
	}
	if savedPerRoom != "" {
		if v, err := strconv.ParseFloat(savedPerRoom, 64); err == nil {
			waterPerRoomOverride = v
			waterPerRoom = v
		}
	} else if waterBill > 0 && occupiedCount > 0 {
		waterPerRoomOverride = waterBill / float64(occupiedCount)
	}

	// Grand totals (room total + water share)
	for i := range roomGroups {
		roomGroups[i].GrandTotal = roomGroups[i].RoomTotal + waterPerRoom
	}

	renderPrivate(w, "admin_meters.html", map[string]interface{}{
		"RoomMeters":        roomGroups,
		"Meters":            meters,
		"RatePerUnit":       ratePerUnit,
		"WaterMeter":        waterMeter,
		"HasWaterMeter":     hasWaterMeter,
		"WaterUnits":            waterUnits,
		"WaterBill":             waterBill,
		"WaterPerRoomOverride":  waterPerRoomOverride,
		"OccupiedCount":         occupiedCount,
		"WaterPerRoom":          waterPerRoom,
		"WaterUnitsPerRoom":     waterUnitsPerRoom,
		"Month":             month,
		"MonthLabel":        formatMonthLabel(month),
		"PrevMonth":         prevMonth,
		"NextMonth":         nextMonth,
		"Active":            "meters",
		"Title":             "Meters & Utilities",
	})
}

// autoCreateMonthlyReadings creates monthly_readings rows for the given month
// by carrying forward current readings from the previous month (or meter defaults if first month).
func autoCreateMonthlyReadings(month string) {
	prevMonth := monthOffset(month, -1)

	// For each active meter where no reading exists for this month, create one
	_, err := db.DB.Exec(`
		INSERT INTO monthly_readings (id, meter_id, billing_month, initial_reading, current_reading)
		SELECT 'MR-' || m.id || '-' || $1,
			m.id, $1,
			COALESCE(
				(SELECT mr2.current_reading FROM monthly_readings mr2
				 WHERE mr2.meter_id = m.id AND mr2.billing_month = $2),
				m.initial_reading
			),
			COALESCE(
				(SELECT mr2.current_reading FROM monthly_readings mr2
				 WHERE mr2.meter_id = m.id AND mr2.billing_month = $2),
				m.initial_reading
			)
		FROM meters m
		WHERE m.is_active = true
		AND NOT EXISTS (
			SELECT 1 FROM monthly_readings mr WHERE mr.meter_id = m.id AND mr.billing_month = $1
		)`, month, prevMonth)
	if err != nil {
		log.Printf("autoCreateMonthlyReadings: %v", err)
	}

	// Repair existing rows whose initial_reading is stale because the
	// previous month's current_reading was entered after this month was created.
	if prevMonth != "" {
		db.DB.Exec(`
			UPDATE monthly_readings mr
			SET initial_reading = mr2.current_reading,
				current_reading = CASE
					WHEN mr.current_reading < mr2.current_reading
					THEN mr2.current_reading
					ELSE mr.current_reading
				END,
				updated_at = NOW()
			FROM monthly_readings mr2
			WHERE mr.meter_id = mr2.meter_id
			  AND mr.billing_month = $1
			  AND mr2.billing_month = $2
			  AND mr.initial_reading < mr2.current_reading`, month, prevMonth)
	}
}

func monthOffset(month string, offset int) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return ""
	}
	return t.AddDate(0, offset, 0).Format("2006-01")
}

// cascadeMonthlyReadings propagates a meter's current_reading change forward
// to subsequent months. When a month's current_reading is updated, the next
// month's initial_reading must be updated to match. If the next month's
// current_reading is also below the new initial (e.g., it was auto-created
// and not yet manually set), it gets bumped up too and the cascade continues.
func cascadeMonthlyReadings(meterID, month string, newCurrent int) {
	for {
		nextMonth := monthOffset(month, 1)
		if nextMonth == "" {
			break
		}

		var oldCurrent int
		err := db.DB.QueryRow(
			`SELECT current_reading FROM monthly_readings
			 WHERE meter_id = $1 AND billing_month = $2`,
			meterID, nextMonth,
		).Scan(&oldCurrent)
		if err != nil {
			break // no row for next month, cascade stops
		}

		// Update next month's initial to reflect this month's new current
		db.DB.Exec(`UPDATE monthly_readings SET initial_reading = $1, updated_at = NOW()
			WHERE meter_id = $2 AND billing_month = $3`, newCurrent, meterID, nextMonth)

		// If the next month's current was auto-set or is now below the
		// new initial, bump it up and continue cascading forward
		if oldCurrent < newCurrent {
			db.DB.Exec(`UPDATE monthly_readings SET current_reading = $1, updated_at = NOW()
				WHERE meter_id = $2 AND billing_month = $3`, newCurrent, meterID, nextMonth)
			month = nextMonth
		} else {
			break // manually set to a higher value, cascade stops
		}
	}
}

func formatMonthLabel(month string) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month
	}
	return t.Format("January 2006")
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
		db.DB.Exec(`INSERT INTO settings (key, value, description) VALUES ('electricity_rate', $1, 'Shared rate per unit (electricity & water)')
			ON CONFLICT (key) DO UPDATE SET value=$1, updated_at=NOW()`, rate)
	} else if action == "save_month" {
		month := r.FormValue("month")
		// Save initial_ readings
		for key, vals := range r.Form {
			if strings.HasPrefix(key, "initial_") && len(vals) > 0 && vals[0] != "" {
				meterID := strings.TrimPrefix(key, "initial_")
				reading, _ := strconv.Atoi(vals[0])
				db.DB.Exec(`UPDATE monthly_readings SET initial_reading = $1, updated_at = NOW()
					WHERE meter_id = $2 AND billing_month = $3`, reading, meterID, month)
			}
		}
		// Save current_ readings
		for key, vals := range r.Form {
			if strings.HasPrefix(key, "current_") && len(vals) > 0 && vals[0] != "" {
				meterID := strings.TrimPrefix(key, "current_")
				reading, _ := strconv.Atoi(vals[0])
				db.DB.Exec(`UPDATE monthly_readings SET current_reading = $1, updated_at = NOW()
					WHERE meter_id = $2 AND billing_month = $3`, reading, meterID, month)
				// Cascade forward: update subsequent months' initial_readings
				cascadeMonthlyReadings(meterID, month, reading)
			}
		}
		// Water meter readings
		var waterMeterID string
		db.DB.QueryRow("SELECT id FROM meters WHERE room_id = 'BUILDING' AND meter_type = 'Water' AND is_active = true").Scan(&waterMeterID)
		if waterMeterID != "" {
			if wi := r.FormValue("water_initial"); wi != "" {
				w, _ := strconv.Atoi(wi)
				db.DB.Exec(`UPDATE monthly_readings SET initial_reading = $1, updated_at = NOW()
					WHERE meter_id = $2 AND billing_month = $3`, w, waterMeterID, month)
			}
			if wc := r.FormValue("water_current"); wc != "" {
				w, _ := strconv.Atoi(wc)
				db.DB.Exec(`UPDATE monthly_readings SET current_reading = $1, updated_at = NOW()
					WHERE meter_id = $2 AND billing_month = $3`, w, waterMeterID, month)
				cascadeMonthlyReadings(waterMeterID, month, w)
			}
		}
		// Save water manual override values (units, amount, per-room)
		saveWaterSetting := func(key, val string) {
			if val != "" {
				db.DB.Exec(`INSERT INTO settings (key, value, description) VALUES ($1, $2, 'water override')
					ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`,
					"water_"+key+"_"+month, val)
			}
		}
		saveWaterSetting("units", r.FormValue("water_units"))
		saveWaterSetting("amount", r.FormValue("water_amount"))
		saveWaterSetting("per_room", r.FormValue("water_per_room"))
		http.Redirect(w, r, "/admin/meters?month="+month+"&saved=1", http.StatusSeeOther)
		return
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


// ─── Bill Generation ──────────────────────────────────────────

func handleGenerateBills(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	month := r.FormValue("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	generateBillsForMonth(month)
	http.Redirect(w, r, "/admin/payments?month="+month+"&msg=Bills+generated", http.StatusSeeOther)
}

func generateBillsForMonth(month string) {
	log.Printf("Generating bills for %s...", month)

	var ratePerUnit float64 = 12
	var rateStr string
	db.DB.QueryRow("SELECT value FROM settings WHERE key = 'electricity_rate'").Scan(&rateStr)
	if rateStr != "" {
		if r, err := strconv.ParseFloat(rateStr, 64); err == nil {
			ratePerUnit = r
		}
	}

	var waterPerRoom float64
	var occupiedCount int
	db.DB.QueryRow("SELECT COUNT(DISTINCT room_id) FROM tenants WHERE room_id IS NOT NULL AND (end_date IS NULL OR end_date = '')").Scan(&occupiedCount)
	if occupiedCount > 0 {
		var waterUnits float64
		db.DB.QueryRow(`
			SELECT COALESCE(SUM(mr.current_reading - mr.initial_reading), 0)
			FROM meters m
			JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
			WHERE m.room_id = 'BUILDING' AND m.meter_type = 'Water' AND m.is_active = true`, month).Scan(&waterUnits)
		var savedPerRoom string
		db.DB.QueryRow("SELECT value FROM settings WHERE key = 'water_per_room_' || $1", month).Scan(&savedPerRoom)
		if savedPerRoom != "" {
			if v, err := strconv.ParseFloat(savedPerRoom, 64); err == nil {
				waterPerRoom = v
			}
		} else if waterUnits > 0 {
			waterPerRoom = (waterUnits / float64(occupiedCount)) * ratePerUnit
		}
	}

	rows, err := db.DB.Query(`
		SELECT t.id, COALESCE(t.room_id,''), COALESCE(t.rent_amount, 0), COALESCE(t.has_maintenance, false),
			COALESCE(r.maintenance_amount, 500)
		FROM tenants t
		JOIN rooms r ON t.room_id = r.id
		WHERE t.status = 'active' AND t.room_id IS NOT NULL`)
	if err != nil {
		log.Printf("ERROR generating bills: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var tenantID, roomID string
		var rentAmount, maintAmt float64
		var hasMaint bool
		if err := rows.Scan(&tenantID, &roomID, &rentAmount, &hasMaint, &maintAmt); err != nil {
			continue
		}

		var elecUnits float64
		db.DB.QueryRow(`
			SELECT COALESCE(SUM(mr.current_reading - mr.initial_reading), 0)
			FROM meters m
			JOIN monthly_readings mr ON mr.meter_id = m.id AND mr.billing_month = $1
			WHERE m.room_id = $2 AND m.is_active = true`, month, roomID).Scan(&elecUnits)

		electricity := elecUnits * ratePerUnit
		maintenance := float64(0)
		if hasMaint {
			maintenance = maintAmt
		}
		total := rentAmount + maintenance + electricity + waterPerRoom

		billID := fmt.Sprintf("BILL-%s-%s", tenantID, month)
		_, err := db.DB.Exec(`
			INSERT INTO bills (id, tenant_id, room_id, billing_month,
				rent_amount, maintenance_amount, electricity_amount,
				water_amount, total_amount, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending')
			ON CONFLICT (tenant_id, billing_month)
			DO UPDATE SET rent_amount = $5, maintenance_amount = $6,
				electricity_amount = $7, water_amount = $8,
				total_amount = $9, status = 'pending'`,
			billID, tenantID, roomID, month,
			rentAmount, maintenance, electricity, waterPerRoom, total)
		if err != nil {
			log.Printf("ERROR upserting bill for %s: %v", tenantID, err)
		} else {
			count++
		}
	}
	log.Printf("Generated %d bills for %s", count, month)
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
