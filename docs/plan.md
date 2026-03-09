# Notes API - Project Plan

A reference implementation for building idiomatic REST APIs in Go. Use this as a model for future projects.

---

## Stack

| Concern   | Choice                             |
| --------- | ---------------------------------- |
| Language  | Go 1.22+                           |
| Router    | chi v5                             |
| Auth      | golang-jwt/jwt v5 + refresh tokens |
| Database  | SQLite via mattn/go-sqlite3        |
| DB access | database/sql directly              |
| Passwords | bcrypt via golang.org/x/crypto     |
| Config    | .env via joho/godotenv             |

---

## Project Structure

```
notes-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point, wires everything together
├── internal/
│   ├── config/
│   │   └── config.go            # Loads env vars (JWT secret, token TTLs, DB path, port)
│   ├── db/
│   │   ├── db.go                # Opens DB connection, runs migrations
│   │   └── migrations/
│   │       └── 001_init.sql     # Schema: users, notes, refresh_tokens
│   ├── middleware/
│   │   ├── auth.go              # JWT validation, injects user_id into context
│   │   └── logger.go            # Request logging
│   ├── handler/
│   │   ├── auth.go              # Register, login, refresh, logout
│   │   └── notes.go             # CRUD for notes
│   ├── model/
│   │   └── model.go             # User, Note, ChecklistItem, RefreshToken structs
│   ├── repository/
│   │   ├── user.go              # DB queries for users
│   │   ├── notes.go             # DB queries for notes
│   │   └── token.go             # DB queries for refresh tokens
│   └── token/
│       └── token.go             # JWT creation, parsing, refresh token generation
├── docs/
│   └── plan.md                  # This file
├── .env.example
├── go.mod
└── go.sum
```

### Why this structure

`internal/` prevents any outside packages from importing your code. This is idiomatic Go for application code. Each layer has one job:

- `handler` handles HTTP concerns (parsing requests, writing responses)
- `repository` handles all SQL
- `model` holds your data shapes with no logic attached
- `token` handles all JWT logic in one place
- `config` reads env vars once at startup and passes them down

This maps to what you know from Express: router -> middleware -> controller -> service/model. Go just makes the boundaries more explicit.

---

## Database Schema

