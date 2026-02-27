package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"desent/model"
)

func (h *Handler) GetBooks(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	all := h.store.ListBooks()
	result := make([]model.Book, 0, len(all))
	for _, b := range all {
		if author == "" || strings.EqualFold(b.Author, author) {
			result = append(result, b)
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })

	if pageStr != "" && limitStr != "" {
		page, err1 := strconv.Atoi(pageStr)
		limit, err2 := strconv.Atoi(limitStr)
		if err1 == nil && err2 == nil && page > 0 && limit > 0 {
			start := (page - 1) * limit
			if start >= len(result) {
				result = []model.Book{}
			} else {
				end := start + limit
				if end > len(result) {
					end = len(result)
				}
				result = result[start:end]
			}
		}
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var b model.Book
	if err := decodeJSON(r, &b); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(b.Title) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	if strings.TrimSpace(b.Author) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "author is required"})
		return
	}
	writeJSON(w, http.StatusCreated, h.store.CreateBook(b))
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	b, ok := h.store.GetBook(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	existing, ok := h.store.GetBook(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	var update model.Book
	if err := decodeJSON(r, &update); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(update.Title) != "" {
		existing.Title = update.Title
	}
	if strings.TrimSpace(update.Author) != "" {
		existing.Author = update.Author
	}
	if update.Year != 0 {
		existing.Year = update.Year
	}
	h.store.UpdateBook(id, existing)
	writeJSON(w, http.StatusOK, existing)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	if !h.store.DeleteBook(id) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
