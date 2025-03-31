package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"ohabits.com/internal/db"
)

var tmpl *template.Template
var DB *pgxpool.Pool // reference your DB connection

// Define helper functions.

func isActive(scheduledDays string, uiIndex int) bool {
	var days []string
	if err := json.Unmarshal([]byte(scheduledDays), &days); err == nil {
		// Use a Sunday-first order for the UI:
		uiWeekdays := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		if uiIndex < 0 || uiIndex >= len(uiWeekdays) {
			return false
		}
		expected := uiWeekdays[uiIndex]
		// Return true if the expected weekday is present anywhere in the days slice.
		for _, d := range days {
			if d == expected {
				return true
			}
		}
		return false
	}
	// Fallback: try as []bool (if stored that way)
	var boolDays []bool
	if err := json.Unmarshal([]byte(scheduledDays), &boolDays); err == nil {
		if uiIndex < 0 || uiIndex >= len(boolDays) {
			return false
		}
		return boolDays[uiIndex]
	}
	return false
}

func list(args ...string) []string {
	return args
}

func substr(s string, start, length int) string {
	runes := []rune(s)
	if start < 0 || start >= len(runes) {
		return ""
	}
	end := start + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

func init() {
	tmpl = template.New("").Funcs(template.FuncMap{
		"now": time.Now,
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"isActive": isActive,
		"list":     list,
		"substr":   substr,
	})
	var err error
	tmpl, err = tmpl.ParseGlob("templates/*.html")
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
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		selectedDate = time.Now()
	}

	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	user, err := db.GetUser(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		User         db.User
		SelectedDate time.Time
	}{
		User:         user,
		SelectedDate: selectedDate,
	}
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render the login template
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "login", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buf.Bytes())
		return
	}

	// If POST, process the login form
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		token, err := db.Login(context.Background(), db.DB, email, password)
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Store the token in a cookie (example approach)
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			// Secure: true,   // enable in production with HTTPS
			// SameSite: http.SameSiteStrictMode,
		})

		// Redirect to home or some protected page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render the signup template
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "signup", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buf.Bytes())
		return
	}

	// If POST, process the signup form
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Optionally, let user provide a displayName or keep it blank
		displayName := ""

		user, err := db.Register(context.Background(), db.DB, email, password, displayName)
		if err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}

		// Generate a token for the new user
		token, err := db.GenerateToken(user.ID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// Store the token in a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
		})

		// Redirect to home or some protected page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

type HabitDisplay struct {
	Habit     db.Habit
	Completed bool
}

// HabitsCompletedByDate renders the habits completion div

