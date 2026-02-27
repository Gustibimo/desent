package handler

import (
	"fmt"
	"math/rand"
	"net/http"
)

func (h *Handler) AuthToken(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &creds); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if creds.Username != "admin" || creds.Password != "password" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	token := fmt.Sprintf("%016x", rand.Int63())
	h.store.AddToken(token)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
