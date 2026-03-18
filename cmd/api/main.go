package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/raworiginal/goNotes/internal/config"
	"github.com/raworiginal/goNotes/internal/db"
	"github.com/raworiginal/goNotes/internal/handler"
	am "github.com/raworiginal/goNotes/internal/middleware"
	repo "github.com/raworiginal/goNotes/internal/repository"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	store := repo.NewStore(database)

	authHandler := handler.NewAuthHandler(store, cfg)
	noteHandler := handler.NewNotesHandler(store, cfg)
	authMiddleware := am.NewAuthMiddleware(cfg)

	r := chi.NewRouter()

	// CORS middleware - loaded from config
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   cfg.CORSMethods,
		AllowedHeaders:   cfg.CORSHeaders,
		AllowCredentials: cfg.CORSCredentials,
		MaxAge:           cfg.CORSMaxAge,
	}))

	r.Use(mw.Logger)
	r.Get("/health", getHealth)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Post("/refresh", authHandler.RefreshToken)
	})

	r.Route("/notes", func(r chi.Router) {
		r.Use(authMiddleware.Handler)
		r.Get("/", noteHandler.List)
		r.Post("/", noteHandler.Create)
		r.Get("/{id}", noteHandler.GetNoteByID)
		r.Put("/{id}", noteHandler.UpdateNote)
		r.Patch("/{id}", noteHandler.PatchNote)
		r.Delete("/{id}", noteHandler.DeleteNote)
	})

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
