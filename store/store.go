package store

import (
	"sync"

	"desent/model"
)

type Store struct {
	mu     sync.RWMutex
	books  map[int]model.Book
	nextID int
	tokens map[string]struct{}
}

func New() *Store {
	return &Store{
		books:  make(map[int]model.Book),
		nextID: 1,
		tokens: make(map[string]struct{}),
	}
}

func (s *Store) AddToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = struct{}{}
}

func (s *Store) ValidateToken(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.tokens[token]
	return ok
}

func (s *Store) CreateBook(b model.Book) model.Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.ID = s.nextID
	s.nextID++
	s.books[b.ID] = b
	return b
}

func (s *Store) GetBook(id int) (model.Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.books[id]
	return b, ok
}

func (s *Store) ListBooks() []model.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Book, 0, len(s.books))
	for _, b := range s.books {
		result = append(result, b)
	}
	return result
}

func (s *Store) UpdateBook(id int, b model.Book) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return false
	}
	s.books[id] = b
	return true
}

func (s *Store) DeleteBook(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return false
	}
	delete(s.books, id)
	return true
}