```sql
-- 001_init.sql

CREATE TABLE IF NOT EXISTS users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    username   TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    type       TEXT NOT NULL CHECK(type IN ('text', 'checklist')),
    body       TEXT NOT NULL DEFAULT '',   -- plain text when type='text'
    items      TEXT,                       -- JSON array when type='checklist'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## API Endpoints

### Auth (public)

| Method | Path           | Description                             |
| ------ | -------------- | --------------------------------------- |
| POST   | /auth/register | Create account                          |
| POST   | /auth/login    | Returns access token + refresh token    |
| POST   | /auth/refresh  | Swap refresh token for new access token |
| POST   | /auth/logout   | Revoke refresh token                    |

### Notes (protected)

All routes require `Authorization: Bearer <access_token>` header.

| Method | Path        | Description                        |
| ------ | ----------- | ---------------------------------- |
| GET    | /notes      | List all notes for the authed user |
| POST   | /notes      | Create a note                      |
| GET    | /notes/{id} | Get a single note                  |
| PUT    | /notes/{id} | Replace a note                     |
| PATCH  | /notes/{id} | Partial update                     |
| DELETE | /notes/{id} | Delete a note                      |

---

## Request / Response Shapes

### Register and Login request

```json
{ "username": "alice", "password": "secret" }
```

### Login response

```json
{
  "access_token": "eyJ...",
  "refresh_token": "a1b2c3..."
}
```

### Note (text type)

```json
{
  "title": "Shopping",
  "type": "text",
  "body": "Milk, eggs, bread"
}
```

### Note (checklist type)

```json
{
  "title": "Shopping",
  "type": "checklist",
  "items": [
    { "text": "Milk", "completed": false },
    { "text": "Eggs", "completed": true }
  ]
}
```

---

## Auth Flow

1. Register creates a user with a bcrypt-hashed password.
2. Login validates credentials and returns a short-lived JWT access token and a long-lived refresh token. The refresh token is a random string stored in the `refresh_tokens` table.
3. The client sends the access token as `Authorization: Bearer <token>` on every protected request.
4. When the access token expires, the client hits `/auth/refresh` with the refresh token to get a new access token.
5. Logout deletes the refresh token row from the DB, revoking it immediately.
6. The `auth` middleware parses and validates the JWT, then injects `user_id` into the request context. Handlers read the user ID from context to scope all DB queries.

---

## Config

`.env.example`:

```
PORT=8080
DB_PATH=./notes.db
JWT_SECRET=change_me_in_production
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
```

---

## Dependencies

```
github.com/go-chi/chi/v5
github.com/golang-jwt/jwt/v5
github.com/mattn/go-sqlite3
golang.org/x/crypto
github.com/joho/godotenv
```

Five dependencies. No ORM, no code generation, no framework.

---

## Build Order

Work through phases in order. Each phase is fully functional before moving to the next.

### Phase 1: Scaffold

- Initialize Go module (`go mod init`)
- Create folder structure
- Write a minimal `main.go` that starts an HTTP server and responds to `GET /health`
- Goal: confirm the server runs and chi is wired up correctly

### Phase 2: Config

- Write `internal/config/config.go`
- Load all env vars at startup using `godotenv`
- Pass the config struct into every component that needs it
- Goal: no magic strings anywhere in the codebase

### Phase 3: Database

- Write `internal/db/db.go` to open the SQLite connection
- Read and execute `001_init.sql` on startup
- Goal: tables exist when the server starts, migration is idempotent

### Phase 4: Models and Repositories

- Write all structs in `internal/model/model.go`
- Write `repository/user.go`: `Create`, `FindByUsername`
- Write `repository/notes.go`: `Create`, `FindByID`, `FindAllByUser`, `Update`, `Delete`
- Write `repository/token.go`: `Create`, `FindByToken`, `Delete`, `DeleteExpired`
- Goal: all DB access lives in one place, handlers never write SQL

### Phase 5: Token Package

- Write `internal/token/token.go`
- JWT: sign a token with user ID and expiry, parse and validate a token
- Refresh token: generate a cryptographically random string
- Goal: auth logic is isolated and testable on its own

### Phase 6: Auth Handlers

- Write `internal/handler/auth.go`
- Implement register, login, refresh, logout
- Goal: full auth flow works end to end, testable with curl

### Phase 7: Auth Middleware

- Write `internal/middleware/auth.go`
- Parse `Authorization` header, validate JWT, inject `user_id` into context
- Write a helper to read `user_id` from context in handlers
- Goal: protected routes reject requests with missing or invalid tokens

### Phase 8: Notes Handlers

- Write `internal/handler/notes.go`
- All handlers read `user_id` from context and pass it to the repository
- Validate that `type` is either `text` or `checklist`
- Return 404 when a note exists but belongs to a different user
- Goal: full CRUD works, users are fully isolated from each other

### Phase 9: Wire Up

- Update `main.go` to assemble all dependencies and register all routes
- Add a request logger middleware
- Goal: complete, running API

---

## Error Handling Conventions

- Return `400` for malformed requests or validation failures
- Return `401` for missing or invalid tokens
- Return `403` for valid token but wrong user (though returning `404` is also acceptable to avoid leaking existence)
- Return `404` for records not found
- Return `500` for unexpected DB errors, log the actual error server-side
- All error responses use the same JSON shape: `{ "error": "message here" }`

---

## Key Go Patterns You Will Learn

- **`internal/` package**: prevents accidental imports, enforces boundaries
- **Dependency injection by hand**: pass deps as struct fields, no magic containers
- **Context for request-scoped values**: how middleware passes data to handlers
- **`database/sql`**: manual query writing, scanning rows into structs
- **Error wrapping**: `fmt.Errorf("findNote: %w", err)` for traceable errors
- **`http.Handler` and chi**: how Go's standard handler interface works
- **Struct tags**: `json:"field_name"` for encoding/decoding request bodies
