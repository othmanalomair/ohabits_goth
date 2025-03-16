package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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

	// Habits ( thinking about the id and date maybe use one and leave the other )
	r.HandleFunc("/api/habits", getHandler).Methods("GET")
	r.HandleFunc("/api/habits/{id}", postHandler).Methods("POST")
	r.HandleFunc("/api/habits/{id}", deleteHandler).Methods("DELETE")

	r.HandleFunc("/api/habits/{date}", getHandler).Methods("GET")
	r.HandleFunc("/api/habits/{date}", postHandler).Methods("POST")

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))

}
