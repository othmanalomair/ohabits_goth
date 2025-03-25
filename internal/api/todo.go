package api

import (
	"encoding/json"
	"net/http"
	"time"

	"ohabits.com/internal/db"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetTodos(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	dateStr, ok := params["date"]
	if !ok {
		http.Error(w, "Missing date parameter", http.StatusBadRequest)
		return
	}

	todos, err := db.GetTodosByDate(db.DB, dateStr, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(todos)
}

func PostTodo(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	// Use a temporary struct to capture input as strings
	type TodoInput struct {
		Text string `json:"text"`
		Date string `json:"date"`
	}
	var input TodoInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse the date from the provided string using the expected layout
	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Build your todo using the parsed date
	todo := db.Todos{
		Text: input.Text,
		Date: parsedDate,
		// Other fields (e.g. Completed, etc.) will be set by default or need to be added here as needed.
	}

	err = db.CreateTodo(db.DB, todo, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(todo)
}

func PutTodo(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var todo db.Todos
	err = json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateTodo(db.DB, todo, id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(todo)

}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	// Extract todo ID from the URL parameters
	params := mux.Vars(r)
	idStr, ok := params["id"]
	if !ok {
		http.Error(w, "Missing todo ID", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteTodo(db.DB, id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Todo deleted successfully"})
}
