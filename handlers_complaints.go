package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"vatsapartment-go/db"
)

// ─── Category Data ──────────────────────────────────────────────

var complaintCategories = map[string][]string{
	"AC":                {"Not cooling", "Gas leakage", "Water dripping from indoor unit", "Not turning on", "Strange noise", "Remote not working"},
	"Fridge":            {"Not cooling", "Water leakage", "Ice buildup", "Strange noise", "Door not sealing", "Light not working"},
	"Water Purifier":    {"Not dispensing water", "Bad taste/smell", "Water leakage", "Filter change needed", "Slow flow"},
	"WiFi / Internet":   {"WiFi not working", "Slow internet", "No connectivity", "Router issue"},
	"Bathroom":          {"Tap not working/leaking", "Toilet seat broken/leaking", "Flush not working", "Flush leaking water", "Drain clogged", "No hot water (geyser)", "Exhaust fan not working"},
	"Electrical/Lights": {"Tube light/bulb not working", "Switch/socket broken", "Fan not working/noisy", "MCB tripping", "Geyser not working"},
	"Plumbing":          {"Pipe leaking", "Low water pressure", "No water supply", "Drainage issue"},
	"Furniture":         {"Bed broken/creaking", "Study table damaged", "Chair broken", "Cupboard not closing", "Mattress issue"},
	"Doors & Windows":   {"Door lock broken", "Door hinge issue", "Window glass broken", "Window not closing"},
	"Others":            {},
}

var categoryOrder = []string{
	"AC", "Fridge", "Water Purifier", "WiFi / Internet",
	"Bathroom", "Electrical/Lights", "Plumbing",
	"Furniture", "Doors & Windows", "Others",
}

var complaintCatIcons = map[string]string{
	"AC":                "cooling",
	"Fridge":            "tv",
	"Water Purifier":    "shower",
	"WiFi / Internet":   "wifi",
	"Bathroom":          "shower",
	"Electrical/Lights": "cooling",
	"Plumbing":          "shower",
	"Furniture":         "chair",
	"Doors & Windows":   "door",
	"Others":            "search",
}

var maintenanceItemTypes = []string{"ac", "fridge", "water_purifier", "wifi", "geyser"}

var maintenanceItemLabels = map[string]string{
	"ac":             "AC",
	"fridge":         "Fridge",
	"water_purifier": "Water Purifier",
	"wifi":           "WiFi",
	"geyser":         "Geyser",
}

// ─── Tenant Complaint Submission ─────────────────────────────────

func handleTenantComplaints(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	if r.Method == "POST" {
		handleTenantComplaintSubmit(w, r, tenantID)
		return
	}

	// GET: render the complaint form
	msg := r.URL.Query().Get("msg")
	errMsg := r.URL.Query().Get("error")

	catsJSON, _ := json.Marshal(complaintCategories)

	renderPrivate(w, "tenant_complaints.html", map[string]interface{}{
		"CategoryOrder":  categoryOrder,
		"CatIcons":       complaintCatIcons,
		"CategoriesJSON": string(catsJSON),
		"Active":         "complaints",
		"Title":          "Report a Problem",
		"Msg":            msg,
		"Error":          errMsg,
	})
}

func handleTenantComplaintSubmit(w http.ResponseWriter, r *http.Request, tenantID string) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Redirect(w, r, "/tenant/complaints?error=File+too+large+(max+10MB)", http.StatusSeeOther)
		return
	}

	category := r.FormValue("category")
	subcategory := r.FormValue("subcategory")
	description := r.FormValue("description")

	// Validate category
	_, valid := complaintCategories[category]
	if !valid {
		http.Redirect(w, r, "/tenant/complaints?error=Invalid+category", http.StatusSeeOther)
		return
	}

	// For "Others", subcategory is empty — use description
	if category == "Others" {
		if description == "" {
			http.Redirect(w, r, "/tenant/complaints?error=Please+describe+the+issue", http.StatusSeeOther)
			return
		}
		subcategory = "Other issue"
	}

	// Validate description is not empty
	if description == "" {
		http.Redirect(w, r, "/tenant/complaints?error=Please+describe+the+issue", http.StatusSeeOther)
		return
	}

	// Get tenant's room
	var roomID string
	err := db.DB.QueryRow("SELECT COALESCE(room_id,'') FROM tenants WHERE id = $1", tenantID).Scan(&roomID)
	if err != nil || roomID == "" {
		http.Redirect(w, r, "/tenant/complaints?error=No+room+assigned", http.StatusSeeOther)
		return
	}

	// Handle image upload
	var imagePath string
	imgFile, imgHeader, imgErr := r.FormFile("image")
	if imgErr == nil {
		defer imgFile.Close()
		ext := ""
		contentType := imgHeader.Header.Get("Content-Type")
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
		filename := fmt.Sprintf("cmp_%s_%d%s", tenantID, time.Now().UnixNano(), ext)
		dst, err := os.Create(filepath.Join("uploads", filename))
		if err == nil {
			io.Copy(dst, imgFile)
			dst.Close()
			imagePath = "/uploads/" + filename
		}
	}

	id := fmt.Sprintf("CMP%d", time.Now().UnixNano())
	_, err = db.DB.Exec(`
		INSERT INTO complaints (id, tenant_id, room_id, category, subcategory, description, image, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')`,
		id, tenantID, roomID, category, subcategory, description, imagePath)
	if err != nil {
		log.Printf("ERROR creating complaint: %v", err)
		http.Redirect(w, r, "/tenant/complaints?error=Failed+to+submit+complaint", http.StatusSeeOther)
		return
	}

	log.Printf("Complaint %s submitted by tenant %s for room %s: %s — %s", id, tenantID, roomID, category, subcategory)
	http.Redirect(w, r, "/tenant/complaints?msg=Complaint+submitted+successfully", http.StatusSeeOther)
}

