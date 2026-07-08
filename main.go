package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"vatsapartment-go/db"
)

var tmpl *template.Template
var startTime = time.Now()

func main() {
	// Init DB
	if err := db.Init(); err != nil {
		log.Fatal("DB init failed:", err)
	}
	db.SeedRooms()
	db.SeedAdmin()
	db.SeedMeters()
	db.SeedBlogPosts()

	// Parse templates
	var err error
	tmpl = template.New("").Funcs(template.FuncMap{
		"icon":     icon,
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"sub":      func(a, b int) int { return a - b },
		"urlquery": url.QueryEscape,
		"jsStr": func(s string) string {
			s = strings.ReplaceAll(s, `\`, `\\`)
			s = strings.ReplaceAll(s, `'`, `\'`)
			s = strings.ReplaceAll(s, "\n", `\n`)
			return s
		},
		"mul": func(a, b interface{}) float64 {
			toF := func(v interface{}) float64 {
				switch x := v.(type) {
				case int:
					return float64(x)
				case float64:
					return x
				}
				return 0
			}
			return toF(a) * toF(b)
		},
		"add": func(a, b interface{}) float64 {
			toF := func(v interface{}) float64 {
				switch x := v.(type) {
				case int:
					return float64(x)
				case float64:
					return x
				}
				return 0
			}
			return toF(a) + toF(b)
		},
	})
	tmpl, err = tmpl.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Template parse failed:", err)
	}

	mux := http.NewServeMux()

	// Static files with caching (1 year for hashed/versioned assets)
	mux.Handle("GET /static/", http.StripPrefix("/static/", withCache(http.FileServer(http.Dir("static")), 8760*time.Hour)))

	// Public root files (favicon, robots, etc)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	// Pages
	mux.HandleFunc("GET /", handleHome)
	mux.HandleFunc("GET /book-now", handleBookNow)
	mux.HandleFunc("GET /room-showcase", handleRoomShowcase)
	mux.HandleFunc("GET /location", handleLocation)
	mux.HandleFunc("GET /contact-us", handleContact)
	mux.HandleFunc("GET /privacy", handlePrivacy)
	mux.HandleFunc("GET /terms", handleTerms)
	mux.HandleFunc("GET /support", handleSupport)
	mux.HandleFunc("GET /blog", handleBlog)
	mux.HandleFunc("GET /blog/post", handleBlogPost)
	mux.HandleFunc("GET /login", handleLogin)
	mux.HandleFunc("POST /login", handleLoginPost)
	mux.HandleFunc("GET /logout", handleLogout)

	// Admin
	mux.HandleFunc("GET /admin/dashboard", handleAdminDashboard)
	mux.HandleFunc("GET /admin/rooms", handleAdminRooms)
	mux.HandleFunc("POST /admin/rooms/save", handleAdminRoomsSave)
	mux.HandleFunc("POST /admin/rooms/add", handleAdminRoomAdd)
	mux.HandleFunc("GET /admin/tenants", handleAdminTenants)
	mux.HandleFunc("POST /admin/tenants/save", handleAdminTenantsSave)
	mux.HandleFunc("GET /admin/payments", handleAdminPayments)
	mux.HandleFunc("POST /admin/payments/save", handleAdminPaymentsSave)
	mux.HandleFunc("POST /admin/bills/generate", handleGenerateBills)
	mux.HandleFunc("GET /admin/meters", handleAdminMeters)
	mux.HandleFunc("POST /admin/meters/save", handleAdminMetersSave)
	mux.HandleFunc("GET /admin/meter-readings", handleAdminMeterReadings)
	mux.HandleFunc("GET /admin/blog", handleAdminBlogList)
	mux.HandleFunc("GET /admin/blog/new", handleAdminBlogForm)
	mux.HandleFunc("POST /admin/blog/save", handleAdminBlogSave)
	mux.HandleFunc("GET /admin/api-keys", handleAdminAPIKeys)
	mux.HandleFunc("POST /admin/api-keys", handleAdminAPIKeys)

	// Tenant Portal
	mux.HandleFunc("GET /tenant/dashboard", handleTenantDashboard)
	mux.HandleFunc("POST /tenant/verification/upload", handleTenantVerificationUpload)
	mux.HandleFunc("POST /tenant/change-password", handleTenantSetPassword)
	mux.HandleFunc("GET /tenant/payments", handleTenantPayments)
	mux.HandleFunc("GET /tenant/logout", handleTenantLogout)

	// Static uploads
	mux.HandleFunc("GET /uploads/", handleTenantUploadedFile)

	// API
	mux.HandleFunc("GET /api/rooms", handleAPIRooms)
	mux.HandleFunc("GET /api/booking-management", handleAPIBookings)
	mux.HandleFunc("GET /api/room-assignments", handleAPIAssignments)
	mux.HandleFunc("GET /api/blog/preview", handleBlogPreview)
	mux.HandleFunc("POST /api/bookings", handleCreateBooking)
	mux.HandleFunc("POST /api/contact", handleContactPost)
	mux.HandleFunc("POST /api/blog", handleAPIBlogCreate)

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","uptime":"%s"}`, time.Since(startTime).String())
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      withSecurityHeaders(withLogging(mux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Vats Apartment running on http://localhost:%s", port)
	log.Printf("RAM: ~10MB | DB: Neon PostgreSQL | Framework: Go stdlib")
	log.Fatal(srv.ListenAndServe())
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func withCache(next http.Handler, maxAge time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%.0f, immutable", maxAge.Seconds()))
		next.ServeHTTP(w, r)
	})
}

func render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Short public cache for edges (CDN), browser revalidates
	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, must-revalidate")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template error (%s): %v", name, err)
		http.Error(w, "Render error", 500)
	}
}

func renderPrivate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// No caching for admin/tenant pages
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template error (%s): %v", name, err)
		http.Error(w, "Render error", 500)
	}
}

func jsonContent(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
