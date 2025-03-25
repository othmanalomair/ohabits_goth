package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"ohabits.com/internal/db"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	user, err := db.GetUser(db.DB, userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)

}

func PutUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	var user db.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateUser(db.DB, user, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)

}
