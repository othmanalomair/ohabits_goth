package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

func GetRate(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	// If there is no params show error message
	if len(params) == 0 {
		http.Error(w, "Date is missing", http.StatusBadRequest)
		return
	}

	if date, ok := params["date"]; ok {
		rate, err := db.GetRateByDate(db.DB, date, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(rate)
	}
}

func PostRate(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	// Use a temporary struct to capture input as strings
	type RateInput struct {
		Rating int    `json:"rating"`
		Date   string `json:"date"`
	}
	var input RateInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse the date from the provided string using the expected layout
	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}

	// Build note using the parsed date
	rate := db.MoodRating{
		Rating: input.Rating,
		Date:   parsedDate,
	}

	if err := db.CreateRate(db.DB, rate, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rate)
}

func PutRate(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	// Extract note ID from URL parameters
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var rate db.MoodRating
	err = json.NewDecoder(r.Body).Decode(&rate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateRate(db.DB, rate, id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rate)
}
