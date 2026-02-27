# desent

A REST API built in Go implementing 8 progressive levels — from a simple ping to a full CRUD service with auth, search, pagination, and error handling.

## Requirements

- Go 1.22+

## Run

```bash
go run main.go
# Server listening on :8080
```

## Project Structure

```
desent/
├── main.go              # Entry point — route wiring
├── go.mod
├── model/
│   └── book.go          # Book struct
├── store/
│   └── store.go         # Thread-safe in-memory store
├── middleware/
│   └── auth.go          # Bearer token auth middleware
└── handler/
    ├── handler.go        # Handler struct, JSON helpers
    ├── ping.go           # GET /ping
    ├── echo.go           # POST /echo
    ├── auth.go           # POST /auth/token
    └── books.go          # Books CRUD + search + pagination
```

## API Reference

### Level 1 — Ping

```
GET /ping
```

**Response** `200`
```json
{ "success": true }
```

---

### Level 2 — Echo

```
POST /echo
```

**Body** any JSON object

**Response** `200` — same JSON echoed back

```json
{ "message": "hello", "number": 42 }
```

---

### Level 3 — CRUD: Create & Read

#### Create a book

```
POST /books
```

**Body**
```json
{
  "title": "The Go Programming Language",
  "author": "Alan Donovan",
  "year": 2015
}
```

**Response** `201`
```json
{ "id": 1, "title": "The Go Programming Language", "author": "Alan Donovan", "year": 2015 }
```

#### List all books

```
GET /books
Authorization: Bearer <token>
```

**Response** `200`
```json
[
  { "id": 1, "title": "The Go Programming Language", "author": "Alan Donovan", "year": 2015 }
]
```

#### Get book by ID

```
GET /books/:id
```

**Response** `200`
```json
{ "id": 1, "title": "The Go Programming Language", "author": "Alan Donovan", "year": 2015 }
```

---

### Level 4 — CRUD: Update & Delete

#### Update a book

```
PUT /books/:id
```

**Body** (all fields optional — only provided fields are updated)
```json
{ "title": "Updated Title" }
```

**Response** `200` — updated book

#### Delete a book

```
DELETE /books/:id
```

**Response** `204 No Content`

---

### Level 5 — Auth Guard

#### Get a token

```
POST /auth/token
```

**Body**
```json
{ "username": "admin", "password": "password" }
```

**Response** `200`
```json
{ "token": "2f548d1fdb45eb03" }
```

Use the token on protected endpoints:

```
Authorization: Bearer 2f548d1fdb45eb03
```

`GET /books` requires a valid token — returns `401` without one.

---

### Level 6 — Search & Paginate

Filter by author:
```
GET /books?author=Alan Donovan
Authorization: Bearer <token>
```

Paginate:
```
GET /books?page=1&limit=2
Authorization: Bearer <token>
```

Both params can be combined:
```
GET /books?author=Alan Donovan&page=1&limit=2
Authorization: Bearer <token>
```

---

### Level 7 — Error Handling

| Scenario | Status | Body |
|---|---|---|
| Missing `title` on POST /books | `400` | `{"error":"title is required"}` |
| Missing `author` on POST /books | `400` | `{"error":"author is required"}` |
| Invalid JSON body | `400` | `{"error":"invalid JSON"}` |
| Book ID not found | `404` | `{"error":"book not found"}` |
| Non-numeric book ID | `404` | `{"error":"book not found"}` |
| Missing/invalid token | `401` | `{"error":"unauthorized"}` |
| Wrong credentials | `401` | `{"error":"invalid credentials"}` |

---

### Level 8 — Boss: Speed Run

All endpoints from levels 1–7 are active simultaneously. Every endpoint responds in under 500ms.

## Example Flow

```bash
# 1. Ping
curl http://localhost:8080/ping

# 2. Get a token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/token \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 3. Create books
curl -s -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"The Go Programming Language","author":"Alan Donovan","year":2015}'

curl -s -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Clean Code","author":"Robert Martin","year":2008}'

# 4. List books
curl http://localhost:8080/books -H "Authorization: Bearer $TOKEN"

# 5. Search by author
curl "http://localhost:8080/books?author=Alan Donovan" -H "Authorization: Bearer $TOKEN"

# 6. Paginate
curl "http://localhost:8080/books?page=1&limit=1" -H "Authorization: Bearer $TOKEN"

# 7. Update
curl -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"The Go Programming Language (2nd ed.)"}'

# 8. Delete
curl -X DELETE http://localhost:8080/books/1
```
