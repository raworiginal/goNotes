# PostgreSQL Migration Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate goNotes from SQLite to PostgreSQL and refactor checklist items from JSON into a dedicated relational table.

**Architecture:** Layer-by-layer implementation starting with infrastructure (Docker, config), then database layer (schema, migrations), then data layer (models, repositories), finally presentation layer (handlers). Each layer builds on the previous; all code is testable in isolation.

**Tech Stack:** PostgreSQL (via lib/pq driver), Go's database/sql, chi (router), docker-compose for local development.

**Design Reference:** See `docs/superpowers/specs/2026-03-12-postgres-migration-design.md`

---

## File Structure

### Files to Create
- `docker-compose.yml` — PostgreSQL service definition

### Files to Modify
- `.env.example` — add DATABASE_URL
- `.env` — set DATABASE_URL for local development
- `go.mod` — add github.com/lib/pq dependency
- `internal/config/config.go` — read DATABASE_URL instead of DB_PATH
- `internal/db/db.go` — switch from SQLite to PostgreSQL driver
- `internal/db/migrations/001_init.sql` — rewrite schema for PostgreSQL
- `internal/model/note.go` — remove Items field, add ID to ChecklistItem
- `internal/repository/notes.go` — separate item loading, use transactions
- `internal/handler/notes.go` — add NoteResponse struct, load items separately

---

## Chunk 1: Infrastructure & Configuration

### Task 1: Create docker-compose.yml

**Files:**
- Create: `docker-compose.yml`

- [ ] **Step 1: Create docker-compose.yml at project root**

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:latest-alpine
    container_name: gonotes-postgres
    environment:
      POSTGRES_DB: gonotes
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

- [ ] **Step 2: Verify docker-compose syntax**

Run: `docker-compose config`
Expected: Valid configuration output with no errors

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml
git commit -m "infra: add docker-compose for postgresql"
```

---

### Task 2: Update .env and .env.example

**Files:**
- Modify: `.env.example`
- Modify: `.env`

- [ ] **Step 1: Update .env.example**

Replace `DB_PATH` with `DATABASE_URL`:

```
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/gonotes
JWT_SECRET=change_me_in_production
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
```

- [ ] **Step 2: Update .env**

Set the same DATABASE_URL:

```
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/gonotes
JWT_SECRET=change_me_in_production
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
```

- [ ] **Step 3: Commit**

```bash
git add .env.example .env
git commit -m "config: update env vars for postgresql"
```

---

### Task 3: Add lib/pq dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add lib/pq dependency**

Run: `go get github.com/lib/pq`

Expected: go.mod updated with `github.com/lib/pq` entry

- [ ] **Step 2: Verify go.mod updated**

Run: `grep "github.com/lib/pq" go.mod`

Expected: Line showing `github.com/lib/pq` with version

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add lib/pq for postgresql driver"
```

---

## Chunk 2: Database Layer

### Task 4: Update internal/config/config.go

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Review current config.go**

Read the file to understand current structure. Identify where `DB_PATH` is loaded.

- [ ] **Step 2: Replace DB_PATH with DATABASE_URL**

In the Config struct, change:
```go
// OLD
DBPath string

// NEW
DatabaseURL string
```

In the Load() function, change:
```go
// OLD
cfg.DBPath = os.Getenv("DB_PATH")

// NEW
cfg.DatabaseURL = os.Getenv("DATABASE_URL")
```

- [ ] **Step 3: Test config loads correctly**

Run: `go build ./internal/config`

Expected: No compilation errors

- [ ] **Step 4: Commit**

```bash
git add internal/config/config.go
git commit -m "config: read DATABASE_URL instead of DB_PATH"
```

---

### Task 5: Update internal/db/db.go

**Files:**
- Modify: `internal/db/db.go`

- [ ] **Step 1: Review current db.go**

Read the file to understand how it currently opens SQLite and runs migrations.

- [ ] **Step 2: Replace SQLite import with PostgreSQL and update Open() function**

Change the import from:
```go
// OLD
_ "github.com/mattn/go-sqlite3"
```

To:
```go
// NEW
_ "github.com/lib/pq"
```

Change the `Open()` function signature and implementation from:
```go
// OLD
func Open(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }
    // Read and execute the migration
    migrationSQL, err := os.ReadFile("internal/db/migrations/001_init.sql")
    // ... execute migration
}
```

