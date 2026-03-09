package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/raworiginal/goNotes/internal/config"
)

func main() {
	cfg := config.Load()
	r := chi.NewRouter()

	r.Get("/health", getHealth)

	fmt.Printf("Server running on  http://localhost:%s \n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		panic(err)
	}
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