// ─── Admin Complaint Management ──────────────────────────────────

func handleAdminComplaints(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "complaints"
	}

	filterStatus := r.URL.Query().Get("status")
	filterRoom := r.URL.Query().Get("room")

	// Build query
	query := `
		SELECT c.id, c.tenant_id, c.room_id, c.category, c.subcategory,
			   COALESCE(c.description, '') as description,
			   COALESCE(c.image, '') as image,
			   c.status,
			   COALESCE(c.admin_notes, '') as admin_notes,
			   TO_CHAR(c.created_at, 'Mon DD, YYYY HH24:MI') as created_at,
			   COALESCE(t.name, 'Unknown') as tenant_name,
			   COALESCE(r.room_number, '--') as room_number
		FROM complaints c
		LEFT JOIN tenants t ON c.tenant_id = t.id
		LEFT JOIN rooms r ON c.room_id = r.id
		WHERE 1=1`
	var args []interface{}
	argN := 1

	if filterStatus != "" {
		query += fmt.Sprintf(" AND c.status = $%d", argN)
		args = append(args, filterStatus)
		argN++
	}
	if filterRoom != "" {
		query += fmt.Sprintf(" AND c.room_id = $%d", argN)
		args = append(args, filterRoom)
		argN++
	}
	query += " ORDER BY c.created_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		log.Printf("ERROR loading complaints: %v", err)
		renderAdminError(w, "Failed to load complaints")
		return
	}
	defer rows.Close()

	type ComplaintRow struct {
		ID, TenantID, RoomID, Category, Subcategory, Description, Image, Status, AdminNotes, CreatedAt string
		TenantName, RoomNumber string
	}
	var complaints []ComplaintRow
	for rows.Next() {
		var cr ComplaintRow
		if err := rows.Scan(&cr.ID, &cr.TenantID, &cr.RoomID, &cr.Category, &cr.Subcategory,
			&cr.Description, &cr.Image, &cr.Status, &cr.AdminNotes, &cr.CreatedAt,
			&cr.TenantName, &cr.RoomNumber); err != nil {
			log.Printf("ERROR scanning complaint: %v", err)
			continue
		}
		complaints = append(complaints, cr)
	}
	if complaints == nil {
		complaints = []ComplaintRow{}
	}

	// Status counts
	var pending, inProgress, resolved int
	db.DB.QueryRow(`
		SELECT
			COALESCE(COUNT(*) FILTER (WHERE status = 'pending'), 0),
			COALESCE(COUNT(*) FILTER (WHERE status = 'in_progress'), 0),
			COALESCE(COUNT(*) FILTER (WHERE status = 'resolved'), 0)
		FROM complaints`).Scan(&pending, &inProgress, &resolved)

	// Room list for filter dropdown
	roomRows, err := db.DB.Query("SELECT id, room_number FROM rooms WHERE id != 'BUILDING' ORDER BY room_number")
	type RoomOpt struct{ ID, RoomNumber string }
	var rooms []RoomOpt
	if err == nil {
		defer roomRows.Close()
		for roomRows.Next() {
			var ro RoomOpt
			roomRows.Scan(&ro.ID, &ro.RoomNumber)
			rooms = append(rooms, ro)
		}
	}
	if rooms == nil {
		rooms = []RoomOpt{}
	}

	// Maintenance log data — build lookup map: roomID -> itemType -> {date, notes}
	maintLookup := make(map[string]map[string]map[string]string)
	maintRows, err := db.DB.Query(`
		SELECT rm.room_id, rm.item_type,
			   COALESCE(rm.last_serviced_date,''), COALESCE(rm.notes,'')
		FROM room_maintenance rm
		ORDER BY rm.room_id, rm.item_type`)
	if err == nil {
		defer maintRows.Close()
		for maintRows.Next() {
			var roomID, itemType, date, notes string
			maintRows.Scan(&roomID, &itemType, &date, &notes)
			if maintLookup[roomID] == nil {
				maintLookup[roomID] = make(map[string]map[string]string)
			}
			maintLookup[roomID][itemType] = map[string]string{"date": date, "notes": notes}
		}
	}

	// Rooms for maintenance log grid
	allRooms, _ := getRoomList()

	// Serialize for JS in maintenance log modal
	allRoomsJSON, _ := json.Marshal(allRooms)
	labelsJSON, _ := json.Marshal(maintenanceItemLabels)
	maintLookupJSON, _ := json.Marshal(maintLookup)

	renderPrivate(w, "admin_complaints.html", map[string]interface{}{
		"Complaints":        complaints,
		"StatusCounts":      map[string]int{"Pending": pending, "InProgress": inProgress, "Resolved": resolved},
		"FilterStatus":      filterStatus,
		"FilterRoom":        filterRoom,
		"Rooms":             rooms,
		"AllRooms":          allRooms,
		"AllRoomsJSON":      string(allRoomsJSON),
		"MaintenanceItems":  maintenanceItemTypes,
		"MaintenanceLabels": maintenanceItemLabels,
		"MaintenanceLabelsJSON": string(labelsJSON),
		"MaintenanceLookup": maintLookup,
		"MaintenanceLookupJSON": string(maintLookupJSON),
		"Tab":               tab,
		"Active":            "complaints",
		"Title":             "Complaints",
	})
}

func handleAdminComplaintsSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	action := r.FormValue("action")
	id := r.FormValue("id")

	if action == "update_status" {
		status := r.FormValue("status")
		if status == "pending" || status == "in_progress" || status == "resolved" {
			db.DB.Exec("UPDATE complaints SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
			log.Printf("Complaint %s status updated to %s", id, status)
		}
	} else if action == "add_notes" {
		notes := r.FormValue("admin_notes")
		db.DB.Exec("UPDATE complaints SET admin_notes = $1, updated_at = NOW() WHERE id = $2", notes, id)
		log.Printf("Admin notes added to complaint %s", id)
	}

	// Preserve filters on redirect
	redirectURL := "/admin/complaints"
	if ref := r.Referer(); ref != "" {
		http.Redirect(w, r, ref, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// ─── Admin Maintenance Log ───────────────────────────────────────

func handleAdminMaintenanceLogSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	r.ParseForm()
	roomID := r.FormValue("room_id")
	itemType := r.FormValue("item_type")
	lastServicedDate := r.FormValue("last_serviced_date")
	notes := r.FormValue("notes")

	if roomID == "" || itemType == "" {
		http.Redirect(w, r, "/admin/complaints?tab=maintenance&error=Missing+room+or+item", http.StatusSeeOther)
		return
	}

	// Upsert: update if exists, insert if not
	id := fmt.Sprintf("RM_%s_%s", roomID, itemType)
	_, err := db.DB.Exec(`
		INSERT INTO room_maintenance (id, room_id, item_type, last_serviced_date, notes, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (room_id, item_type)
		DO UPDATE SET last_serviced_date = $4, notes = $5, updated_at = NOW()`,
		id, roomID, itemType, lastServicedDate, notes)
	if err != nil {
		log.Printf("ERROR saving maintenance log: %v", err)
		http.Redirect(w, r, "/admin/complaints?tab=maintenance&error=Failed+to+save", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/complaints?tab=maintenance&msg=Maintenance+date+updated", http.StatusSeeOther)
}

// ─── Helpers ─────────────────────────────────────────────────────

func getRoomList() ([]map[string]string, error) {
	rows, err := db.DB.Query("SELECT id, room_number FROM rooms WHERE id != 'BUILDING' ORDER BY room_number")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []map[string]string
	for rows.Next() {
		var id, num string
		rows.Scan(&id, &num)
		rooms = append(rooms, map[string]string{"ID": id, "RoomNumber": num})
	}
	if rooms == nil {
		rooms = []map[string]string{}
	}
	return rooms, nil
}
