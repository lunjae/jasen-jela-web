package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5174")

	response := HealthResponse{
		Status:  "ok",
		Message: "API is running!",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Faild to encode respoonse", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/health", healthHandler)

	fmt.Println("Server listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
