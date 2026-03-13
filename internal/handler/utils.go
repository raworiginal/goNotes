package handler

import (
	"encoding/json"
	"net/http"
)

func jsonError(w http.ResponseWriter, msg string, code int) {
	h := w.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
