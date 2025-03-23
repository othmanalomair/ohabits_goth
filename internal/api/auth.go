package api

import (
	"encoding/json"
	"net/http"

	"ohabits.com/internal/db"
)

// Register handler
func Register(w http.ResponseWriter, r *http.Request) {
	var user db.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Email == "" || user.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	createdUser, err := db.Register(r.Context(), db.DB, user.Email, user.Password, user.DisplayName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(createdUser)

}

// Login handler

func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := db.Login(r.Context(), db.DB, creds.Email, creds.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})

}
