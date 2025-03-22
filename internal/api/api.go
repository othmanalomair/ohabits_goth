package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
)

type Response struct {
	Message string `json:"message"`
}

func GetHabits(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	// If there is no params get all habits
	if len(params) == 0 {
		habits, err := db.GetAllHabits(db.DB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(Response{Message: fmt.Sprintf("Get request for %v", habits)})
	}
	// If there is a date param, get habits for that date

}
