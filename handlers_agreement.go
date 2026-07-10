package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"vatsapartment-go/db"
)

// ─── Agreement data ─────────────────────────────────────────

type agreementData struct {
	Active            string
	Title             string
	TenantID          string
	TenantName        string
	TenantPhone       string
	TenantEmail       string
	RoomNumber        string
	RentAmount        float64
	SecurityDeposit   float64
	LockInPeriod      int
	CheckInDate       string
	EndDate           string
	HasMaintenance    bool
	MaintenanceAmount float64
	ElectricityRate   float64
	PropertyAddress   string
	AgreementVersion  string
	GeneratedAt       string
	AgreementTerms    string // processed terms with placeholders replaced
	AgreementNo       string
	IsPrint           bool
}

func buildAgreementData(tenantID string) agreementData {
	var d agreementData
	d.TenantID = tenantID
	d.GeneratedAt = time.Now().Format("02 January 2006")

	// Query tenant + room
	var endDateStr string
	var maintAmt float64
	err := db.DB.QueryRow(`
		SELECT t.name, t.phone, COALESCE(t.email, ''),
			COALESCE(t.rent_amount, 0), COALESCE(t.security_deposit, 0),
			COALESCE(t.security_lock_in_period, 0),
			COALESCE(t.check_in_date::text, ''),
			COALESCE(t.end_date, ''),
			COALESCE(t.has_maintenance, false),
			COALESCE(r.room_number, ''),
			COALESCE(r.maintenance_amount, 500)
		FROM tenants t
		LEFT JOIN rooms r ON t.room_id = r.id
		WHERE t.id = $1`, tenantID).Scan(
		&d.TenantName, &d.TenantPhone, &d.TenantEmail,
		&d.RentAmount, &d.SecurityDeposit, &d.LockInPeriod,
		&d.CheckInDate, &endDateStr, &d.HasMaintenance,
		&d.RoomNumber, &maintAmt)
	if err != nil {
		log.Printf("ERROR building agreement data for %s: %v", tenantID, err)
		return d
	}
	d.MaintenanceAmount = maintAmt
	if endDateStr != "" {
		d.EndDate = endDateStr
	}

	// Read settings
	var terms, address, rateStr string
	db.DB.QueryRow(`SELECT COALESCE((SELECT value FROM settings WHERE key = 'agreement_terms'), '')`).Scan(&terms)
	db.DB.QueryRow(`SELECT COALESCE((SELECT value FROM settings WHERE key = 'property_address'), '')`).Scan(&address)
	db.DB.QueryRow(`SELECT COALESCE((SELECT value FROM settings WHERE key = 'electricity_rate'), '12')`).Scan(&rateStr)

	d.PropertyAddress = address
	if d.PropertyAddress == "" {
		d.PropertyAddress = "Near Apna Chai Wala, LPU, Jalandhar, Punjab"
	}

	// Parse electricity rate
	d.ElectricityRate = 12
	if rateStr != "" {
		var r float64
		if _, err := fmt.Sscanf(rateStr, "%f", &r); err == nil {
			d.ElectricityRate = r
		}
	}

	// Replace placeholders in terms
	terms = strings.ReplaceAll(terms, "[RENT_AMOUNT]", fmt.Sprintf("%.0f", d.RentAmount))
	terms = strings.ReplaceAll(terms, "[DEPOSIT_AMOUNT]", fmt.Sprintf("%.0f", d.SecurityDeposit))
	terms = strings.ReplaceAll(terms, "[LOCK_IN_PERIOD]", fmt.Sprintf("%d", d.LockInPeriod))
	terms = strings.ReplaceAll(terms, "[ELECTRICITY_RATE]", fmt.Sprintf("%.0f", d.ElectricityRate))

	// Format terms for HTML: split on blank lines into paragraphs,
	// then replace single newlines with <br> within each paragraph
	var paragraphs []string
	for _, para := range strings.Split(terms, "\n\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		para = strings.ReplaceAll(para, "\n", "<br>")
		paragraphs = append(paragraphs, "<p>"+para+"</p>")
	}
	d.AgreementTerms = strings.Join(paragraphs, "\n")

	// Agreement number
	d.AgreementNo = fmt.Sprintf("VATS/AG/%s/%s", time.Now().Format("2006-01"), tenantID)

	return d
}

// ─── Admin agreement view ───────────────────────────────────

func handleAdminTenantAgreement(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	tenantID := r.URL.Query().Get("id")
	if tenantID == "" {
		http.Redirect(w, r, "/admin/tenants?error=Missing+tenant+ID", http.StatusSeeOther)
		return
	}

	d := buildAgreementData(tenantID)
	if d.TenantName == "" {
		http.Redirect(w, r, "/admin/tenants?error=Tenant+not+found", http.StatusSeeOther)
		return
	}
	d.Active = "tenants"
	d.Title = "Rent Agreement — " + d.TenantName

	renderPrivate(w, "admin_agreement.html", d)
}

// ─── Tenant agreement view ──────────────────────────────────

func handleTenantAgreement(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := requireTenant(w, r)
	if !ok {
		return
	}

	d := buildAgreementData(tenantID)
	d.Active = "agreement"
	d.Title = "My Rent Agreement"

	renderPrivate(w, "tenant_agreement.html", d)
}

// ─── Print agreement (standalone, no layout) ───────────────

func handleAgreementPrint(w http.ResponseWriter, r *http.Request) {
	// Determine tenant ID — admin passes ?id=, tenant uses session
	var tenantID string

	// Try admin auth first
	adminCookie, _ := r.Cookie("session")
	if adminCookie != nil && adminCookie.Value != "" {
		// admin is logged in, get tenant ID from query param
		tenantID = r.URL.Query().Get("id")
	} else {
		// Try tenant auth
		tenantCookie, err := r.Cookie("tenant_session")
		if err == nil && tenantCookie.Value != "" {
			parts := strings.SplitN(tenantCookie.Value, ":", 2)
			if len(parts) == 2 && parts[0] == "tenant" {
				tenantID = parts[1]
			}
		}
	}

	if tenantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	d := buildAgreementData(tenantID)
	if d.TenantName == "" {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}
	d.IsPrint = true

	render(w, "agreement_print.html", d)
}
