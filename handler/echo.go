package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) Echo(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if err := decodeJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	writeJSON(w, http.StatusOK, body)
}
