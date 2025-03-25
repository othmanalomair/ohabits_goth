package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

func GetNoteByDate(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	// If there is no params show error message
	if len(params) == 0 {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	if date, ok := params["date"]; ok {
		note, err := db.GetNoteByDate(db.DB, date, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(note)
	}

}

func PostNote(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	// Use a temporary struct to capture input as strings
	type NoteInput struct {
		Text string `json:"text"`
		Date string `json:"date"`
	}
	var input NoteInput
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
	note := db.Notes{
		Text: input.Text,
		Date: parsedDate,
	}

	if err := db.CreateNote(db.DB, note, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(note)
}

func PutNote(w http.ResponseWriter, r *http.Request) {
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

	var note db.Notes
	err = json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateNote(db.DB, note, id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(note)
}

func DeleteNote(w http.ResponseWriter, r *http.Request) {
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
	if err := db.DeleteNote(db.DB, id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Note deleted successfully"})
}
