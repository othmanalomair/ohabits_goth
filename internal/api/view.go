package api

import (
	"encoding/json"
	"net/http"
	"time"

	"ohabits.com/internal/db"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetView(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from request context.
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusBadRequest)
		return
	}

	// Expect the URL parameter "month", e.g., "2025-03".
	params := mux.Vars(r)
	month, ok := params["month"]
	if !ok {
		http.Error(w, "Month is missing", http.StatusBadRequest)
		return
	}

	// Validate month format.
	_, err := time.Parse("2006-01", month)
	if err != nil {
		http.Error(w, "Invalid month format. Expected YYYY-MM", http.StatusBadRequest)
		return
	}

	// Get the aggregated view for the month.
	viewData, err := db.GetViewByMonth(db.DB, month, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response JSON with the month name as key.
	t, _ := time.Parse("2006-01", month)
	monthName := t.Format("January")
	response := map[string]interface{}{
		monthName: viewData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