func HabitsCompletedByDate(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Use current date as string (or from query).
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	// Parse selected date.
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		selectedDate = time.Now()
	}

	completedHabits, err := db.GetHabitsCompletedByDate(db.DB, dateStr, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Habits       []db.HabitCompletion
		SelectedDate time.Time
	}{
		Habits:       completedHabits,
		SelectedDate: selectedDate,
	}

	// Render only the habits completion div partial.
	err = tmpl.ExecuteTemplate(w, "habits_completion", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HabitsByDay(w http.ResponseWriter, r *http.Request) {
	// Extract authenticated userID from context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Use the "date" query parameter, or default to today.
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		selectedDate = time.Now()
	}
	day := selectedDate.Weekday().String()

	// Retrieve habits scheduled for the selected day.
	habits, err := db.GetHabitsByDay(db.DB, userID, day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var habitDisplays []HabitDisplay
	for _, h := range habits {
		// Try to retrieve the completion record for each habit for the selected date.
		hc, err := db.GetHabitCompletionByHabitAndDate(db.DB, h.ID, userID, dateStr)
		completed := false
		if err == nil {
			completed = hc.Completed
		}
		habitDisplays = append(habitDisplays, HabitDisplay{
			Habit:     h,
			Completed: completed,
		})
	}

	data := struct {
		Day          string
		Habits       []HabitDisplay
		SelectedDate time.Time
	}{
		Day:          day,
		Habits:       habitDisplays,
		SelectedDate: selectedDate,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "habits_by_day", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

func ToggleHabitCompletion(w http.ResponseWriter, r *http.Request) {
	// Extract userID from context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Get the habit ID from the URL.
	vars := mux.Vars(r)
	habitID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get date from query or form.
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = r.FormValue("date")
	}
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Try to get the habit completion record for the selected date.
	hc, err := db.GetHabitCompletionByHabitAndDate(db.DB, habitID, userID, dateStr)
	if err != nil {
		// If not found, create a new record using the selected date.
		if err == pgx.ErrNoRows || err.Error() == "no rows in result set" {
			habit, err := db.GetHabitByID(db.DB, habitID, userID)
			var habitName string
			if err == nil {
				habitName = habit.Name
			}
			// Use the selected date (parsed) instead of time.Now()
			newDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				newDate = time.Now()
			}
			newHC := db.HabitCompletion{
				HabitID:   habitID,
				HabitName: habitName,
				UserID:    userID,
				Completed: true, // default toggled on
				Date:      newDate,
			}
			if err := db.CreateHabitCompletion(db.DB, newHC, userID); err != nil {
				http.Error(w, "Failed to create habit completion", http.StatusInternalServerError)
				return
			}
			// Instead of re-fetching from the DB, use the new record.
			hc = &newHC
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Toggle the Completed field.
		hc.Completed = !hc.Completed
		if err := db.UpdateHabitCompletion(db.DB, *hc, userID); err != nil {
			http.Error(w, "Failed to update habit completion", http.StatusInternalServerError)
			return
		}
		// Re-read to ensure we have the latest data.
		hc, err = db.GetHabitCompletionByHabitAndDate(db.DB, habitID, userID, dateStr)
		if err != nil {
			http.Error(w, "Failed to retrieve updated habit completion", http.StatusInternalServerError)
			return
		}
	}

	// Render the habit item template.
	if err := tmpl.ExecuteTemplate(w, "habit_item", hc); err != nil {
		http.Error(w, "Failed to render habit item", http.StatusInternalServerError)
		return
	}
}

func Todos(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Get date from query or form:
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = r.FormValue("date")
	}
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Parse the selected date.
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		selectedDate = time.Now()
	}

	// Fetch todos for that day.
	todos, err := db.GetTodosByDate(db.DB, dateStr, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a data structure that includes both todos and the selected date.
	data := struct {
		Todos        []db.Todos
		SelectedDate time.Time
	}{
		Todos:        todos,
		SelectedDate: selectedDate,
	}

	// Render only the todos div partial.
	err = tmpl.ExecuteTemplate(w, "todos", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ToggleTodoCompletion(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Extract the todo ID from the URL.
	vars := mux.Vars(r)
	todoIDStr := vars["id"]
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// Retrieve the current todo from the database.
	todo, err := db.GetTodoByID(db.DB, todoID, userID)
	if err != nil {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Toggle the Completed status.
	todo.Completed = !todo.Completed

	// Update the todo in the database.
	if err := db.UpdateTodo(db.DB, *todo, userID); err != nil {
		// Log the error on the server to help debugging.
		log.Printf("Error updating todo: %v", err)
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	// Set proper content type and render only the updated todo snippet.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "todos_item", todo); err != nil {
		log.Printf("Error rendering todo template: %v", err)
		http.Error(w, "Failed to render todo", http.StatusInternalServerError)
		return
	}
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Extract the todo ID from the URL.
	vars := mux.Vars(r)
	todoIDStr := vars["id"]
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// Delete the todo from the database.
	if err := db.DeleteTodo(db.DB, todoID, userID); err != nil {
		log.Printf("Error deleting todo: %v", err)
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	// Instead of a 204 No Content, return a 200 OK with an empty body.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(""))
}

func NewTodo(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Check for date in URL query and form value.
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = r.FormValue("date")
	}
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Parse the date.
	d, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		d = time.Now()
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data.
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form data: %v", err)
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}
	text := r.FormValue("text")

	// Create a new todo item.
	todo := db.Todos{
		UserID:    userID,
		Date:      d,
		Completed: false,
		Text:      text,
	}

	// Insert the new todo item into the database and get its new ID.
	newID, err := db.CreateTodo(db.DB, todo, userID)
	if err != nil {
		log.Printf("Error inserting todo: %v", err)
		http.Error(w, "Failed to insert todo", http.StatusInternalServerError)
		return
	}
	todo.ID = newID // Now the todo has a valid ID.

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "todos_item", todo); err != nil {
		log.Printf("Error rendering todo template: %v", err)
		http.Error(w, "Failed to render todo", http.StatusInternalServerError)
		return
	}
}

func Notes(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	// Get the date from the query string; default to today.
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	note, err := db.GetNoteByDate(db.DB, dateStr, userID)
	if err != nil {
		// If no note is found, we want to display an empty note instead of an error.
		// Check for pgx.ErrNoRows (or your equivalent error) and if so, use an empty note.
		if err.Error() == "no rows in result set" { // adjust according to your error handling
			note = db.Notes{
				Date: time.Now(), // or parse dateStr
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		Notes db.Notes
	}{
		Notes: note,
	}

	if err := tmpl.ExecuteTemplate(w, "notes", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SaveNote(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	// Parse the form.
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	text := r.FormValue("text")
	dateStr := r.FormValue("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	note, err := db.GetNoteByDate(db.DB, dateStr, userID)
	if err != nil {
		log.Printf("Error getting note: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if note.ID == uuid.Nil {
		// No note exists; create a new one.
		newNote := db.Notes{
			UserID: userID,
			Date:   date,
			Text:   text,
		}
		if err := db.CreateNote(db.DB, newNote, userID); err != nil {
			log.Printf("Error creating note: %v", err)
			http.Error(w, "Failed to create note", http.StatusInternalServerError)
			return
		}
		// Retrieve the newly created note.
		note, err = db.GetNoteByDate(db.DB, dateStr, userID)
		if err != nil {
			log.Printf("Error retrieving note after creation: %v", err)
			http.Error(w, "Failed to retrieve note", http.StatusInternalServerError)
			return
		}
	} else {
		// Note exists; update its text.
		note.Text = text
		if err := db.UpdateNote(db.DB, note, note.ID, userID); err != nil {
			log.Printf("Error updating note: %v", err)
			http.Error(w, "Failed to update note", http.StatusInternalServerError)
			return
		}
		// Retrieve the updated note.
		note, err = db.GetNoteByDate(db.DB, dateStr, userID)
		if err != nil {
			log.Printf("Error retrieving updated note: %v", err)
			http.Error(w, "Failed to retrieve updated note", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Notes db.Notes
	}{
		Notes: note,
	}
	if err := tmpl.ExecuteTemplate(w, "notes", data); err != nil {
		log.Printf("Error rendering note template: %v", err)
		http.Error(w, "Failed to render note", http.StatusInternalServerError)
		return
	}
}

func Mood(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	d, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		d = time.Now()
	}

	mr, err := db.GetMoodRatingByDate(db.DB, dateStr, userID)
	if err != nil {
		// Check for no rows using a direct comparison or error string.
		if err == pgx.ErrNoRows || err.Error() == "no rows in result set" {
			// No mood rating exists; use a default MoodRating.
			mr = db.MoodRating{
				UserID: userID,
				Date:   d,
				Rating: 0,
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if mr.Date.IsZero() {
		mr.Date = d
	}

	if err := tmpl.ExecuteTemplate(w, "mood_rating", mr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SaveMood(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	ratingStr := r.FormValue("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		http.Error(w, "Invalid rating", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	mr, err := db.GetMoodRatingByDate(db.DB, dateStr, userID)
	if err != nil {
		// Use error string comparison as a fallback.
		if err == pgx.ErrNoRows || err.Error() == "no rows in result set" {
			newMR := db.MoodRating{
				UserID: userID,
				Date:   date,
				Rating: rating,
			}
			if err := db.CreateRate(db.DB, newMR, userID); err != nil {
				http.Error(w, "Failed to create mood rating", http.StatusInternalServerError)
				return
			}
			mr, err = db.GetMoodRatingByDate(db.DB, dateStr, userID)
			if err != nil {
				http.Error(w, "Failed to retrieve mood rating", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Record exists; update its rating.
		mr.Rating = rating
		if err := db.UpdateRate(db.DB, mr, mr.ID, userID); err != nil {
			http.Error(w, "Failed to update mood rating", http.StatusInternalServerError)
			return
		}
		mr, err = db.GetMoodRatingByDate(db.DB, dateStr, userID)
		if err != nil {
			http.Error(w, "Failed to retrieve updated mood rating", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "mood_rating", mr); err != nil {
		http.Error(w, "Failed to render mood rating", http.StatusInternalServerError)
		return
	}
}

func extractCardioInfo(cardioJSON []byte) (string, string) {
	var arr []interface{}
	if err := json.Unmarshal(cardioJSON, &arr); err != nil {
		return "", ""
	}
	var name string
	var duration string
	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			name = s
		}
	}
	if len(arr) > 1 {
		duration = fmt.Sprintf("%v", arr[1])
	}
	return name, duration
}

func WorkoutLoging(w http.ResponseWriter, r *http.Request) {
	// Extract userID from context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Get the selected date from query, default to today.
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		selectedDate = time.Now()
	}

	workouts, err := db.GetAllWorkouts(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	workoutLog, err := db.GetWorkoutLogByDate(db.DB, dateStr, userID)
	if err != nil {
		if err == pgx.ErrNoRows || err.Error() == "no rows in result set" {
			workoutLog = db.WorkoutLog{
				UserID: userID,
				Date:   selectedDate,
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	cardioName, cardioDuration := extractCardioInfo(workoutLog.Cardio)
	data := struct {
		SelectedDate   time.Time
		Workout        []db.Workout
		WorkoutLog     db.WorkoutLog
		CardioName     string
		CardioDuration string
	}{
		SelectedDate:   selectedDate,
		Workout:        workouts,
		WorkoutLog:     workoutLog,
		CardioName:     cardioName,
		CardioDuration: cardioDuration,
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "workout_loging", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

func SaveWorkoutLog(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}
	weightStr := r.FormValue("weight")
	cardioStr := r.FormValue("cardio")
	durationStr := r.FormValue("duration")
	workoutName := r.FormValue("workout_name")
	completedExercisesStr := r.FormValue("completed_exercises")
	dateStr := r.FormValue("date")

	var weight float64
	if weightStr != "" {
		var err error
		weight, err = strconv.ParseFloat(weightStr, 64)
		if err != nil {
			http.Error(w, "Invalid weight", http.StatusBadRequest)
			return
		}
	}

	// Combine cardio and duration into valid JSON.
	var cardioJSON string
	if cardioStr != "" && durationStr != "" {
		cardioJSON = fmt.Sprintf(`["%s", %s]`, cardioStr, durationStr)
	} else {
		cardioJSON = "[]"
	}

	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
		return
	}

	logEntry := db.WorkoutLog{
		UserID:             userID,
		Name:               workoutName,
		CompletedExercises: []byte(completedExercisesStr),
		Cardio:             []byte(cardioJSON),
		Weight:             weight,
		Date:               date,
	}

	existingLog, err := db.GetWorkoutLogByDate(db.DB, dateStr, userID)
	if err != nil && !(err == pgx.ErrNoRows || err.Error() == "no rows in result set") {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err == nil {
		err = db.UpdateWorkoutLog(db.DB, existingLog.ID, logEntry, userID)
		if err != nil {
			http.Error(w, "Failed to update workout log", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.CreateWorkoutLog(db.DB, logEntry, userID, date)
		if err != nil {
			http.Error(w, "Failed to create workout log", http.StatusInternalServerError)
			return
		}
	}
	WorkoutLoging(w, r)
}

func GetWorkoutExercises(w http.ResponseWriter, r *http.Request) {
	// Extract userID from the request context.
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	workoutIDStr := r.URL.Query().Get("workout")
	// Instead of erroring out, if missing, render default content.
	if workoutIDStr == "" {
		w.Write([]byte(`<div class="exercise-item">Select a workout to see exercises</div>`))
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
	if err := tmpl.ExecuteTemplate(w, "exercise_list", workout); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func CalendarHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "calendar", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

// HabitsPage renders the habits page with a list of all habits.
func HabitsPage(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Get all habits for this user (using the Habit entity)
	habits, err := db.GetAllHabits(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Also fetch the user for the header
	user, err := db.GetUser(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Habits []db.Habit
		User   db.User
	}{
		Habits: habits,
		User:   user,
	}

	// Render the full habits page
	if err := tmpl.ExecuteTemplate(w, "habits", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddHabit handles adding a new habit (POST /habits)
func AddHabit(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("habit_name")
	if name == "" {
		http.Error(w, "Habit name required", http.StatusBadRequest)
		return
	}
	// Default scheduled days: a JSON array of 7 false values.
	defaultDays := `[false,false,false,false,false,false,false]`
	newHabit := db.Habit{
		UserID:        userID,
		Name:          name,
		ScheduledDays: []byte(defaultDays),
	}
	if err := db.CreateHabit(db.DB, newHabit, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Return updated habits list partial.
	habits, err := db.GetAllHabits(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Habits []db.Habit
	}{Habits: habits}
	tmpl.ExecuteTemplate(w, "habits_list", data)
}

// EditHabitForm renders a form to edit a habit (GET /habits/{id}/edit-form)
func EditHabitForm(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	habitIDStr := vars["id"]
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		http.Error(w, "Invalid habit ID", http.StatusBadRequest)
		return
	}
	habit, err := db.GetHabitByID(db.DB, habitID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "habit_edit", habit)
}

// EditHabit handles saving an edited habit (POST /habits/{id}/edit)
func EditHabit(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	habitIDStr := vars["id"]
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		http.Error(w, "Invalid habit ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	newName := r.FormValue("habit_name")
	if newName == "" {
		http.Error(w, "Habit name required", http.StatusBadRequest)
		return
	}
	habit, err := db.GetHabitByID(db.DB, habitID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	habit.Name = newName
	if err := db.UpdateHabit(db.DB, habit, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Return updated habit partial.
	tmpl.ExecuteTemplate(w, "habit_item_", habit)
}

// DeleteHabit handles deleting a habit (POST /habits/{id}/delete)
func DeleteHabit(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	habitIDStr := vars["id"]
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		http.Error(w, "Invalid habit ID", http.StatusBadRequest)
		return
	}
	habit, err := db.GetHabitByID(db.DB, habitID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := db.DeleteHabit(db.DB, habit); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Return updated habits list partial.
	habits, err := db.GetAllHabits(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Habits []db.Habit
	}{Habits: habits}
	tmpl.ExecuteTemplate(w, "habits_list", data)

}

// ToggleHabitDay toggles the scheduled status for a given day (POST /habits/{id}/toggle?day=X)
func ToggleHabitDay(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	habitIDStr := vars["id"]
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		http.Error(w, "Invalid habit ID", http.StatusBadRequest)
		return
	}
	dayStr := r.URL.Query().Get("day")
	if dayStr == "" {
		http.Error(w, "Missing day index", http.StatusBadRequest)
		return
	}
	dayIndex, err := strconv.Atoi(dayStr)
	if err != nil || dayIndex < 0 || dayIndex > 6 {
		http.Error(w, "Invalid day index", http.StatusBadRequest)
		return
	}
	habit, err := db.GetHabitByID(db.DB, habitID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Using weekday strings (new approach)
	var schedule []string
	if err := json.Unmarshal(habit.ScheduledDays, &schedule); err != nil {
		// Initialize with no days active, or default as needed
		schedule = []string{}
	}

	weekdayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	weekday := weekdayNames[dayIndex]
	// get the weekday name for the day index

	// Check if the weekday is already scheduled.
	found := false
	for i, d := range schedule {
		if d == weekday {
			// Remove the day from the schedule.
			schedule = append(schedule[:i], schedule[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		// Add the day to the schedule.
		schedule = append(schedule, weekday)
	}
	newSchedule, err := json.Marshal(schedule)

	if err != nil {
		http.Error(w, "Failed to marshal schedule", http.StatusInternalServerError)
		return
	}
	habit.ScheduledDays = newSchedule
	// After updating the habit in the database…
	if err := db.UpdateHabit(db.DB, habit, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated habit partial using the correct template name.
	if err := tmpl.ExecuteTemplate(w, "habit_item_", habit); err != nil {
		http.Error(w, "Failed to render habit item", http.StatusInternalServerError)
		return
	}

}

// CancelHabitEdit returns the habit item partial to cancel editing.
func CancelHabitEdit(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	habitIDStr := vars["id"]
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		http.Error(w, "Invalid habit ID", http.StatusBadRequest)
		return
	}
	habit, err := db.GetHabitByID(db.DB, habitID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Render the habit_item_ partial.
	if err := tmpl.ExecuteTemplate(w, "habit_item_", habit); err != nil {
		http.Error(w, "Failed to render habit item", http.StatusInternalServerError)
		return
	}
}
