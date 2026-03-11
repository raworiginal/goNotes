package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/raworiginal/goNotes/internal/config"
	"github.com/raworiginal/goNotes/internal/token"
)

type AuthMiddleware struct {
	cfg *config.Config
}

type contextKey string

const UserContextKey = contextKey("user_id")

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

func UserIDFromContext(r *http.Request) (int, error) {
	userID := r.Context().Value(UserContextKey)
	if intUserID, ok := userID.(int); !ok {
		return 0, fmt.Errorf("failed to extract UserID from context")
	} else {
		return intUserID, nil
	}
}

func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing or invalid token"}`, http.StatusUnauthorized)
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, `{"error": "invalid authorization header"}`, http.StatusUnauthorized)
			return
		}
		userToken := headerParts[1]
		userID, err := token.ParseAccessToken(userToken, am.cfg.JWTSecret)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserContextKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
