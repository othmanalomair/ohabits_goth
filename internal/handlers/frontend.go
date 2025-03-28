package handlers

import (
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