To:
```go
// NEW
func Open(databaseURL string) (*sql.DB, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }
    // Read and execute the migration
    migrationSQL, err := os.ReadFile("internal/db/migrations/001_init.sql")
    if err != nil {
        return nil, fmt.Errorf("read migration: %w", err)
    }
    if _, err := db.Exec(string(migrationSQL)); err != nil {
        return nil, fmt.Errorf("execute migration: %w", err)
    }
    return db, nil
}
```

The migration execution logic stays the same, but now it will run the PostgreSQL schema instead of SQLite.

- [ ] **Step 3: Update cmd/api/main.go to pass DATABASE_URL**

Find where `db.Open()` is called (likely in main.go). Change:
```go
// OLD
database, err := db.Open(cfg.DBPath)

// NEW
database, err := db.Open(cfg.DatabaseURL)
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./cmd/api`

Expected: No compilation errors

- [ ] **Step 5: Commit**

```bash
git add internal/db/db.go cmd/api/main.go
git commit -m "db: switch to postgresql driver (lib/pq)"
```

---

### Task 6: Rewrite internal/db/migrations/001_init.sql

**Files:**
- Modify: `internal/db/migrations/001_init.sql`

- [ ] **Step 1: Backup current file (optional but safe)**

The file is in version control, so no need for manual backup, but know what you're replacing.

- [ ] **Step 2: Replace entire file with PostgreSQL schema**

