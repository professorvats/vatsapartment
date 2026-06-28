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

// SeedMeters creates default meters for all rooms (Electric, Inverter, Water)
// plus a building-level water meter. Safe to run multiple times (idempotent).
// Also deduplicates — keeps only 1 meter per room+type (best one: has meter_number or highest reading).
func SeedMeters() error {
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

	meterTypes := []string{"Electric", "Inverter", "Water"}
	for _, r := range rooms {
		for _, mt := range meterTypes {
			var count int
			DB.QueryRow("SELECT COUNT(*) FROM meters WHERE room_id = $1 AND meter_type = $2", r.id, mt).Scan(&count)
			if count == 0 {
				id := fmt.Sprintf("MTR-%s-%s", r.number, mt)
				num := fmt.Sprintf("%s-%s", r.number, mt)
				_, err := DB.Exec(`INSERT INTO meters (id, room_id, meter_type, meter_number, initial_reading, current_reading, is_active)
					VALUES ($1, $2, $3, $4, 0, 0, 1)`, id, r.id, mt, num)
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
			VALUES ('MTR-BUILDING-Water', 'BUILDING', 'Water', 'BLDG-Water', 0, 0, 1)`)
		if err != nil {
			return fmt.Errorf("seed building water meter: %w", err)
		}
		log.Println("  ✓ Building Water Meter created")
	}

	return nil
}
