package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

func GetWorkouts(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)

	if len(params) == 0 {
		workouts, err := db.GetAllWorkouts(db.DB, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(struct {
			Workouts []db.Workout `json:"workouts"`
		}{
			Workouts: workouts,
		})
	} else {
		workoutID, err := uuid.Parse(params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		workout, err := db.GetWorkout(db.DB, workoutID, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(struct {
			Workout db.Workout `json:"workout"`
		}{
			Workout: workout,
		})
	}

}

func PostWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	var workout db.Workout
	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.CreateWorkout(db.DB, workout, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(workout)

}

func PutWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)

	if len(params) == 0 {
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}

	workoutID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var workout db.Workout
	err = json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateWorkout(db.DB, workoutID, workout, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(workout)

}

func DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)

	if len(params) == 0 {
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}

	workoutID, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.DeleteWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{
		Message: "Workout deleted successfully",
	})

}

func GetWorkoutLog(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	dateStr, ok := params["date"]
	if !ok {
		http.Error(w, "Missing workout date", http.StatusBadRequest)
		return
	}

	logDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	log, err := db.GetWorkoutLogByDate(db.DB, logDate, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(log)
}

func PostWorkoutLog(w http.ResponseWriter, r *http.Request) {

	// Extract user ID from request context

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)

	if len(params) == 0 {
		http.Error(w, "Missing workout date", http.StatusBadRequest)
		return
	}

	workoutDateStr, ok := params["date"]
	if !ok {
		http.Error(w, "Missing workout date", http.StatusBadRequest)
		return
	}

	// Parse date string like "2025-03-22"
	workoutDate, err := time.Parse("2006-01-02", workoutDateStr)
	if err != nil {
		http.Error(w, "Invalid date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	var workoutLog db.WorkoutLog
	err = json.NewDecoder(r.Body).Decode(&workoutLog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.CreateWorkoutLog(db.DB, workoutLog, userID, workoutDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(workoutLog)

}
