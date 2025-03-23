package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"ohabits.com/internal/api"
	"ohabits.com/internal/util"
)

type Response struct {
	Message string `json:"message"`
}

// Handler functions

func getHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("Get request for %v", params)})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("Post reques for %v", params)})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("Delete reques for %v", params)})
}

func Server() {

	r := mux.NewRouter()

	// Login & Register (No Auth)
	r.HandleFunc("/api/register", api.Register).Methods("POST")
	r.HandleFunc("/api/login", api.Login).Methods("POST")

	// Protected Routes (Auth)
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(util.AuthMiddleware)

	// Habits (i need to fix the date in /api/habits_completed so it can accept (2025-03-22) insted of (2025-03-22T00:00:00Z) )
	protected.HandleFunc("/habits", api.GetHabits).Methods("GET")
	protected.HandleFunc("/habits", api.PostHabits).Methods("POST")
	protected.HandleFunc("/habits", api.PutHabits).Methods("PUT")
	protected.HandleFunc("/habits", api.DeleteHabit).Methods("DELETE")

	protected.HandleFunc("/habits_completed/{date}", api.GetHabitsCompleted).Methods("GET")
	protected.HandleFunc("/habits_completed", api.PostHabitCompleted).Methods("POST")
	protected.HandleFunc("/habits_completed", api.PutHabitCompleted).Methods("PUT")


	// Workout
	r.HandleFunc("/api/workout", getHandler).Methods("GET")
	r.HandleFunc("/api/wokrout/{id}", getHandler).Methods("GET")
	r.HandleFunc("/api/wokrout/{id}", postHandler).Methods("POST")
	r.HandleFunc("/api/workout/{id}", deleteHandler).Methods("DELETE")

	r.HandleFunc("/api/workout/{date}", getHandler).Methods("GET")
	r.HandleFunc("/api/workout/{date}", postHandler).Methods("POST")

	// Todo
	r.HandleFunc("/api/todo/{date}", getHandler).Methods("GET")
	r.HandleFunc("/api/todo/{date}", postHandler).Methods("POST")

	// Note
	r.HandleFunc("/api/note/{date}", getHandler).Methods("GET")
	r.HandleFunc("/api/note/{date}", postHandler).Methods("POST")
	r.HandleFunc("/api/note/{date}", deleteHandler).Methods("DELETE")

	// Rate
	r.HandleFunc("/api/rate/{date}", getHandler).Methods("GET")
	r.HandleFunc("/api/rate/{date}", postHandler).Methods("POST")

	// View mode
	r.HandleFunc("/api/view/{month}", getHandler).Methods("GET")

	// User
	r.HandleFunc("/api/profile/{id}", getHandler).Methods("GET")
	r.HandleFunc("/api/profile/{id}", postHandler).Methods("POST")

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))

}
