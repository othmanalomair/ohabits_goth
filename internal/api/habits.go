package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

type Response struct {
	Message string `json:"message"`
}

func GetHabits(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	// If there is no params get all habits
	if len(params) == 0 {
		habits, err := db.GetAllHabits(db.DB, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(struct {
			Habits []db.Habit `json:"habits"`
		}{
			Habits: habits,
		})
	}
}

func PostHabits(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var habit db.Habit
	err := json.NewDecoder(r.Body).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = db.CreateHabit(db.DB, habit, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(habit)
}

func PutHabits(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var habit db.Habit
	err := json.NewDecoder(r.Body).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = db.UpdateHabit(db.DB, habit, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(habit)
}

func DeleteHabit(w http.ResponseWriter, r *http.Request) {
	var habit db.Habit
	err := json.NewDecoder(r.Body).Decode(&habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = db.DeleteHabit(db.DB, habit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("habit deleted")})
}

func GetHabitsCompleted(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	// If there is no params show error message
	if len(params) == 0 {
		json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("date is required")})
	}
	// If there is a date param, get habits for that date
	if date, ok := params["date"]; ok {
		habits, err := db.GetHabitsCompletedByDate(db.DB, date, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(struct {
			Habits []db.HabitCompletion `json:"habits"`
		}{
			Habits: habits,
		})
	}
}

func PostHabitCompleted(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Use a temporary struct to capture input as string
	type temp struct {
		HabitID uuid.UUID `json:"habit_id"`
		Date    string    `json:"date"`
	}

	var tempDate temp
	err := json.NewDecoder(r.Body).Decode(&tempDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("2006-01-02", tempDate.Date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	habitCompletion := db.HabitCompletion{
		HabitID: tempDate.HabitID,
		Date:    parsedDate,
	}

	err = db.CreateHabitCompletion(db.DB, habitCompletion, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(habitCompletion)
}

func PutHabitCompleted(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var habitCompletion db.HabitCompletion
	err := json.NewDecoder(r.Body).Decode(&habitCompletion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = db.UpdateHabitCompletion(db.DB, habitCompletion, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(habitCompletion)
}
