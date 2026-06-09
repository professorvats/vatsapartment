package main

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"vatsapartment-go/db"
	"golang.org/x/crypto/bcrypt"
)

type Room struct {
	ID         string
	RoomNumber string
	Floor      int
	Type       string
	Price      float64
	Status     string
}

type FloorGroup struct {
	Label string
	Rooms []Room
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	render(w, "home.html", nil)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleBookNow(w http.ResponseWriter, r *http.Request) {
	rooms, err := getRooms()
	if err != nil {
		http.Error(w, "Failed to load rooms", 500)
		return
	}
	grouped := groupByFloor(rooms)
	render(w, "booknow.html", map[string]interface{}{
		"Floors": grouped,
		"Rooms":  rooms,
	})
}

func handleRoomShowcase(w http.ResponseWriter, r *http.Request) {
	render(w, "showcase.html", nil)
}

func handleLocation(w http.ResponseWriter, r *http.Request) {
	render(w, "location.html", nil)
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	render(w, "contact.html", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	render(w, "login.html", nil)
}

func handlePrivacy(w http.ResponseWriter, r *http.Request) {
	render(w, "privacy.html", nil)
}

func handleTerms(w http.ResponseWriter, r *http.Request) {
	render(w, "terms.html", nil)
}

func handleSupport(w http.ResponseWriter, r *http.Request) {
	render(w, "support.html", nil)
}

func handleLoginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var hash string
	err := db.DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		render(w, "login.html", map[string]interface{}{"Error": "Invalid username or password"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func getRooms() ([]Room, error) {
	rows, err := db.DB.Query(`SELECT r.id, r.room_number, r.floor, r.type, r.price,
		COALESCE(b.status, 'available') as status
		FROM rooms r
		LEFT JOIN bookings b ON r.id = b.room_id AND b.status = 'active'
		ORDER BY r.floor, r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var r Room
		if err := rows.Scan(&r.ID, &r.RoomNumber, &r.Floor, &r.Type, &r.Price, &r.Status); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func groupByFloor(rooms []Room) []FloorGroup {
	floorMap := make(map[int][]Room)
	for _, r := range rooms {
		floorMap[r.Floor] = append(floorMap[r.Floor], r)
	}

	labels := map[int]string{1: "GROUND FLOOR", 2: "FIRST FLOOR", 3: "SECOND FLOOR", 4: "ROOFTOP"}
	order := []int{3, 2, 1, 4}
	var groups []FloorGroup
	for _, f := range order {
		if rooms, ok := floorMap[f]; ok {
			groups = append(groups, FloorGroup{Label: labels[f], Rooms: rooms})
		}
	}
	return groups
}

func handleContactPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")

	log.Printf("Contact form: name=%s email=%s", name, email)

	smtpHost := getEnv("SMTP_HOST", "smtp.gmail.com")
	smtpPort := getEnv("SMTP_PORT", "587")
	smtpUser := getEnv("SMTP_USER", "hament.mailbox@gmail.com")
	smtpPass := getEnv("SMTP_PASSWORD", "")

	body := fmt.Sprintf("From: %s <%s>\r\nSubject: Contact Form - Vats Apartment\r\n\r\nName: %s\r\nEmail: %s\r\n\r\nMessage:\r\n%s",
		name, email, name, email, message)

	if smtpPass != "" {
		auth := smtp.PlainAuth("", smtpUser, smtpPass, strings.Split(smtpHost, ":")[0])
		to := []string{smtpUser}
		msg := []byte("To: " + smtpUser + "\r\n" + body)
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		if err := smtp.SendMail(addr, auth, smtpUser, to, msg); err != nil {
			log.Printf("SMTP send error: %v", err)
		} else {
			log.Printf("Contact email sent to %s", smtpUser)
		}
	} else {
		log.Printf("SMTP not configured, contact message logged only")
	}

	render(w, "contact.html", map[string]interface{}{"Success": "Message sent successfully! We'll get back to you soon."})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
