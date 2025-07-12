package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/image/draw"
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
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"value": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"formatCardio": func(cardio json.RawMessage) string {
			if len(cardio) == 0 {
				return "-"
			}
			var arr []interface{}
			if err := json.Unmarshal([]byte(cardio), &arr); err != nil || len(arr) < 2 {
				return "-"
			}
			name, ok1 := arr[0].(string)
			duration, ok2 := arr[1].(float64)
			if !ok1 || !ok2 {
				return "-"
			}
			return fmt.Sprintf("%s (%dmin)", name, int(duration))
		},
		"nl2br": func(text string) template.HTML {
			escaped := template.HTMLEscapeString(text)
			withBreaks := strings.ReplaceAll(escaped, "\n", "<br>")
			return template.HTML(withBreaks)
		},
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

	// Update the URL query so that WorkoutLoging uses the correct date.
	r.URL.RawQuery = "date=" + dateStr

	// Call the WorkoutLoging handler to re-render the page.
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

// WorkoutPlanPage renders the workout plans page.
func WorkoutPlanPage(w http.ResponseWriter, r *http.Request) {
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
	workouts, err := db.GetAllWorkouts(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Wrap each workout in a local struct to include an "Open" flag.
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	var plans []WorkoutPlan
	for _, wkt := range workouts {
		plans = append(plans, WorkoutPlan{Workout: wkt, Open: false})
	}
	data := struct {
		WorkoutPlans []WorkoutPlan
	}{
		WorkoutPlans: plans,
	}
	if err := tmpl.ExecuteTemplate(w, "workout_plan", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateWorkoutPlan creates a new workout plan.
func CreateWorkoutPlan(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("workout_name")
	newWorkout := db.Workout{
		UserID:    userID,
		Name:      name,
		Day:       "N/A",
		Exercises: []db.Exercise{},
	}
	if err := db.CreateWorkout(db.DB, newWorkout, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// After successful creation, fetch all workouts.
	workouts, err := db.GetAllWorkouts(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	var plans []WorkoutPlan
	for _, wkt := range workouts {
		plans = append(plans, WorkoutPlan{Workout: wkt, Open: false})
	}
	data := struct {
		WorkoutPlans []WorkoutPlan
	}{
		WorkoutPlans: plans,
	}
	// Return the container partial.
	if err := tmpl.ExecuteTemplate(w, "workout_plans_container", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ToggleWorkoutPlan toggles the open/closed state of a workout plan.
func ToggleWorkoutPlan(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Toggle open state: for simplicity, use a query parameter.
	openParam := r.URL.Query().Get("open")
	open := true
	if openParam == "false" {
		open = false
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: open}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_item", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteWorkoutPlan deletes a workout plan.
func DeleteWorkoutPlan(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	if err := db.DeleteWorkout(db.DB, workoutID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// After deletion, re-fetch workouts.
	workouts, err := db.GetAllWorkouts(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	var plans []WorkoutPlan
	for _, wkt := range workouts {
		plans = append(plans, WorkoutPlan{Workout: wkt, Open: false})
	}
	data := struct {
		WorkoutPlans []WorkoutPlan
	}{
		WorkoutPlans: plans,
	}
	// Return the container partial.
	if err := tmpl.ExecuteTemplate(w, "workout_plans_container", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CancelWorkoutPlanEdit re-renders the workout plan item (cancelling the edit mode).
func CancelWorkoutPlanEdit(w http.ResponseWriter, r *http.Request) {
	// This handler re-renders the workout plan item in non-edit mode.
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: false}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_item", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// EditWorkoutPlanForm renders the edit form for a workout plan.
func EditWorkoutPlanForm(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: true}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_edit", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// EditWorkoutPlan handles saving updates from the edit form.
func EditWorkoutPlan(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("workout_name")
	day := r.FormValue("day")
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workout.Name = name
	workout.Day = day
	if err := db.UpdateWorkout(db.DB, workoutID, workout, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: false}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_item", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// UpdateWorkoutPlanDay updates the day selection for a workout plan.
func UpdateWorkoutPlanDay(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	newDay := r.FormValue("day")
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workout.Day = newDay
	if err := db.UpdateWorkout(db.DB, workoutID, workout, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: true}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_item", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddWorkoutExercise handles adding a new exercise to a workout plan.
func AddWorkoutExercise(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	exerciseName := r.FormValue("exercise_name")
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newOrder := len(workout.Exercises) + 1
	newExercise := db.Exercise{
		Order: newOrder,
		Name:  exerciseName,
	}
	workout.Exercises = append(workout.Exercises, newExercise)
	if err := db.UpdateWorkout(db.DB, workoutID, workout, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: true}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_item", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SaveWorkoutPlan finalizes and "saves" the workout plan (here we simply close the plan view).
func SaveWorkoutPlan(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: false}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_item", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DeleteWorkoutExercise(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	orderStr := vars["order"]
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		http.Error(w, "Invalid exercise order", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newExercises := []db.Exercise{}
	for _, ex := range workout.Exercises {
		if ex.Order != order {
			newExercises = append(newExercises, ex)
		}
	}
	// Reassign orders sequentially.
	for i := range newExercises {
		newExercises[i].Order = i + 1
	}
	workout.Exercises = newExercises
	if err := db.UpdateWorkout(db.DB, workoutID, workout, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	plan := WorkoutPlan{Workout: workout, Open: true}
	if err := tmpl.ExecuteTemplate(w, "workout_plan_item", plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func EditWorkoutExerciseForm(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	orderStr := vars["order"]
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		http.Error(w, "Invalid exercise order", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var exercise db.Exercise
	found := false
	for _, ex := range workout.Exercises {
		if ex.Order == order {
			exercise = ex
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Exercise not found", http.StatusNotFound)
		return
	}
	data := struct {
		WorkoutID uuid.UUID
		Exercise  db.Exercise
	}{
		WorkoutID: workout.ID,
		Exercise:  exercise,
	}
	if err := tmpl.ExecuteTemplate(w, "workout_exercise_edit", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func EditWorkoutExercise(w http.ResponseWriter, r *http.Request) {
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	orderStr := vars["order"]
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		http.Error(w, "Invalid exercise order", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	newName := r.FormValue("exercise_name")
	if newName == "" {
		http.Error(w, "Exercise name required", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var updatedExercise db.Exercise
	updated := false
	for i, ex := range workout.Exercises {
		if ex.Order == order {
			workout.Exercises[i].Name = newName
			updatedExercise = workout.Exercises[i]
			updated = true
			break
		}
	}
	if !updated {
		http.Error(w, "Exercise not found", http.StatusNotFound)
		return
	}
	if err := db.UpdateWorkout(db.DB, workoutID, workout, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type WorkoutExerciseData struct {
		WorkoutID uuid.UUID
		Exercise  db.Exercise
	}
	data := WorkoutExerciseData{
		WorkoutID: workout.ID,
		Exercise:  updatedExercise,
	}
	if err := tmpl.ExecuteTemplate(w, "workout_exercise_item", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func CancelWorkoutExerciseEdit(w http.ResponseWriter, r *http.Request) {
	// Re-render the specific workout exercise item to cancel exercise edit.
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
	workoutIDStr := vars["id"]
	workoutID, err := uuid.Parse(workoutIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}
	orderStr := vars["order"]
	order, err := strconv.Atoi(orderStr)
	if err != nil {
		http.Error(w, "Invalid exercise order", http.StatusBadRequest)
		return
	}
	workout, err := db.GetWorkout(db.DB, workoutID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var exercise db.Exercise
	found := false
	for _, ex := range workout.Exercises {
		if ex.Order == order {
			exercise = ex
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Exercise not found", http.StatusNotFound)
		return
	}
	type WorkoutExerciseData struct {
		WorkoutID uuid.UUID
		Exercise  db.Exercise
	}
	data := WorkoutExerciseData{
		WorkoutID: workout.ID,
		Exercise:  exercise,
	}
	if err := tmpl.ExecuteTemplate(w, "workout_exercise_item", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get the month query parameter; default to current month ("YYYY-MM")
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// Retrieve the daily view for the month.
	dailyViews, err := db.GetViewByMonth(db.DB, month, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Compute previous and next month values.
	t, err := time.Parse("2006-01", month)
	if err != nil {
		t = time.Now()
	}
	prevMonth := t.AddDate(0, -1, 0).Format("2006-01")
	nextMonth := t.AddDate(0, 1, 0).Format("2006-01")
	currentMonth := t.Format("January 2006")

	data := struct {
		DailyViews   []db.DailyView
		PrevMonth    string
		NextMonth    string
		CurrentMonth string
	}{
		DailyViews:   dailyViews,
		PrevMonth:    prevMonth,
		NextMonth:    nextMonth,
		CurrentMonth: currentMonth,
	}

	if err := tmpl.ExecuteTemplate(w, "view", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NotesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get the month query parameter; default to current month ("YYYY-MM")
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// Retrieve the daily view for the month.
	dailyNotes, err := db.GetNotesByMonth(db.DB, month, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Compute previous and next month values.
	t, err := time.Parse("2006-01", month)
	if err != nil {
		t = time.Now()
	}
	prevMonth := t.AddDate(0, -1, 0).Format("2006-01")
	nextMonth := t.AddDate(0, 1, 0).Format("2006-01")
	currentMonth := t.Format("January 2006")

	data := struct {
		DailyNote    []db.DailyNote
		PrevMonth    string
		NextMonth    string
		CurrentMonth string
	}{
		DailyNote:    dailyNotes,
		PrevMonth:    prevMonth,
		NextMonth:    nextMonth,
		CurrentMonth: currentMonth,
	}

	if err := tmpl.ExecuteTemplate(w, "mnotes", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
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
		User           db.User
		SuccessMessage string
		ErrorMessage   string
	}{
		User:           user,
		SuccessMessage: "",
		ErrorMessage:   "",
	}
	if err := tmpl.ExecuteTemplate(w, "profile", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderProfileWithSuccess(w http.ResponseWriter, user db.User, message string) {
	data := struct {
		User           db.User
		SuccessMessage string
		ErrorMessage   string
	}{
		User:           user,
		SuccessMessage: message,
		ErrorMessage:   "",
	}
	if err := tmpl.ExecuteTemplate(w, "profile", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderProfileWithError(w http.ResponseWriter, user db.User, message string) {
	data := struct {
		User           db.User
		SuccessMessage string
		ErrorMessage   string
	}{
		User:           user,
		SuccessMessage: "",
		ErrorMessage:   message,
	}
	if err := tmpl.ExecuteTemplate(w, "profile", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure we accept multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil { // limit to 10MB
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
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

	// Get form values.
	email := r.FormValue("email")
	displayName := r.FormValue("display_name")
	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// Check if password change is requested
	passwordChangeRequested := currentPassword != "" || newPassword != "" || confirmPassword != ""
	
	var user db.User
	var err error
	
	if passwordChangeRequested {
		// Get user with password for verification
		user, err = db.GetUserWithPassword(db.DB, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Validate all password fields are provided
		if currentPassword == "" || newPassword == "" || confirmPassword == "" {
			renderProfileWithError(w, user, "All password fields are required")
			return
		}

		// Validate new passwords match
		if newPassword != confirmPassword {
			renderProfileWithError(w, user, "New passwords do not match")
			return
		}

		// Verify current password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
			renderProfileWithError(w, user, "Current password is incorrect")
			return
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)
	} else {
		// Get user without password for regular updates
		user, err = db.GetUser(db.DB, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	user.Email = email
	user.DisplayName = displayName

	// Process file upload if provided.
	file, handler, err := r.FormFile("profile_picture")
	if err == nil {
		defer file.Close()

		// Validate that the file is an image.
		contentType := handler.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			renderProfileWithError(w, user, "Only image files are allowed")
			return
		}

		// Decode the image.
		img, _, err := image.Decode(file)
		if err != nil {
			renderProfileWithError(w, user, "Failed to decode image")
			return
		}

		// Resize image to 140x140 using golang.org/x/image/draw.
		dst := image.NewRGBA(image.Rect(0, 0, 140, 140))
		// Using bilinear scaling.
		draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

		// Create directory: static/images/profile/<userid>/
		dirPath := filepath.Join("static", "images", "profile", userID.String())
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			http.Error(w, "Failed to create image directory", http.StatusInternalServerError)
			return
		}

		// Generate a random file name.
		randomName := uuid.New().String() + ".jpg"
		filePath := filepath.Join(dirPath, randomName)

		outFile, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to create image file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		// Save the resized image as JPEG.
		if err := jpeg.Encode(outFile, dst, &jpeg.Options{Quality: 85}); err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}

		// Update the user's AvatarURL (using a relative path).
		relativePath := "/" + filepath.ToSlash(filePath)
		user.AvatarURL = &relativePath
	}
	// else: if no file is uploaded, we simply don't change the AvatarURL.

	// Update the user in the database.
	if passwordChangeRequested {
		if err := db.UpdateUserWithPassword(db.DB, user, userID); err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
	} else {
		if err := db.UpdateUser(db.DB, user, userID); err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
	}

	// Render profile page with success message
	renderProfileWithSuccess(w, user, "Profile updated successfully!")
}

func MoveWorkoutUp(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := userIDValue.(uuid.UUID)

	// Get workout ID from URL
	vars := mux.Vars(r)
	workoutID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	// Move workout up
	if err := db.MoveWorkoutUp(db.DB, workoutID, userID); err != nil {
		http.Error(w, "Failed to move workout up", http.StatusInternalServerError)
		return
	}

	// Re-render the workout plans list
	workouts, err := db.GetAllWorkouts(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	var plans []WorkoutPlan
	for _, wkt := range workouts {
		plans = append(plans, WorkoutPlan{Workout: wkt, Open: false})
	}
	data := struct {
		WorkoutPlans []WorkoutPlan
	}{
		WorkoutPlans: plans,
	}
	tmpl.ExecuteTemplate(w, "workout_plans_list", data)
}

func MoveWorkoutDown(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := userIDValue.(uuid.UUID)

	// Get workout ID from URL
	vars := mux.Vars(r)
	workoutID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	// Move workout down
	if err := db.MoveWorkoutDown(db.DB, workoutID, userID); err != nil {
		http.Error(w, "Failed to move workout down", http.StatusInternalServerError)
		return
	}

	// Re-render the workout plans list
	workouts, err := db.GetAllWorkouts(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type WorkoutPlan struct {
		Workout db.Workout
		Open    bool
	}
	var plans []WorkoutPlan
	for _, wkt := range workouts {
		plans = append(plans, WorkoutPlan{Workout: wkt, Open: false})
	}
	data := struct {
		WorkoutPlans []WorkoutPlan
	}{
		WorkoutPlans: plans,
	}
	tmpl.ExecuteTemplate(w, "workout_plans_list", data)
}

func SignOutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session/cookie.
	// For example, if you use a cookie "session_id", set its MaxAge to -1 to delete it:
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Optionally clear other session data if needed.

	// For HTMX, you can issue a full-page redirect by setting the HX-Redirect header:
	w.Header().Set("HX-Redirect", "/login")
}
