package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"desent/store"
)

type Middleware struct {
	store *store.Store
}

func New(s *store.Store) *Middleware {
	return &Middleware{store: s}
}

func (m *Middleware) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			jsonError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		if !m.store.ValidateToken(token) {
			jsonError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next(w, r)
	}
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
