package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func Init() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://neondb_owner:npg_rIeMGo3CVYF9@ep-tiny-resonance-abne6jfo-pooler.eu-west-2.aws.neon.tech/neondb?sslmode=require"
	}

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	DB.SetMaxOpenConns(5)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(1 * time.Minute)
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	log.Println("DB connected, running migrations...")
	if err := runMigrations(); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	log.Println("Migrations complete")
	return nil
}

func runMigrations() error {
	migrations := []struct {
		name string
		sql  string
	}{
		{"users", `
			CREATE TABLE IF NOT EXISTS users (
				id TEXT PRIMARY KEY,
				username TEXT NOT NULL UNIQUE,
				password_hash TEXT NOT NULL,
				role TEXT DEFAULT 'admin',
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"rooms", `
			CREATE TABLE IF NOT EXISTS rooms (
				id TEXT PRIMARY KEY,
				room_number TEXT NOT NULL,
				floor INTEGER NOT NULL,
				type TEXT NOT NULL,
				price DOUBLE PRECISION NOT NULL,
				images TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"tenants", `
			CREATE TABLE IF NOT EXISTS tenants (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				email TEXT UNIQUE,
				phone TEXT NOT NULL UNIQUE,
				aadhaar_number TEXT,
				alternate_contact TEXT,
				emergency_contact TEXT,
				address TEXT,
				status TEXT DEFAULT 'active',
				agreement_document TEXT,
				document_photo TEXT,
				college_id TEXT,
				check_in_date TEXT,
				security_deposit DOUBLE PRECISION,
				security_lock_in_period INTEGER,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"payments", `
			CREATE TABLE IF NOT EXISTS payments (
				id TEXT PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				amount DOUBLE PRECISION NOT NULL,
				payment_date TEXT NOT NULL,
				payment_method TEXT,
				transaction_id TEXT UNIQUE,
				status TEXT DEFAULT 'pending',
				late_fee DOUBLE PRECISION DEFAULT 0,
				month_covered TEXT,
				notes TEXT,
				receipt_url TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"bookings", `
			CREATE TABLE IF NOT EXISTS bookings (
				id TEXT PRIMARY KEY,
				room_id TEXT NOT NULL,
				rent_amount DOUBLE PRECISION NOT NULL,
				security_deposit DOUBLE PRECISION,
				security_lock_in_period INTEGER,
				check_in_date TEXT NOT NULL,
				check_out_date TEXT,
				status TEXT DEFAULT 'active',
				notes TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"people", `
			CREATE TABLE IF NOT EXISTS people (
				id TEXT PRIMARY KEY,
				booking_id TEXT NOT NULL,
				name TEXT NOT NULL,
				phone TEXT NOT NULL,
				email TEXT,
				aadhaar_number TEXT,
				document_photo TEXT,
				college_id TEXT,
				is_primary INTEGER DEFAULT 0,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"room_assignments", `
			CREATE TABLE IF NOT EXISTS room_assignments (
				id TEXT PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				room_id TEXT NOT NULL,
				rent_amount DOUBLE PRECISION NOT NULL,
				start_date TEXT NOT NULL,
				end_date TEXT,
				is_active INTEGER DEFAULT 1,
				share_percentage DOUBLE PRECISION,
				notes TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"meters", `
			CREATE TABLE IF NOT EXISTS meters (
				id TEXT PRIMARY KEY,
				room_id TEXT NOT NULL,
				meter_type TEXT NOT NULL,
				meter_number TEXT NOT NULL,
				meter_name TEXT,
				initial_reading INTEGER DEFAULT 0,
				current_reading INTEGER DEFAULT 0,
				installation_date TEXT,
				is_active INTEGER DEFAULT 1,
				notes TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"meter_readings", `
			CREATE TABLE IF NOT EXISTS meter_readings (
				id TEXT PRIMARY KEY,
				meter_id TEXT NOT NULL,
				reading INTEGER NOT NULL,
				reading_date TEXT NOT NULL,
				photo_url TEXT,
				extracted_by_ai INTEGER DEFAULT 0,
				ai_provider TEXT,
				ai_confidence TEXT,
				billing_period TEXT,
				notes TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"settings", `
			CREATE TABLE IF NOT EXISTS settings (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				description TEXT,
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"media", `
			CREATE TABLE IF NOT EXISTS media (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL UNIQUE,
				content_type TEXT NOT NULL,
				data BYTEA NOT NULL,
				created_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"blog_posts", `
			CREATE TABLE IF NOT EXISTS blog_posts (
				id TEXT PRIMARY KEY,
				title TEXT NOT NULL,
				slug TEXT NOT NULL UNIQUE,
				excerpt TEXT,
				content TEXT NOT NULL,
				image_url TEXT,
				author TEXT DEFAULT 'Vats Apartment',
				status TEXT DEFAULT 'draft',
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"api_keys", `
			CREATE TABLE IF NOT EXISTS api_keys (
				id TEXT PRIMARY KEY,
				key TEXT NOT NULL UNIQUE,
				name TEXT NOT NULL,
				permissions TEXT DEFAULT 'blog',
				created_at TIMESTAMPTZ DEFAULT NOW(),
				last_used_at TIMESTAMPTZ
			)`},
		{"tenant_password", `
			ALTER TABLE tenants ADD COLUMN IF NOT EXISTS password_hash TEXT`},
		{"tenant_verifications", `
			CREATE TABLE IF NOT EXISTS tenant_verifications (
				id TEXT PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				lpu_id_photo TEXT,
				aadhar_photo TEXT,
				status TEXT DEFAULT 'not_submitted',
				submitted_at TIMESTAMPTZ,
				verified_at TIMESTAMPTZ,
				notes TEXT,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"tenant_passes", `
			CREATE TABLE IF NOT EXISTS tenant_passes (
				id TEXT PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				pass_number TEXT NOT NULL UNIQUE,
				issued_by TEXT,
				issued_at TIMESTAMPTZ DEFAULT NOW(),
				is_active INTEGER DEFAULT 1,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			)`},
		{"idx_tv_tenant", `CREATE INDEX IF NOT EXISTS idx_tv_tenant ON tenant_verifications(tenant_id)`},
		{"idx_ra_tenant", `CREATE INDEX IF NOT EXISTS idx_ra_tenant ON room_assignments(tenant_id)`},
		{"idx_pay_tenant", `CREATE INDEX IF NOT EXISTS idx_pay_tenant ON payments(tenant_id)`},
		{"idx_tenants_status", `CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status)`},
	{"building_room", `INSERT INTO rooms (id, room_number, floor, type, price) VALUES ('BUILDING', 'BLDG', 0, 'Utility', 0) ON CONFLICT (id) DO NOTHING`},
	{"building_water_meter", `INSERT INTO meters (id, room_id, meter_type, meter_number, initial_reading, current_reading, is_active) VALUES ('MTR-BUILDING-Water', 'BUILDING', 'Water', 'BLDG-Water', 0, 0, true) ON CONFLICT (id) DO NOTHING`},
	{"monthly_readings", `
		CREATE TABLE IF NOT EXISTS monthly_readings (
			id TEXT PRIMARY KEY,
			meter_id TEXT NOT NULL,
			billing_month TEXT NOT NULL,
			initial_reading INTEGER DEFAULT 0,
			current_reading INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(meter_id, billing_month)
		)`},
	{"seed_monthly_readings", `
		INSERT INTO monthly_readings (id, meter_id, billing_month, initial_reading, current_reading)
		SELECT 'MR-' || id || '-' || TO_CHAR(NOW(), 'YYYY-MM'),
			id, TO_CHAR(NOW(), 'YYYY-MM'), initial_reading, current_reading
		FROM meters WHERE is_active = true
		ON CONFLICT (meter_id, billing_month) DO NOTHING`},
	{"idx_mr_month", `CREATE INDEX IF NOT EXISTS idx_mr_month ON monthly_readings(billing_month)`},
	{"idx_mr_meter", `CREATE INDEX IF NOT EXISTS idx_mr_meter ON monthly_readings(meter_id)`},
		{"tenant_maintenance", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS has_maintenance BOOLEAN DEFAULT false`},
		{"maintenance_amount_setting", `INSERT INTO settings (key, value) VALUES ('maintenance_amount', '500') ON CONFLICT (key) DO NOTHING`},
		{"room_maintenance_amount", `ALTER TABLE rooms ADD COLUMN IF NOT EXISTS maintenance_amount DOUBLE PRECISION DEFAULT 500`},

		// ─── Bills table ──────────────────────────────────
		{"bills", `
			CREATE TABLE IF NOT EXISTS bills (
				id TEXT PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				room_id TEXT NOT NULL,
				billing_month TEXT NOT NULL,
				rent_amount DOUBLE PRECISION NOT NULL,
				maintenance_amount DOUBLE PRECISION DEFAULT 0,
				electricity_amount DOUBLE PRECISION DEFAULT 0,
				water_amount DOUBLE PRECISION DEFAULT 0,
				late_fee DOUBLE PRECISION DEFAULT 0,
				total_amount DOUBLE PRECISION NOT NULL,
				status TEXT DEFAULT 'pending',
				generated_at TIMESTAMPTZ DEFAULT NOW(),
				notes TEXT,
				UNIQUE(tenant_id, billing_month)
			)`},
		{"bill_discount", `ALTER TABLE bills ADD COLUMN IF NOT EXISTS discount_amount DOUBLE PRECISION DEFAULT 0`},
		{"bill_discount_note", `ALTER TABLE bills ADD COLUMN IF NOT EXISTS discount_note TEXT DEFAULT ''`},
		{"payment_bill_id", `ALTER TABLE payments ADD COLUMN IF NOT EXISTS bill_id TEXT`},
		// ─── Indexes ──────────────────────────────────────
		{"idx_ra_room_id", `CREATE INDEX IF NOT EXISTS idx_ra_room_id ON room_assignments(room_id)`},
		{"idx_ra_tenant_active", `CREATE INDEX IF NOT EXISTS idx_ra_tenant_active ON room_assignments(tenant_id, is_active)`},
		{"idx_meters_room_id", `CREATE INDEX IF NOT EXISTS idx_meters_room_id ON meters(room_id)`},
		{"idx_mr_meter_id", `CREATE INDEX IF NOT EXISTS idx_mr_meter_id ON meter_readings(meter_id)`},
		{"idx_pay_month_covered", `CREATE INDEX IF NOT EXISTS idx_pay_month_covered ON payments(month_covered)`},
		{"idx_pay_status", `CREATE INDEX IF NOT EXISTS idx_pay_status ON payments(status)`},
		{"idx_pay_tenant_month", `CREATE INDEX IF NOT EXISTS idx_pay_tenant_month ON payments(tenant_id, month_covered)`},
		{"idx_tenants_phone", `CREATE INDEX IF NOT EXISTS idx_tenants_phone ON tenants(phone)`},
		{"idx_blog_slug", `CREATE INDEX IF NOT EXISTS idx_blog_slug ON blog_posts(slug)`},
		{"idx_blog_status", `CREATE INDEX IF NOT EXISTS idx_blog_status ON blog_posts(status)`},
		{"idx_bills_tenant_month", `CREATE INDEX IF NOT EXISTS idx_bills_tenant_month ON bills(tenant_id, billing_month)`},
		{"idx_bills_status", `CREATE INDEX IF NOT EXISTS idx_bills_status ON bills(status)`},
		// ─── Merge passes+verifications into tenants ─────
		{"t_pass_number", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS pass_number TEXT`},
		{"t_pass_issued", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS pass_issued_at TIMESTAMPTZ`},
		{"t_lpu_photo", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS lpu_id_photo TEXT`},
		{"t_aadhar_photo", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS aadhar_photo TEXT`},
		{"t_ver_status", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS verification_status TEXT DEFAULT 'not_submitted'`},
		{"t_ver_submitted", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS verification_submitted_at TIMESTAMPTZ`},
		{"t_ver_verified", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS verification_verified_at TIMESTAMPTZ`},
		{"t_ver_notes", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS verification_notes TEXT`},
		// ─── Merge room_assignments into tenants ──────────
		{"t_room_id", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS room_id TEXT`},
		{"t_rent_amount", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS rent_amount DOUBLE PRECISION`},
		{"t_end_date", `ALTER TABLE tenants ADD COLUMN IF NOT EXISTS end_date TEXT`},
		// ─── Backfill from old tables before dropping ─────
		{"backfill_pass", `DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tenant_passes') THEN UPDATE tenants t SET pass_number = tp.pass_number, pass_issued_at = tp.issued_at FROM tenant_passes tp WHERE tp.tenant_id = t.id AND tp.is_active = 1 AND t.pass_number IS NULL; END IF; END $$`},
		{"backfill_ver", `DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tenant_verifications') THEN UPDATE tenants t SET lpu_id_photo = tv.lpu_id_photo, aadhar_photo = tv.aadhar_photo, verification_status = COALESCE(tv.status, 'not_submitted'), verification_submitted_at = tv.submitted_at, verification_verified_at = tv.verified_at, verification_notes = tv.notes FROM tenant_verifications tv WHERE tv.tenant_id = t.id AND t.verification_status = 'not_submitted' AND tv.created_at = (SELECT MAX(tv2.created_at) FROM tenant_verifications tv2 WHERE tv2.tenant_id = t.id); END IF; END $$`},
		{"backfill_ra", `DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'room_assignments') THEN UPDATE tenants t SET room_id = ra.room_id, rent_amount = COALESCE(ra.rent_amount, 0), end_date = ra.end_date, check_in_date = COALESCE(t.check_in_date, ra.start_date::timestamp) FROM room_assignments ra WHERE ra.tenant_id = t.id AND ra.is_active = 1 AND t.room_id IS NULL; END IF; END $$`},
		// ─── Drop dead/merged tables ──────────────────────
		{"drop_tp", `DROP TABLE IF EXISTS tenant_passes`},
		{"drop_tv", `DROP TABLE IF EXISTS tenant_verifications`},
		{"drop_pp", `DROP TABLE IF EXISTS people`},
		{"drop_bk", `DROP TABLE IF EXISTS bookings`},
		{"drop_ra", `DROP TABLE IF EXISTS room_assignments`},

		{"payment_paid_to", `ALTER TABLE payments ADD COLUMN IF NOT EXISTS paid_to TEXT DEFAULT ''`},
	}

	for _, m := range migrations {
		if _, err := DB.Exec(m.sql); err != nil {
			return fmt.Errorf("%s: %w", m.name, err)
		}
		log.Printf("  ✓ %s", m.name)
	}

	return nil
}

func SeedRooms() error {
	rooms := []struct {
		id, roomNumber, roomType string
		floor                    int
		price                    float64
	}{
		{"101", "101", "Standard Double", 1, 9000},
		{"102", "102", "Standard Double", 1, 10000},
		{"201", "201", "Standard Double", 2, 9000},
		{"202", "202", "Standard Double", 2, 10000},
		{"301", "301", "Standard Double", 3, 9000},
		{"302", "302", "Standard Double", 3, 9000},
	}
	for _, r := range rooms {
		_, err := DB.Exec(
			`INSERT INTO rooms (id, room_number, floor, type, price) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO UPDATE SET room_number=$2, floor=$3, type=$4, price=$5`,
			r.id, r.roomNumber, r.floor, r.roomType, r.price,
		)
		if err != nil {
			return err
		}
	}
	_, err := DB.Exec(`DELETE FROM rooms WHERE id NOT IN ('101', '102', '201', '202', '301', '302', 'BUILDING')`)
	return err
}

func SeedAdmin() error {
	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		log.Println("Admin user already exists, skipping seed")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("Vats@2024"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = DB.Exec(
		`INSERT INTO users (id, username, password_hash, role) VALUES ($1, $2, $3, 'admin')`,
		"usr_admin", "admin", string(hash),
	)
	if err != nil {
		return err
	}
	log.Println("Admin user seeded: admin / Vats@2024")
	return nil
}

func SeedMedia() error {
	media := []struct {
		id, name, ctype, path string
	}{
		{"med_logo", "logo.webp", "image/webp", "static/logo.webp"},
		{"med_white_logo", "white-logo.webp", "image/webp", "static/white-logo.webp"},
		{"med_hero", "hero-2.webp", "image/webp", "static/hero-2.webp"},
		{"med_favicon", "favicon.ico", "image/x-icon", "static/favicon.ico"},
		{"med_photo360", "photo360.webp", "image/webp", "static/photo360.webp"},
		{"med_video", "vats-apartment-tour.mp4", "video/mp4", "static/vats-apartment-tour.mp4"},
	}

	for _, m := range media {
		var count int
		DB.QueryRow("SELECT COUNT(*) FROM media WHERE name = $1", m.name).Scan(&count)
		if count > 0 {
			log.Printf("Media %s already exists, skipping", m.name)
			continue
		}

		data, err := os.ReadFile(m.path)
		if err != nil {
			log.Printf("WARNING: Cannot read %s: %v", m.path, err)
			continue
		}

		_, err = DB.Exec(
			`INSERT INTO media (id, name, content_type, data) VALUES ($1, $2, $3, $4)`,
			m.id, m.name, m.ctype, data,
		)
		if err != nil {
			return err
		}
		log.Printf("Media seeded: %s (%d bytes)", m.name, len(data))
	}
	return nil
}

// SeedMeters creates default meters for all rooms (Electric, Inverter)
// plus a building-level water meter. Safe to run multiple times (idempotent).
// Also deduplicates — keeps only 1 meter per room+type (best one: has meter_number or highest reading).
func SeedMeters() error {
		// Remove per-room water meters — only building water meter should exist
		DB.Exec(`DELETE FROM monthly_readings WHERE meter_id IN (SELECT id FROM meters WHERE meter_type = 'Water' AND room_id != 'BUILDING')`)
		DB.Exec(`DELETE FROM meter_readings WHERE meter_id IN (SELECT id FROM meters WHERE meter_type = 'Water' AND room_id != 'BUILDING')`)
			DB.Exec(`DELETE FROM meters WHERE meter_type = 'Water' AND room_id != 'BUILDING'`)
	// Deduplicate: keep only the best meter per room_id + meter_type
	dedupRows, err := DB.Query(`SELECT room_id, meter_type FROM meters WHERE room_id != 'BUILDING' GROUP BY room_id, meter_type HAVING COUNT(*) > 1`)
	if err == nil {
		defer dedupRows.Close()
		for dedupRows.Next() {
			var rid, mtype string
			dedupRows.Scan(&rid, &mtype)
			// Keep the one with a non-empty meter_number, or highest current_reading
			var bestID string
			DB.QueryRow(`SELECT id FROM meters WHERE room_id = $1 AND meter_type = $2
				ORDER BY CASE WHEN meter_number != '' THEN 0 ELSE 1 END, current_reading DESC LIMIT 1`, rid, mtype).Scan(&bestID)
			if bestID != "" {
				res, _ := DB.Exec(`DELETE FROM meters WHERE room_id = $1 AND meter_type = $2 AND id != $3`, rid, mtype, bestID)
				if n, _ := res.RowsAffected(); n > 0 {
					log.Printf("  ✓ Dedup: Room %s %s — removed %d duplicate(s)", rid, mtype, n)
				}
			}
		}
	}

	type roomInfo struct{ id, number string }
	rows, err := DB.Query("SELECT id, room_number FROM rooms ORDER BY room_number")
	if err != nil {
		return fmt.Errorf("seed meters: %w", err)
	}
	defer rows.Close()

	var rooms []roomInfo
	for rows.Next() {
		var r roomInfo
		if err := rows.Scan(&r.id, &r.number); err != nil {
			return err
		}
		rooms = append(rooms, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	meterTypes := []string{"Electric", "Inverter"}
	for _, r := range rooms {
		for _, mt := range meterTypes {
			var count int
			DB.QueryRow("SELECT COUNT(*) FROM meters WHERE room_id = $1 AND meter_type = $2", r.id, mt).Scan(&count)
			if count == 0 {
				id := fmt.Sprintf("MTR-%s-%s", r.number, mt)
				num := fmt.Sprintf("%s-%s", r.number, mt)
				_, err := DB.Exec(`INSERT INTO meters (id, room_id, meter_type, meter_number, initial_reading, current_reading, is_active)
					VALUES ($1, $2, $3, $4, 0, 0, true)`, id, r.id, mt, num)
				if err != nil {
					return fmt.Errorf("seed meter %s/%s: %w", r.number, mt, err)
				}
				log.Printf("  ✓ Meter: Room %s — %s", r.number, mt)
			}
		}
	}

	// Ensure BUILDING room exists for utility meters (FK constraint)
	DB.Exec(`INSERT INTO rooms (id, room_number, floor, type, price) VALUES ('BUILDING', 'BLDG', 0, 'Utility', 0) ON CONFLICT (id) DO NOTHING`)

	// Building-level water meter
	var bcount int
	DB.QueryRow("SELECT COUNT(*) FROM meters WHERE room_id = 'BUILDING' AND meter_type = 'Water'").Scan(&bcount)
	if bcount == 0 {
		_, err := DB.Exec(`INSERT INTO meters (id, room_id, meter_type, meter_number, initial_reading, current_reading, is_active)
			VALUES ('MTR-BUILDING-Water', 'BUILDING', 'Water', 'BLDG-Water', 0, 0, true)`)
		if err != nil {
			return fmt.Errorf("seed building water meter: %w", err)
		}
		log.Println("  ✓ Building Water Meter created")
	}

	return nil
}

// SeedBlogPosts inserts initial blog posts (idempotent — skips if slugs already exist).
func SeedBlogPosts() error {
	type post struct {
		id, title, slug, excerpt, content, imageURL, author, status string
	}

	posts := []post{
		{
			id:      "blog_seed_1",
			title:   "Best Rooms Near LPU University (2026): Affordable PG & Hostel Alternative Guide",
			slug:    "best-rooms-near-lpu-university-2026",
			excerpt: "Looking for rooms near LPU University? Discover why Vats Apartment is the #1 PG choice for LPU students. Fully furnished rooms starting ₹9,000/mo with WiFi, AC, and 24/7 security — just 10 mins from campus.",
			content: `<h2>Why Students Are Choosing Rooms Near LPU Over Hostels</h2>
<p>If you're an LPU student searching for a <strong>room near LPU University</strong>, you've probably already heard the hostel horror stories — cramped rooms, strict curfews, shared bathrooms, and zero privacy. It's no surprise that more and more LPU students are making the switch to private PG accommodations near campus.</p>
<p>But with so many options, how do you find the <strong>best room near LPU</strong> that's actually worth your money?</p>
<p>We've put together this complete guide to help you make the right choice — and show you why <strong>Vats Apartment</strong> has become the top-rated PG accommodation for LPU students in 2026.</p>
<h2>What to Look for in a Room Near LPU</h2>
<p>Before you sign any rental agreement, here are the <strong>7 must-check factors</strong> when hunting for rooms near LPU University:</p>
<h3>1. Distance from LPU Campus</h3>
<p>This is non-negotiable. Anything more than 15 minutes from campus adds up fast — especially during exam season when every minute counts. <strong>Vats Apartment is located just 10 minutes from LPU</strong>, near Apna Chai Wala, making it one of the most convenient locations for daily commuters.</p>
<h3>2. Fully Furnished or Empty Room?</h3>
<p>Many "budget" rooms near LPU come completely empty — you'll need to buy a bed, table, chair, almirah, and sometimes even a fan. A <strong>fully furnished room</strong> saves you ₹15,000-25,000 in upfront setup costs. Vats Apartment provides everything: <strong>double bed, almirah, table & chair, smart TV, fridge, and AC</strong> — just bring your suitcase.</p>
<h3>3. WiFi Quality</h3>
<p>As an LPU student, fast internet isn't optional — it's essential for online classes, assignments, and late-night study sessions. Vats Apartment offers <strong>high-speed fiber WiFi included in the rent</strong>, with coverage throughout the building.</p>
<h3>4. Security & CCTV</h3>
<p>Safety should be your top priority, especially if you're living away from home for the first time. The building has <strong>24/7 CCTV surveillance</strong> and secure entry. Parents of our current tenants consistently rate safety as their #1 reason for choosing us over other PGs near LPU.</p>
<h3>5. Monthly Rent & Hidden Costs</h3>
<p>Transparent pricing matters. Our rooms start at <strong>₹9,000/month</strong> with no hidden charges. That's inclusive of WiFi, maintenance, and security. Plus, <strong>single tenants get ₹500 OFF every month</strong> after the first month!</p>
<h3>6. Kitchen & Bathroom Quality</h3>
<p>Shared bathrooms with 10 other people? No thanks. Every room at Vats Apartment comes with a <strong>private, modern kitchen</strong> (with fridge and built-in storage) and a <strong>full private bathroom</strong> with western toilet, shower, and hot water geyser.</p>
<h3>7. Roommate Flexibility</h3>
<p>The rooms are spacious enough to share with a roommate, <strong>splitting the ₹9,000 rent</strong> down to just ₹4,500 per person — cheaper than most LPU hostels!</p>
<h2>Vats Apartment vs LPU Hostel: Honest Comparison</h2>
<table style="width:100%; border-collapse: collapse; margin: 1.5rem 0; font-size: 0.9rem;">
<thead><tr style="background: #f3f4f3;"><th style="padding: 12px; text-align: left; border: 1px solid #ddd;">Feature</th><th style="padding: 12px; text-align: left; border: 1px solid #ddd;">LPU Hostel</th><th style="padding: 12px; text-align: left; border: 1px solid #ddd; background: #e8f5e9;">Vats Apartment</th></tr></thead>
<tbody>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>Monthly Cost</strong></td><td style="padding: 10px; border: 1px solid #ddd;">₹8,000 - ₹12,000</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>₹9,000 (split = ₹4,500)</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>Privacy</strong></td><td style="padding: 10px; border: 1px solid #ddd;">Shared room (2-4 people)</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>Private room</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>Bathroom</strong></td><td style="padding: 10px; border: 1px solid #ddd;">Shared (floor-wise)</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>Private attached bathroom</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>Kitchen</strong></td><td style="padding: 10px; border: 1px solid #ddd;">Mess food only</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>Private kitchen with fridge</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>Curfew</strong></td><td style="padding: 10px; border: 1px solid #ddd;">Yes (strict timings)</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>No curfew</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>WiFi</strong></td><td style="padding: 10px; border: 1px solid #ddd;">Campus WiFi (slow)</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>High-speed fiber (included)</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>Entertainment</strong></td><td style="padding: 10px; border: 1px solid #ddd;">None</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>Smart TV included</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;"><strong>Security</strong></td><td style="padding: 10px; border: 1px solid #ddd;">Campus security</td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>24/7 CCTV + secure building</strong></td></tr>
</tbody></table>
<h2>What's Included in Your Room</h2>
<p>When you book a room at Vats Apartment, here's exactly what you get:</p>
<ul><li>✅ <strong>Double bed</strong> with premium mattress</li><li>✅ <strong>Spacious almirah</strong> for all your belongings</li><li>✅ <strong>Study table & chair</strong></li><li>✅ <strong>Smart TV</strong> with wall mount</li><li>✅ <strong>Refrigerator</strong> in the kitchen</li><li>✅ <strong>Air conditioner</strong> (select rooms)</li><li>✅ <strong>High-speed WiFi</strong> — unlimited</li><li>✅ <strong>Private bathroom</strong> with geyser</li><li>✅ <strong>Modern modular kitchen</strong></li><li>✅ <strong>24/7 security cameras</strong></li></ul>
<h2>Location: Everything Within Walking Distance</h2>
<ul><li>🏫 <strong>LPU Campus</strong> — 10 minutes walk</li><li>🚗 <strong>Auto Stand</strong> — 200 meters</li><li>🛒 <strong>Grocery stores</strong> — 200 meters</li><li>🏋️ <strong>Gym</strong> — 200 meters</li><li>☕ <strong>Apna Chai Wala</strong> — right next door!</li><li>🍽️ <strong>Restaurants & dhabas</strong> — within 5 minutes</li></ul>
<h2>How to Book Your Room Near LPU</h2>
<ol><li><strong>Check Availability</strong> — Visit our <a href="/book-now">booking page</a> to see available rooms.</li><li><strong>Schedule a Visit</strong> — Come see the room in person. Available Monday-Saturday.</li><li><strong>Complete Booking</strong> — Pay a small booking amount of just ₹2,000 to confirm.</li><li><strong>Move In</strong> — Bring your suitcase and move into your new home!</li></ol>
<div style="background: #f3f4f3; border: 1px solid #ddd; border-radius: 12px; padding: 1.5rem; margin: 2rem 0; text-align: center;">
<h3 style="margin-top: 0;">⚠️ Only 2 Rooms Left for This Semester!</h3>
<p style="margin-bottom: 0.5rem;">Rooms are filling fast — don't wait until the last minute.</p>
<a href="/book-now" style="display: inline-block; background: #111; color: white; padding: 12px 32px; border-radius: 8px; text-decoration: none; font-weight: 600; margin-top: 1rem;">Check Availability Now →</a>
</div>
<h2>Frequently Asked Questions</h2>
<h3>Q: How far is Vats Apartment from LPU campus?</h3>
<p><strong>A:</strong> Just 10 minutes by walk. Located near Apna Chai Wala, it's one of the closest premium PG accommodations to LPU.</p>
<h3>Q: What is the monthly rent?</h3>
<p><strong>A:</strong> Fully furnished rooms start at <strong>₹9,000/month</strong>. Split with a roommate = ₹4,500/month each.</p>
<h3>Q: Is it better than the LPU hostel?</h3>
<p><strong>A:</strong> Most tenants are former LPU hostel students who switched for more privacy, better amenities, and no curfews.</p>
<h3>Q: Can I share the room with a friend?</h3>
<p><strong>A:</strong> Absolutely! Rooms are spacious and designed for single or double occupancy.</p>
<h2>The Bottom Line</h2>
<p>Finding the <strong>right room near LPU University</strong> doesn't have to be stressful. With Vats Apartment, you get a premium, fully furnished room with all modern amenities — at a price that's actually affordable for students.</p>
<p><a href="/book-now" style="font-weight: 600; text-decoration: underline;">→ Check Available Rooms Now</a></p>
<p style="margin-top: 2rem; font-size: 0.85rem; color: #666;">📍 Near Apna Chai Wala, LPU, Jalandhar, Punjab | 📞 <a href="tel:+919992937447">+91 99929 37447</a> | 💬 <a href="https://wa.me/919992937447">WhatsApp Us</a></p>`,
			imageURL: "",
			author:   "Vats Apartment",
			status:   "published",
		},
		{
			id:      "blog_seed_2",
			title:   "PG Near LPU University: Complete Student Housing Guide 2026",
			slug:    "pg-near-lpu-university-student-housing-guide-2026",
			excerpt: "Searching for a PG near LPU University? Compare prices, amenities, and locations. Learn why fully furnished PGs with private rooms are the smartest choice for LPU students.",
			content: `<h2>Why a Good PG Near LPU Matters</h2>
<p>Choosing the right <strong>PG near LPU University</strong> can make or break your college experience. A bad PG means sleepless nights, terrible food, and constant stress. A good one gives you the peace of mind to focus on what really matters — your studies and your future.</p>
<p>With hundreds of PGs and hostels around LPU, how do you pick the right one? This guide covers everything you need to know before making your decision.</p>
<h2>Types of PG Accommodations Near LPU</h2>
<h3>1. Shared Room PGs (₹4,000 - ₹6,000/month)</h3>
<p>The cheapest option. You share a room with 2-4 other students. Expect shared bathrooms, basic furniture, and often unreliable WiFi. Good for students on a tight budget, but be prepared to compromise on privacy and comfort.</p>
<h3>2. Single Room PGs (₹6,000 - ₹8,000/month)</h3>
<p>Your own room, but shared common areas. Better privacy than shared rooms, but you'll still share bathrooms and kitchen space with other tenants. Quality varies widely — some are well-maintained, others are not.</p>
<h3>3. Premium Fully Furnished PGs (₹9,000 - ₹12,000/month)</h3>
<p>This is where <strong>Vats Apartment</strong> sits. You get a <strong>private room, private bathroom, and private kitchen</strong> — essentially a studio apartment. Fully furnished with modern furniture, high-speed WiFi, AC, Smart TV, fridge, and 24/7 security. The higher rent is offset by better amenities and zero compromise on quality of life.</p>
<h2>What Makes Vats Apartment Different</h2>
<p>Unlike traditional PGs that cut corners, Vats Apartment was built from the ground up as a <strong>premium student accommodation</strong>. Here's what sets it apart:</p>
<ul>
<li>🏠 <strong>Private everything</strong> — your own room, bathroom, and kitchen. No sharing with strangers.</li>
<li>🛋️ <strong>Actually furnished</strong> — double bed, almirah, study table, Smart TV, fridge, AC. Just bring your clothes.</li>
<li>🔒 <strong>Serious about security</strong> — 24/7 CCTV, secure entry, and a landlord who lives nearby.</li>
<li>📍 <strong>Prime location</strong> — 10 minutes walk to LPU campus, 200m to auto stand and grocery stores.</li>
<li>💸 <strong>No hidden costs</strong> — WiFi, maintenance, and security are all included. Single tenants get ₹500 OFF every month.</li>
</ul>
<h2>Cost Comparison: PG vs Hostel vs Apartment</h2>
<table style="width:100%; border-collapse: collapse; margin: 1.5rem 0; font-size: 0.9rem;">
<thead><tr style="background: #f3f4f3;"><th style="padding: 12px; text-align: left; border: 1px solid #ddd;">Option</th><th style="padding: 12px; text-align: left; border: 1px solid #ddd;">Monthly Cost</th><th style="padding: 12px; text-align: left; border: 1px solid #ddd;">Privacy</th><th style="padding: 12px; text-align: left; border: 1px solid #ddd;">Best For</th></tr></thead>
<tbody>
<tr><td style="padding: 10px; border: 1px solid #ddd;">LPU Hostel</td><td style="padding: 10px; border: 1px solid #ddd;">₹8,000-12,000</td><td style="padding: 10px; border: 1px solid #ddd;">Low (shared)</td><td style="padding: 10px; border: 1px solid #ddd;">Campus life enthusiasts</td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;">Budget PG</td><td style="padding: 10px; border: 1px solid #ddd;">₹4,000-6,000</td><td style="padding: 10px; border: 1px solid #ddd;">Low (shared)</td><td style="padding: 10px; border: 1px solid #ddd;">Tight budget</td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;">Standard PG</td><td style="padding: 10px; border: 1px solid #ddd;">₹6,000-8,000</td><td style="padding: 10px; border: 1px solid #ddd;">Medium</td><td style="padding: 10px; border: 1px solid #ddd;">Balance seekers</td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>Vats Apartment</strong></td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>₹9,000</strong></td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>Full privacy</strong></td><td style="padding: 10px; border: 1px solid #ddd; background: #e8f5e9;"><strong>Quality-focused students</strong></td></tr>
<tr><td style="padding: 10px; border: 1px solid #ddd;">Rental Apartment</td><td style="padding: 10px; border: 1px solid #ddd;">₹10,000-15,000+</td><td style="padding: 10px; border: 1px solid #ddd;">High</td><td style="padding: 10px; border: 1px solid #ddd;">Long-term residents</td></tr>
</tbody></table>
<h2>Red Flags to Watch Out For</h2>
<p>When visiting PGs near LPU, keep an eye out for these warning signs:</p>
<ul>
<li>🚩 <strong>No proper rental agreement</strong> — always insist on written documentation.</li>
<li>🚩 <strong>Dirty bathrooms</strong> — if they can't keep common areas clean during a visit, it won't get better.</li>
<li>🚩 <strong>Hidden charges</strong> — some PGs advertise low rent then add "maintenance," "electricity," and "service" fees.</li>
<li>🚩 <strong>Poor WiFi</strong> — test the internet speed during your visit. It's critical for online classes.</li>
<li>🚩 <strong>No security</strong> — if there's no CCTV or secure entry, look elsewhere.</li>
</ul>
<h2>Why Location Matters More Than You Think</h2>
<p>A PG that's 30 minutes from campus means 1 hour of daily commuting — that's 30 hours per month you could spend studying, sleeping, or socializing. <strong>Vats Apartment's 10-minute walk to LPU</strong> saves you precious time every single day.</p>
<p>Plus, being near grocery stores, restaurants, and the auto stand means you're never stranded. Late-night study sessions? Group projects? Weekend outings? Everything is accessible.</p>
<h2>Ready to Find Your PG Near LPU?</h2>
<p>Stop scrolling through endless listings. <strong>Visit Vats Apartment today</strong> and see for yourself why LPU students consistently choose us over hostels and budget PGs.</p>
<p><a href="/book-now" style="font-weight: 600; text-decoration: underline;">→ Check Available Rooms Now</a></p>
<p style="margin-top: 2rem; font-size: 0.85rem; color: #666;">📍 Near Apna Chai Wala, LPU, Jalandhar, Punjab | 📞 <a href="tel:+919992937447">+91 99929 37447</a></p>`,
			imageURL: "",
			author:   "Vats Apartment",
			status:   "published",
		},
		{
			id:      "blog_seed_3",
			title:   "Student Living Near LPU: Tips for First-Time Renters in Jalandhar",
			slug:    "student-living-near-lpu-tips-first-time-renters-jalandhar",
			excerpt: "Moving to Jalandhar for LPU? Learn essential tips for first-time renters — from budgeting and documentation to finding safe, comfortable accommodation near campus.",
			content: `<h2>Moving to Jalandhar for LPU? Here's What You Need to Know</h2>
<p>Congratulations on getting into LPU! Now comes the next big decision — <strong>where to live</strong>. If you're moving to Jalandhar for the first time, finding the right accommodation can feel overwhelming. This guide will walk you through everything you need to know as a first-time renter.</p>
<h2>Step 1: Set Your Budget</h2>
<p>Before you start looking at rooms, determine what you can realistically afford. Here's a practical breakdown of monthly expenses for an LPU student:</p>
<ul>
<li><strong>Rent:</strong> ₹4,500 - ₹9,000 (depending on sharing and amenities)</li>
<li><strong>Food:</strong> ₹3,000 - ₹5,000 (self-cooking saves money)</li>
<li><strong>Transport:</strong> ₹500 - ₹1,500 (minimal if you live close to campus)</li>
<li><strong>Utilities & WiFi:</strong> ₹0 - ₹1,500 (often included in premium PGs)</li>
<li><strong>Miscellaneous:</strong> ₹2,000 - ₹3,000 (laundry, toiletries, entertainment)</li>
</ul>
<p><strong>Total:</strong> Approximately ₹10,000 - ₹20,000 per month. At Vats Apartment, with rent starting at ₹9,000 (which includes WiFi, maintenance, and security), you can keep your total monthly costs under ₹15,000 — especially if you split the room with a roommate.</p>
<h2>Step 2: Choose the Right Location</h2>
<p>The area around LPU has several neighborhoods popular with students. Here's how they compare:</p>
<ul>
<li><strong>Near Apna Chai Wala (Vats Apartment area):</strong> Closest to LPU (10 min walk), grocery stores and auto stand within 200m. The most convenient location for daily commuters.</li>
<li><strong>Lawgate:</strong> 15-20 minutes from LPU, slightly cheaper but farther. Good options if you don't mind the extra commute.</li>
<li><strong>Phagwara Road:</strong> 20-25 minutes from LPU. More residential, fewer student-focused amenities.</li>
</ul>
<h2>Step 3: Know What Documents You'll Need</h2>
<p>When renting a room or PG near LPU, keep these documents ready:</p>
<ul>
<li>✅ Government ID (Aadhaar card, driver's license, or passport)</li>
<li>✅ College ID or admission letter from LPU</li>
<li>✅ Passport-size photographs (2-3 copies)</li>
<li>✅ Parent/guardian contact information</li>
<li>✅ Security deposit amount (varies by property)</li>
</ul>
<h2>Step 4: Inspect Before You Pay</h2>
<p>Never book a room based on photos alone. Always visit in person and check:</p>
<ul>
<li>🔍 <strong>Water pressure and hot water</strong> — turn on the shower and geyser.</li>
<li>🔍 <strong>Electrical points</strong> — are there enough for your laptop, phone charger, and appliances?</li>
<li>🔍 <strong>Phone signal and WiFi speed</strong> — run a quick speed test on your phone.</li>
<li>🔍 <strong>Ventilation and natural light</strong> — rooms without windows get depressing fast.</li>
<li>🔍 <strong>Lock quality</strong> — your room door should have a proper lock that you control.</li>
<li>🔍 <strong>Cleanliness</strong> — check corners, bathroom tiles, and kitchen counters.</li>
</ul>
<h2>Step 5: Understand Your Rental Agreement</h2>
<p>A proper rental agreement protects both you and the landlord. Make sure it includes:</p>
<ul>
<li>📝 Monthly rent amount and due date</li>
<li>📝 Security deposit amount and refund conditions</li>
<li>📝 Notice period (usually 1 month)</li>
<li>📝 What's included (WiFi, maintenance, electricity, water)</li>
<li>📝 Rules about guests, overnight visitors, and quiet hours</li>
<li>📝 Inventory of provided furniture and appliances</li>
</ul>
<h2>Why Vats Apartment Is Ideal for First-Time Renters</h2>
<p>If you're new to renting, a professionally managed PG like Vats Apartment takes the stress out of the process:</p>
<ul>
<li>🏠 <strong>Everything is set up</strong> — furniture, WiFi, kitchen, bathroom. You just bring your suitcase.</li>
<li>📋 <strong>Clear, transparent agreement</strong> — no hidden fees, no surprise charges.</li>
<li>🔒 <strong>Safe and secure</strong> — 24/7 CCTV gives both you and your parents peace of mind.</li>
<li>👥 <strong>Community of students</strong> — your neighbors are LPU students too, so you'll fit right in.</li>
<li>📍 <strong>Can't beat the location</strong> — 10 minutes to campus means more sleep and less stress.</li>
</ul>
<h2>Final Tips for LPU Freshers</h2>
<ol>
<li><strong>Don't wait until the last minute.</strong> The best rooms near LPU fill up weeks before the semester starts. Book early to secure your spot.</li>
<li><strong>Talk to current tenants.</strong> When you visit, ask existing residents about their experience. Honest feedback is invaluable.</li>
<li><strong>Consider a roommate.</strong> Sharing a room at Vats Apartment cuts your rent in half — from ₹9,000 to ₹4,500 per person.</li>
<li><strong>Keep emergency contacts handy.</strong> Save the landlord's number, local police station, and nearest hospital in your phone.</li>
<li><strong>Explore the neighborhood.</strong> Know where the nearest grocery store, pharmacy, and auto stand are before you need them.</li>
</ol>
<p>Moving to a new city is a big step — but with the right accommodation, it can also be the start of an incredible chapter in your life. <strong>Welcome to Jalandhar, and welcome to LPU!</strong></p>
<p><a href="/book-now" style="font-weight: 600; text-decoration: underline;">→ Check Available Rooms at Vats Apartment</a></p>
<p style="margin-top: 2rem; font-size: 0.85rem; color: #666;">📍 Near Apna Chai Wala, LPU, Jalandhar, Punjab | 📞 <a href="tel:+919992937447">+91 99929 37447</a> | 💬 <a href="https://wa.me/919992937447">WhatsApp Us</a></p>`,
			imageURL: "",
			author:   "Vats Apartment",
			status:   "published",
		},
	}

	inserted := 0
	for _, p := range posts {
		var count int
		DB.QueryRow("SELECT COUNT(*) FROM blog_posts WHERE slug = $1", p.slug).Scan(&count)
		if count > 0 {
			log.Printf("Blog post already exists: %s", p.slug)
			continue
		}

		_, err := DB.Exec(
			`INSERT INTO blog_posts (id, title, slug, excerpt, content, image_url, author, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`,
			p.id, p.title, p.slug, p.excerpt, p.content, p.imageURL, p.author, p.status,
		)
		if err != nil {
			return fmt.Errorf("seed blog post %q: %w", p.slug, err)
		}
		log.Printf("  ✓ Blog post seeded: %s", p.title)
		inserted++
	}

	if inserted > 0 {
		log.Printf("Blog seed complete: %d new posts inserted", inserted)
	}
	return nil
}
