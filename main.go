package main

import (
	"fmt"
	"net/http"

	"desent/handler"
	"desent/middleware"
	"desent/store"
)

func main() {
	s := store.New()
	h := handler.New(s)
	mw := middleware.New(s)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", h.Ping)
	mux.HandleFunc("POST /echo", h.Echo)

	mux.HandleFunc("POST /auth/token", h.AuthToken)

	mux.HandleFunc("GET /books", mw.Auth(h.GetBooks))
	mux.HandleFunc("POST /books", h.CreateBook)
	mux.HandleFunc("GET /books/{id}", h.GetBook)
	mux.HandleFunc("PUT /books/{id}", h.UpdateBook)
	mux.HandleFunc("DELETE /books/{id}", h.DeleteBook)

	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
