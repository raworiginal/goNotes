package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/raworiginal/goNotes/internal/config"
	repo "github.com/raworiginal/goNotes/internal/repository"
	"github.com/raworiginal/goNotes/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	store *repo.Store
	cfg   *config.Config
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

func NewAuthHandler(store *repo.Store, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		store: store,
		cfg:   cfg,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.store.Users.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		jsonError(w, "username taken", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password required", http.StatusBadRequest)
		return
	}

	user, err := h.store.Users.FindUserByUsername(req.Username)
	if err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	var res LoginResponse
	res.AccessToken, err = token.SignAccessToken(user.ID, h.cfg.JWTSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	res.RefreshToken, err = token.GenerateRefreshToken()
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	refreshExpiry := time.Now().Add(h.cfg.RefreshTokenTTL)
	if err := h.store.Tokens.CreateToken(user.ID, res.RefreshToken, refreshExpiry); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshToken
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, err := h.store.Tokens.FindUserIDByToken(req.RefreshToken)
	if err != nil {
		if err == repo.ErrNotFound {
			jsonError(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var res AccessToken
	res.AccessToken, err = token.SignAccessToken(userID, h.cfg.JWTSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshToken
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.store.Tokens.DeleteToken(req.RefreshToken); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
