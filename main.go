package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// --- Models ---

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year,omitempty"`
}

// --- Store ---

type Store struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
	tokens map[string]struct{}
}

var db = &Store{
	books:  make(map[int]Book),
	nextID: 1,
	tokens: make(map[string]struct{}),
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// --- Middleware ---

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		db.mu.RLock()
		_, ok := db.tokens[token]
		db.mu.RUnlock()
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		next(w, r)
	}
}

// --- Level 1: Ping ---

func ping(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// --- Level 2: Echo ---

func echo(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if err := decodeJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	writeJSON(w, http.StatusOK, body)
}

// --- Level 5: Auth Token ---

func authToken(w http.ResponseWriter, r *http.Request) {
	token := fmt.Sprintf("%016x", rand.Int63())
	db.mu.Lock()
	db.tokens[token] = struct{}{}
	db.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// --- Level 3 + 6: Get Books (search & pagination) ---

func getBooks(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	db.mu.RLock()
	result := make([]Book, 0, len(db.books))
	for _, b := range db.books {
		if author != "" && !strings.EqualFold(b.Author, author) {
			continue
		}
		result = append(result, b)
	}
	db.mu.RUnlock()

	// Sort by ID for consistent ordering
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].ID < result[j-1].ID; j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}

	// Pagination
	if pageStr != "" && limitStr != "" {
		page, err1 := strconv.Atoi(pageStr)
		limit, err2 := strconv.Atoi(limitStr)
		if err1 == nil && err2 == nil && page > 0 && limit > 0 {
			start := (page - 1) * limit
			if start >= len(result) {
				result = []Book{}
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

// --- Level 3: Create Book (with Level 7 validation) ---

func createBook(w http.ResponseWriter, r *http.Request) {
	var b Book
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

	db.mu.Lock()
	b.ID = db.nextID
	db.nextID++
	db.books[b.ID] = b
	db.mu.Unlock()

	writeJSON(w, http.StatusCreated, b)
}

// --- Level 3: Get Book by ID (with Level 7 not-found) ---

func getBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	db.mu.RLock()
	b, ok := db.books[id]
	db.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// --- Level 4: Update Book ---

func updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var update Book
	if err := decodeJSON(r, &update); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	db.mu.Lock()
	b, ok := db.books[id]
	if !ok {
		db.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	if strings.TrimSpace(update.Title) != "" {
		b.Title = update.Title
	}
	if strings.TrimSpace(update.Author) != "" {
		b.Author = update.Author
	}
	if update.Year != 0 {
		b.Year = update.Year
	}
	db.books[id] = b
	db.mu.Unlock()

	writeJSON(w, http.StatusOK, b)
}

// --- Level 4: Delete Book ---

func deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	db.mu.Lock()
	_, ok := db.books[id]
	if !ok {
		db.mu.Unlock()
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "book not found"})
		return
	}
	delete(db.books, id)
	db.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// --- Main ---

func main() {
	mux := http.NewServeMux()

	// Level 1: Ping
	mux.HandleFunc("GET /ping", ping)

	// Level 2: Echo
	mux.HandleFunc("POST /echo", echo)

	// Level 5: Auth
	mux.HandleFunc("POST /auth/token", authToken)

	// Books (GET /books is protected via auth middleware)
	mux.HandleFunc("GET /books", authMiddleware(getBooks))
	mux.HandleFunc("POST /books", createBook)
	mux.HandleFunc("GET /books/{id}", getBook)
	mux.HandleFunc("PUT /books/{id}", updateBook)
	mux.HandleFunc("DELETE /books/{id}", deleteBook)

	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