```sql
-- 001_init.sql - PostgreSQL schema for goNotes

CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notes (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    type       VARCHAR(20) NOT NULL CHECK(type IN ('text', 'checklist')),
    body       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS checklist_items (
    id         SERIAL PRIMARY KEY,
    note_id    INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    completed  BOOLEAN NOT NULL DEFAULT false,
    position   INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

- [ ] **Step 3: Verify SQL syntax (optional)**

If you have psql installed locally, you can validate syntax. Otherwise, the database will catch errors on first run.

- [ ] **Step 4: Commit**

```bash
git add internal/db/migrations/001_init.sql
git commit -m "db: rewrite migration for postgresql schema"
```

---

## Chunk 3: Data Models

### Task 7: Update internal/model/note.go

**Files:**
- Modify: `internal/model/note.go`

- [ ] **Step 1: Review current note.go**

Identify the `Note` and `ChecklistItem` struct definitions.

- [ ] **Step 2: Update Note struct**

Remove the `Items` field entirely:

```go
// OLD
type Note struct {
    ID        int
    UserID    int
    Title     string
    Type      string
    Body      string
    Items     string  // JSON
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NEW
type Note struct {
    ID        int
    UserID    int
    Title     string
    Type      string
    Body      string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

- [ ] **Step 3: Update ChecklistItem struct**

Add `ID` field:

```go
// OLD
type ChecklistItem struct {
    Text      string
    Completed bool
}

// NEW
type ChecklistItem struct {
    ID        int
    Text      string
    Completed bool
}
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./internal/model`

Expected: No compilation errors

- [ ] **Step 5: Commit**

```bash
git add internal/model/note.go
git commit -m "model: remove Items from Note, add ID to ChecklistItem"
```

---

## Chunk 4: Repository Layer

### Task 8: Update internal/repository/notes.go

**Files:**
- Modify: `internal/repository/notes.go`

The repository uses package-level functions (not struct methods). Update these functions to:
1. Remove the `items` field from all queries
2. Add new `FindChecklistItems()` function
3. Update `CreateNote()` to insert items in a transaction
4. Update `UpdateNote()` to use transactions

- [ ] **Step 1: Review current notes.go**

Understand the current function signatures and how they query the `items` field.

- [ ] **Step 2: Add FindChecklistItems() function**

Add this new function:

```go
func FindChecklistItems(db *sql.DB, noteID int) ([]model.ChecklistItem, error) {
    rows, err := db.Query(
        "SELECT id, text, completed FROM checklist_items WHERE note_id = $1 ORDER BY position",
        noteID,
    )
    if err != nil {
        return nil, fmt.Errorf("query checklist items: %w", err)
    }
    defer rows.Close()

    var items []model.ChecklistItem
    for rows.Next() {
        var item model.ChecklistItem
        if err := rows.Scan(&item.ID, &item.Text, &item.Completed); err != nil {
            return nil, fmt.Errorf("scan item: %w", err)
        }
        items = append(items, item)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate items: %w", err)
    }
    return items, nil
}
```

- [ ] **Step 3: Update CreateNote() function**

Change the function to accept items and use a transaction:

```go
// OLD signature:
// func CreateNote(db *sql.DB, note *model.Note) (*model.Note, error)

// NEW signature and implementation:
func CreateNote(db *sql.DB, note *model.Note, items []model.ChecklistItem) (*model.Note, error) {
    tx, err := db.Begin()
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()

    // Insert note (no items column)
    query := "INSERT INTO notes (user_id, title, type, body, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at"
    err = tx.QueryRow(query, note.UserID, note.Title, note.Type, note.Body, time.Now(), time.Now()).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("insert note: %w", err)
    }

    // Insert items if type is "checklist"
    if note.Type == "checklist" {
        for i, item := range items {
            _, err := tx.Exec(
                "INSERT INTO checklist_items (note_id, text, completed, position) VALUES ($1, $2, $3, $4)",
                note.ID, item.Text, item.Completed, i+1, // position starts at 1
            )
            if err != nil {
                return nil, fmt.Errorf("insert item: %w", err)
            }
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit tx: %w", err)
    }
    return note, nil
}
```

- [ ] **Step 4: Update FindNoteByID() function**

Remove the `items` column from the SELECT:

```go
// OLD:
// query := "SELECT id, user_id, title, type, body, items, created_at, updated_at FROM notes Where id = ? AND user_id = ?"

// NEW:
func FindNoteByID(db *sql.DB, userID, noteID int) (*model.Note, error) {
    var note model.Note
    query := "SELECT id, user_id, title, type, body, created_at, updated_at FROM notes WHERE id = $1 AND user_id = $2"
    err := db.QueryRow(query, noteID, userID).Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.CreatedAt, &note.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, sql.ErrNoRows
    }
    if err != nil {
        return nil, err
    }
    return &note, nil
}
```

- [ ] **Step 5: Update FindAllNotesByUser() function**

Remove the `items` column from the SELECT:

```go
// OLD:
// query := "SELECT id, user_id, title, type, body, items, created_at, updated_at FROM notes Where user_id = ? ORDER BY created_at DESC"

// NEW:
func FindAllNotesByUser(db *sql.DB, userID int) ([]*model.Note, error) {
    var notes []*model.Note
    query := "SELECT id, user_id, title, type, body, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY created_at DESC"
    rows, err := db.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("query notes: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var note model.Note
        err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.CreatedAt, &note.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("scan note: %w", err)
        }
        notes = append(notes, &note)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows error: %w", err)
    }
    return notes, nil
}
```

- [ ] **Step 6: Update UpdateNote() function**

Change to accept items and use a transaction:

```go
// OLD:
// func UpdateNote(db *sql.DB, note *model.Note) (*model.Note, error)

// NEW:
func UpdateNote(db *sql.DB, note *model.Note, items []model.ChecklistItem) (*model.Note, error) {
    tx, err := db.Begin()
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()

    // Update note (no items column)
    query := "UPDATE notes SET title = $1, type = $2, body = $3, updated_at = $4 WHERE user_id = $5 AND id = $6"
    _, err = tx.Exec(query, note.Title, note.Type, note.Body, time.Now(), note.UserID, note.ID)
    if err != nil {
        return nil, fmt.Errorf("update note: %w", err)
    }

    // Delete old items
    _, err = tx.Exec("DELETE FROM checklist_items WHERE note_id = $1", note.ID)
    if err != nil {
        return nil, fmt.Errorf("delete old items: %w", err)
    }

    // Insert new items if type is "checklist"
    if note.Type == "checklist" {
        for i, item := range items {
            _, err := tx.Exec(
                "INSERT INTO checklist_items (note_id, text, completed, position) VALUES ($1, $2, $3, $4)",
                note.ID, item.Text, item.Completed, i+1,
            )
            if err != nil {
                return nil, fmt.Errorf("insert item: %w", err)
            }
        }
    }

    // Re-fetch the note to get updated timestamps
    query = "SELECT id, user_id, title, type, body, created_at, updated_at FROM notes WHERE user_id = $1 AND id = $2"
    err = tx.QueryRow(query, note.UserID, note.ID).Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.CreatedAt, &note.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("fetch updated note: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit tx: %w", err)
    }
    return note, nil
}
```

- [ ] **Step 7: Verify compilation**

Run: `go build ./internal/repository`

Expected: No compilation errors

- [ ] **Step 8: Commit**

```bash
git add internal/repository/notes.go
git commit -m "repository: remove items field, add FindChecklistItems, use transactions"
```

---

## Chunk 5: Handler Layer

### Task 9: Update internal/handler/notes.go

**Files:**
- Modify: `internal/handler/notes.go`

The handlers need to:
1. Define a `NoteResponse` struct that combines Note and Items for API responses
2. Load items separately after loading notes
3. Pass items to the repository on Create and Update

- [ ] **Step 1: Review current notes.go**

Understand the current handler structure and how it calls the repository.

- [ ] **Step 2: Add NoteResponse struct**

Add this at the top of the notes.go file:

```go
// NoteResponse is the JSON response shape for API clients
// It combines Note and items into a single response object
type NoteResponse struct {
    ID        int              `json:"id"`
    UserID    int              `json:"user_id"`
    Title     string           `json:"title"`
    Type      string           `json:"type"`
    Body      string           `json:"body"`
    Items     []model.ChecklistItem `json:"items,omitempty"`
    CreatedAt time.Time        `json:"created_at"`
    UpdatedAt time.Time        `json:"updated_at"`
}

// ToResponse converts a Note + items into a NoteResponse for the API
func toNoteResponse(note *model.Note, items []model.ChecklistItem) NoteResponse {
    return NoteResponse{
        ID:        note.ID,
        UserID:    note.UserID,
        Title:     note.Title,
        Type:      note.Type,
        Body:      note.Body,
        Items:     items,
        CreatedAt: note.CreatedAt,
        UpdatedAt: note.UpdatedAt,
    }
}
```

- [ ] **Step 3: Update GET /notes/{id} handler**

Modify to load items separately:

```go
// OLD
func (h *NotesHandler) GetNote(w http.ResponseWriter, r *http.Request) {
    // ...
    note, _ := h.repo.FindByID(ctx, id)
    json.NewEncoder(w).Encode(note)
}

// NEW
func (h *NotesHandler) GetNote(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(int)
    idStr := chi.URLParam(r, "id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, `{"error": "invalid note id"}`, http.StatusBadRequest)
        return
    }

    note, err := h.repo.FindByID(r.Context(), id)
    if err != nil {
        http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
        return
    }
    if note == nil || note.UserID != userID {
        http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
        return
    }

    // Load items if checklist
    var items []model.ChecklistItem
    if note.Type == "checklist" {
        var err error
        items, err = h.repo.FindChecklistItems(r.Context(), note.ID)
        if err != nil {
            http.Error(w, `{"error": "failed to load items"}`, http.StatusInternalServerError)
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(toNoteResponse(note, items))
}
```

- [ ] **Step 4: Update POST /notes (Create) handler**

Modify to pass items to repository:

```go
// OLD
func (h *NotesHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Title string                `json:"title"`
        Type  string                `json:"type"`
        Body  string                `json:"body"`
        Items string                `json:"items"` // JSON string
    }
    // ...
    h.repo.Create(ctx, note)
}

// NEW
func (h *NotesHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(int)

    var req struct {
        Title string              `json:"title"`
        Type  string              `json:"type"`
        Body  string              `json:"body"`
        Items []model.ChecklistItem `json:"items"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
        return
    }

    if req.Type != "text" && req.Type != "checklist" {
        http.Error(w, `{"error": "type must be 'text' or 'checklist'"}`, http.StatusBadRequest)
        return
    }

    note := &model.Note{
        UserID: userID,
        Title:  req.Title,
        Type:   req.Type,
        Body:   req.Body,
    }

    if err := h.repo.Create(r.Context(), note, req.Items); err != nil {
        http.Error(w, `{"error": "failed to create note"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(toNoteResponse(note, req.Items))
}
```

- [ ] **Step 5: Update PUT /notes/{id} (Update) handler**

Modify signature to pass items:

```go
// OLD
func (h *NotesHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
    // ...
    h.repo.Update(ctx, note)
}

// NEW
func (h *NotesHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(int)
    idStr := chi.URLParam(r, "id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, `{"error": "invalid note id"}`, http.StatusBadRequest)
        return
    }

    var req struct {
        Title string              `json:"title"`
        Type  string              `json:"type"`
        Body  string              `json:"body"`
        Items []model.ChecklistItem `json:"items"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
        return
    }

    if req.Type != "text" && req.Type != "checklist" {
        http.Error(w, `{"error": "type must be 'text' or 'checklist'"}`, http.StatusBadRequest)
        return
    }

    note, err := h.repo.FindByID(r.Context(), id)
    if err != nil {
        http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
        return
    }
    if note == nil || note.UserID != userID {
        http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
        return
    }

    note.Title = req.Title
    note.Type = req.Type
    note.Body = req.Body

    if err := h.repo.Update(r.Context(), note, req.Items); err != nil {
        http.Error(w, `{"error": "failed to update note"}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(toNoteResponse(note, req.Items))
}
```

- [ ] **Step 6: Verify compilation**

Run: `go build ./cmd/api`

Expected: No compilation errors

- [ ] **Step 7: Commit**

```bash
git add internal/handler/notes.go
git commit -m "handlers: add NoteResponse struct, load items separately"
```

---

## Chunk 6: Integration & Testing

### Task 10: End-to-End Testing

**Files:**
- No code changes; testing only

- [ ] **Step 1: Start PostgreSQL**

Run: `docker-compose up -d`

Expected: Output showing postgres container starting

- [ ] **Step 2: Verify database is up**

Run: `docker-compose exec postgres psql -U postgres -d gonotes -c "SELECT 1;"`

Expected: Output showing `1`

- [ ] **Step 3: Build the API**

Run: `go build -o gonotes-api ./cmd/api`

Expected: Binary created with no errors

- [ ] **Step 4: Start the API**

Run: `./gonotes-api`

Expected: Server starts, no panic errors (logs show listening on port 8080)

- [ ] **Step 5: Test Register**

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}'
```

Expected: Response with `access_token` and `refresh_token`

- [ ] **Step 6: Extract access token**

Save the access token from the response (e.g., `TOKEN=eyJ...`)

- [ ] **Step 7: Test Create Text Note**

```bash
curl -X POST http://localhost:8080/notes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Text Note",
    "type": "text",
    "body": "Some text content",
    "items": []
  }'
```

Expected: 201 response with note details, `items` is empty array

- [ ] **Step 8: Test Create Checklist Note**

```bash
curl -X POST http://localhost:8080/notes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Shopping List",
    "type": "checklist",
    "body": "",
    "items": [
      {"text": "Milk", "completed": false},
      {"text": "Bread", "completed": false},
      {"text": "Eggs", "completed": true}
    ]
  }'
```

Expected: 201 response with items in order (Milk, Bread, Eggs), each with `completed` status

- [ ] **Step 9: Test GET Note**

Extract the note ID from the previous response and:

```bash
curl http://localhost:8080/notes/<note_id> \
  -H "Authorization: Bearer $TOKEN"
```

Expected: 200 response matching the created note with items in order

- [ ] **Step 10: Test UPDATE Note**

```bash
curl -X PUT http://localhost:8080/notes/<note_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Shopping",
    "type": "checklist",
    "body": "",
    "items": [
      {"text": "Milk", "completed": true},
      {"text": "Butter", "completed": false}
    ]
  }'
