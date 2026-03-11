package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/raworiginal/goNotes/internal/config"
	repo "github.com/raworiginal/goNotes/internal/repository"
	"github.com/raworiginal/goNotes/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db  *sql.DB
	cfg *config.Config
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
}

func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:  db,
		cfg: cfg,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"`, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error": "username and password required"`, http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := repo.CreateUser(h.db, req.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, `{"error": "username taken"}`, http.StatusBadRequest)
		return
	}

	// Return User
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login a user
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Parse json
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"`, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error": "username and password required"`, http.StatusBadRequest)
		return
	}

	// Get user
	user, err := repo.FindByUsername(h.db, req.Username)
	if err != nil {
		http.Error(w, `{"error": "username not found"}`, http.StatusBadRequest)
		return
	}

	// Compare hashedPassword
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error": "invalid credentials"`, http.StatusBadRequest)
		return
	}

	// Generate access token
	var res LoginResponse
	res.AccessToken, err = token.SignAccessToken(user.ID, h.cfg.JWTSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	res.RefreshToken, err = token.GenerateRefreshToken()
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Store refresh token in database
	refreshExpiry := time.Now().Add(h.cfg.RefreshTokenTTL)
	if err := repo.CreateToken(h.db, user.ID, res.RefreshToken, refreshExpiry); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Return User
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Parse json
	var req RefreshToken
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"`, http.StatusBadRequest)
		return
	}
	// Look up token with repo.FindUserIDByToken
	userID, err := repo.FindUserIDByToken(h.db, req.RefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	// Generate new access token for that user
	var res AccessToken
	res.AccessToken, err = token.SignAccessToken(userID, h.cfg.JWTSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	// Return 200 with new access token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshToken
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if err := repo.DeleteToken(h.db, req.RefreshToken); err != nil {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
