package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"ohabits.com/internal/db"
)

var tmpl *template.Template

func init() {
	// Parse all templates recursively (base, index, partials)
	var err error
	tmpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
	tmpl, err = tmpl.ParseGlob("templates/partials/*.html")
	if err != nil {
		log.Fatalf("Error parsing partial templates: %v", err)
	}
}

// IndexHandler renders the full index page.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Render the base template which includes index.html content
	err := tmpl.ExecuteTemplate(w, "base.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HabitsCompletedByDate renders the habits completion div
func HabitsCompletedByDate(w http.ResponseWriter, r *http.Request) {
	// In a real scenario, extract the authenticated user's ID from context and the HabitsCompletedByDate
	// Here we use a dummy user id for demonstration.
	userID, _ := uuid.Parse("fdf545f7-e615-4709-a1ca-4e6f11cd6a76")
	date := "2025-03-25"
	completedHabits, err := db.GetHabitsCompletedByDate(db.DB, date, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Habits []db.HabitCompletion
	}{
		Habits: completedHabits,
	}

	// Render only the habits completion div partial.
	err = tmpl.ExecuteTemplate(w, "habits_completion", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Todos(w http.ResponseWriter, r *http.Request) {
	// In a real scenario, extract the authenticated user's ID from context and the Todos
	// Here we use a dummy user id for demonstration.
	userID, _ := uuid.Parse("fdf545f7-e615-4709-a1ca-4e6f11cd6a76")
	date := "2025-03-25"
	todos, err := db.GetTodosByDate(db.DB, date, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Todos []db.Todos
	}{
		Todos: todos,
	}

	// Render only the todos div partial.
	err = tmpl.ExecuteTemplate(w, "todos", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Notes(w http.ResponseWriter, r *http.Request) {
	// In a real scenario, extract the authenticated user's ID from context and the Notes
	// Here we use a dummy user id for demonstration.
	userID, _ := uuid.Parse("fdf545f7-e615-4709-a1ca-4e6f11cd6a76")
	date := "2025-03-25"
	notes, err := db.GetNoteByDate(db.DB, date, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Notes db.Notes
	}{
		Notes: notes,
	}

	// Render only the notes div partial.
	err = tmpl.ExecuteTemplate(w, "notes", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Mood(w http.ResponseWriter, r *http.Request) {
	// In a real scenario, extract the authenticated user's ID from context and the Notes
	// Here we use a dummy user id for demonstration.
	userID, _ := uuid.Parse("fdf545f7-e615-4709-a1ca-4e6f11cd6a76")
	date := "2025-03-25"
	rate, err := db.GetRateByDate(db.DB, date, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render only the notes div partial.
	err = tmpl.ExecuteTemplate(w, "mood_rating", rate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WorkoutLoging(w http.ResponseWriter, r *http.Request) {
	// In a real scenario, extract the authenticated user's ID from context and the Notes
	// Here we use a dummy user id for demonstration.
	userID, _ := uuid.Parse("fdf545f7-e615-4709-a1ca-4e6f11cd6a76")
	date := "2025-03-25"
	workout, err := db.GetAllWorkouts(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	workoutLog, err := db.GetWorkoutLogByDate(db.DB, date, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Workout    []db.Workout
		WorkoutLog db.WorkoutLog
	}{
		Workout:    workout,
		WorkoutLog: workoutLog,
	}

	// Render only the notes div partial.
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "workout_loging", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(buf.Bytes())
}

func GetWorkoutExercises(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse("fdf545f7-e615-4709-a1ca-4e6f11cd6a76")
	workoutIDStr := r.URL.Query().Get("workout")
	if workoutIDStr == "" {
		http.Error(w, "Missing workout id", http.StatusBadRequest)
		return
	}
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// data := struct {
	// 	Exercises db.Workout // Contains workout.Exercises
	// }{
	// 	Exercises: workout,
	// }

	err = tmpl.ExecuteTemplate(w, "exercise_list", workout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