```

Expected: 200 response with updated items (Milk marked complete, Butter added)

- [ ] **Step 11: Test DELETE Note**

```bash
curl -X DELETE http://localhost:8080/notes/<note_id> \
  -H "Authorization: Bearer $TOKEN"
```

Expected: 204 response

- [ ] **Step 12: Verify cascade delete**

Query the database to confirm items are gone:

```bash
docker-compose exec postgres psql -U postgres -d gonotes -c "SELECT * FROM checklist_items WHERE note_id = <note_id>;"
```

Expected: No rows returned

- [ ] **Step 13: Stop the server and database**

Press Ctrl+C to stop the API. Then:

```bash
docker-compose down
```

Expected: Containers stopped cleanly

---

### Task 11: Final Verification & Cleanup

**Files:**
- No code changes; review only

- [ ] **Step 1: Verify all commits are in order**

Run: `git log --oneline | head -20`

Expected: All commits from this migration present (docker-compose, env, deps, config, db changes, model, repository, handler)

- [ ] **Step 2: Check git status**

Run: `git status`

Expected: Clean working tree (no uncommitted changes)

- [ ] **Step 3: Verify no references to SQLite remain**

Run: `grep -r "sqlite" --include="*.go" .`

Expected: No results (all sqlite references should be gone)

- [ ] **Step 4: Build one final time**

Run: `go build ./cmd/api`

Expected: Binary builds with no errors or warnings

- [ ] **Step 5: Review spec vs. implementation**

Check that all items in the design spec (models, repository signatures, handlers) match what was implemented. Reference `docs/superpowers/specs/2026-03-12-postgres-migration-design.md`.

Expected: No discrepancies

---

## Success Criteria

✅ All tasks completed with passing tests
✅ All commits made
✅ API starts without errors
✅ Checklist items stored in separate table and retrieved in order
✅ Transactions prevent partial writes
✅ All layers (config, db, models, repository, handlers) updated
✅ No SQLite references remain in code

