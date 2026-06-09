package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"vatsapartment-go/db"
)

func handleAPIRooms(w http.ResponseWriter, r *http.Request) {
	jsonContent(w)
	rooms, err := getRooms()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"rooms": []Room{}})
		return
	}
	if rooms == nil {
		rooms = []Room{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"rooms": rooms})
}

func handleAPIBookings(w http.ResponseWriter, r *http.Request) {
	jsonContent(w)
	rows, err := db.DB.Query("SELECT id, room_id, status FROM bookings")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"bookings": []interface{}{}})
		return
	}
	defer rows.Close()

	type Booking struct {
		ID     string `json:"id"`
		RoomID string `json:"roomId"`
		Status string `json:"status"`
	}
	var bookings []Booking
	for rows.Next() {
		var b Booking
		rows.Scan(&b.ID, &b.RoomID, &b.Status)
		bookings = append(bookings, b)
	}
	if bookings == nil {
		bookings = []Booking{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"bookings": bookings})
}

func handleAPIAssignments(w http.ResponseWriter, r *http.Request) {
	jsonContent(w)
	rows, err := db.DB.Query("SELECT id, room_id, is_active FROM room_assignments")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"assignments": []interface{}{}})
		return
	}
	defer rows.Close()

	type Assignment struct {
		ID       string `json:"id"`
		RoomID   string `json:"roomId"`
		IsActive bool   `json:"isActive"`
	}
	var assignments []Assignment
	for rows.Next() {
		var a Assignment
		var active int
		rows.Scan(&a.ID, &a.RoomID, &active)
		a.IsActive = active == 1
		assignments = append(assignments, a)
	}
	if assignments == nil {
		assignments = []Assignment{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"assignments": assignments})
}

func handleCreateBooking(w http.ResponseWriter, r *http.Request) {
	jsonContent(w)
	var body struct {
		Room     string  `json:"room"`
		RoomName string  `json:"roomName"`
		RoomType string  `json:"roomType"`
		Price    float64 `json:"price"`
		Date     string  `json:"date"`
		Name     string  `json:"name"`
		Phone    string  `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error":"invalid body"}`)
		return
	}

	id := fmt.Sprintf("BK%d", len(body.Name)*100+len(body.Phone))
	_, err := db.DB.Exec(
		`INSERT INTO bookings (id, room_id, rent_amount, check_in_date, status) VALUES ($1, $2, $3, $4, 'active')`,
		id, body.Room, body.Price, body.Date,
	)
	if err != nil {
		log.Printf("Booking insert error: %v", err)
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error":"failed to create booking"}`)
		return
	}
	fmt.Fprintf(w, `{"success":true,"id":"%s"}`, id)
}
