# PostgreSQL Migration with Checklist Items Refactoring

**Date:** 2026-03-12
**Status:** Design Phase
**Author:** Design Brainstorm Session

---

## Overview

Migrate the goNotes API from SQLite to PostgreSQL and refactor the database schema to move checklist items from a JSON field into a dedicated relational table. This is a fresh-start migration with no existing data to preserve.

---

## Goals

1. **Modernize storage**: Switch from SQLite to PostgreSQL for better concurrency, scalability, and relational features
2. **Proper schema design**: Create a dedicated `checklist_items` table instead of storing items as JSON
3. **Learning outcome**: Demonstrate idiomatic PostgreSQL design and Go patterns for working with relational databases
4. **Docker integration**: Run PostgreSQL locally via docker-compose for consistent development environments

---

## Current State vs Target State

### Current Schema (SQLite)

```sql
CREATE TABLE users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    username   TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    type       TEXT NOT NULL CHECK(type IN ('text', 'checklist')),
    body       TEXT NOT NULL DEFAULT '',
    items      TEXT,  -- JSON array when type='checklist'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE refresh_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Target Schema (PostgreSQL)

```sql
CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notes (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    type       VARCHAR(20) NOT NULL CHECK(type IN ('text', 'checklist')),
    body       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE checklist_items (
    id         SERIAL PRIMARY KEY,
    note_id    INTEGER NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    completed  BOOLEAN NOT NULL DEFAULT false,
    position   INTEGER NOT NULL  -- Tracks insertion order
);

CREATE TABLE refresh_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Key Changes:**
- Remove `items` JSON field from `notes` table
- Create new `checklist_items` table with `position` field to preserve order
- Switch from SQLite types (INTEGER, DATETIME, TEXT) to PostgreSQL types (SERIAL, TIMESTAMP, VARCHAR)

---

## Docker Compose Setup

**File location:** Create `docker-compose.yml` at the **project root** (same level as `go.mod`, `cmd/`, `internal/`, `docs/`)

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

**Commit to git:** Yes, this should be committed to git so all developers use the same database setup.

**Usage:**
- `docker-compose up -d` to start PostgreSQL
- `docker-compose down` to stop
- Default connection string: `postgres://postgres:postgres@localhost:5432/gonotes`

**.env file:**
Update `.env.example` and `.env` to include:
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/gonotes
```

You can remove `DB_PATH` since SQLite is no longer used. The config should require `DATABASE_URL` to be set.

---

## Code Changes Required

### 1. Configuration Layer (`internal/config/config.go`)

**Current:** Reads `DB_PATH` for SQLite file location
**Target:** Read `DATABASE_URL` environment variable for PostgreSQL connection string

Environment variable example:
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/gonotes
```

### 2. Database Layer (`internal/db/db.go`)

**Current:**
- Uses `mattn/go-sqlite3` driver
- Imports `database/sql` with `_ "github.com/mattn/go-sqlite3"`

**Target:**
- Uses `lib/pq` (PostgreSQL driver for Go)
- Imports `database/sql` with `_ "github.com/lib/pq"`
- Parses `DATABASE_URL` and opens connection with `sql.Open("postgres", dsn)`
- Migration file changes (see below)

**New dependency to add:**
```
github.com/lib/pq
```

### 3. Migration Files

**Current:** `internal/db/migrations/001_init.sql` (SQLite syntax)

**Target:** Rewrite `001_init.sql` for PostgreSQL with the new schema shown above

The migration remains idempotent with `CREATE TABLE IF NOT EXISTS`.

### 4. Models (`internal/model/`)

**`note.go` changes:**
- Remove the `Items` field entirely from the Note struct (was holding JSON)
- Update `ChecklistItem` struct to include an `ID` field: `ID int`
- This means the Note struct returned by the repository will NOT include items; they must be loaded separately

Example updated structs:
```go
type Note struct {
    ID        int
    UserID    int
    Title     string
    Type      string
    Body      string
    CreatedAt time.Time
    UpdatedAt time.Time
    // NO Items field
}

type ChecklistItem struct {
    ID        int    // NEW: add this
    Text      string
    Completed bool
    // NO Position field in the struct (internal to DB only)
}
```

**No changes needed:**
- `user.go`
- `token.go`

### 5. Repository Layer (`internal/repository/notes.go`)

**Current approach:**
- `Create()`: Marshals checklist items to JSON, stores in `items` field
- `FindByID()`: Unmarshals JSON into items, returns complete Note
- `Update()`: Re-marshals JSON

**Target approach:**
- `Create(note *Note, items []ChecklistItem)`: Inserts note in a transaction, then loops through items to insert each one with position
- `FindByID(id int)`: Returns Note WITHOUT items (items must be loaded separately)
- `FindChecklistItems(noteID int)`: NEW function to load items for a note, ordered by position
- `Update(note *Note, items []ChecklistItem)`: Updates note fields in transaction, deletes old items, inserts new items
- `Delete()`: Cascades automatically via FK constraint

**Key design decision:** The repository will have separate methods for loading notes and loading items. Handlers are responsible for calling both when needed.

**New method signatures:**
```go
// In repository/notes.go
func (r *NotesRepository) Create(ctx context.Context, note *Note, items []ChecklistItem) error {
    // BEGIN TRANSACTION
    // Insert note
    // Loop through items, INSERT each with position (1, 2, 3...)
    // COMMIT TRANSACTION
}

func (r *NotesRepository) FindChecklistItems(ctx context.Context, noteID int) ([]ChecklistItem, error) {
    // SELECT id, text, completed FROM checklist_items
    // WHERE note_id = $1 ORDER BY position
}

func (r *NotesRepository) Update(ctx context.Context, note *Note, items []ChecklistItem) error {
    // BEGIN TRANSACTION
    // UPDATE note SET ...
    // DELETE FROM checklist_items WHERE note_id = $1
    // Loop through items, INSERT each with position (1, 2, 3...)
    // COMMIT TRANSACTION
}
```

**SQL queries to write:**
```sql
-- Insert a checklist item
INSERT INTO checklist_items (note_id, text, completed, position)
VALUES ($1, $2, $3, $4)

-- Get items for a note
SELECT id, text, completed FROM checklist_items
WHERE note_id = $1
ORDER BY position

-- Delete items for a note (when updating)
DELETE FROM checklist_items WHERE note_id = $1
```

**Position assignment:** When inserting items in `Create()` or `Update()`, assign position as `1, 2, 3...` based on the order they appear in the `items []ChecklistItem` slice passed from the handler.

### 6. Handlers (`internal/handler/notes.go`)

**Changes required:**

For **read operations** (`GET /notes/{id}`):
- Call `repo.FindByID(id)` to get the Note
- Call `repo.FindChecklistItems(noteID)` to load items
- Combine them in a response struct that includes both Note and Items
- This ensures the API response still shows `"items": [...]` for checklist notes

For **create operations** (`POST /notes`):
- Parse request body as before
- Extract items array from request
- Call `repo.Create(note, items)` instead of `repo.Create(note)`

For **update operations** (`PUT/PATCH /notes/{id}`):
- Parse request body as before
- Extract items array from request
- Call `repo.Update(note, items)` instead of `repo.Update(note)`

**Response helper:** Create a response struct that combines Note + Items for API responses, since the Note struct no longer contains Items:
```go
type NoteResponse struct {
    ID        int              `json:"id"`
    UserID    int              `json:"user_id"`
    Title     string           `json:"title"`
    Type      string           `json:"type"`
    Body      string           `json:"body"`
    Items     []ChecklistItem  `json:"items,omitempty"`  // Only for checklist type
    CreatedAt time.Time        `json:"created_at"`
    UpdatedAt time.Time        `json:"updated_at"`
}
```

This keeps the API contract unchanged while separating the database schema concerns.

---

## API Compatibility

**Request and response shapes remain unchanged for clients:**

```json
{
  "id": 1,
  "title": "Shopping",
  "type": "checklist",
  "items": [
    { "text": "Milk", "completed": false },
    { "text": "Eggs", "completed": true }
  ],
  "created_at": "2026-03-12T10:00:00Z",
  "updated_at": "2026-03-12T10:00:00Z"
}
```

**Internal changes:**
- Handlers will use a `NoteResponse` struct to combine the Note and ChecklistItems before sending to the client
- This keeps the API contract unchanged while separating database concerns
- The items no longer have an `id` field in the API response (clients don't need to reference individual items yet)

---

## Transaction Handling

When creating or updating a note with items, both operations must succeed together or both must fail. Use database transactions:

```go
func (r *NotesRepository) Create(ctx context.Context, note *Note, items []ChecklistItem) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()  // Rolled back if anything fails

    // INSERT note using tx
    // INSERT each item using tx
    // If any INSERT fails, all changes are rolled back

    return tx.Commit().Error  // Only commits if all succeeds
}
```

Same pattern for `Update()`. This prevents partial writes if an item insert fails partway through.

---

## Testing Strategy

1. **Spin up PostgreSQL:** `docker-compose up -d`
2. **Update `.env`:** Set `DATABASE_URL=postgres://postgres:postgres@localhost:5432/gonotes`
3. **Start the API:** `go run ./cmd/api`
4. **Test manually with curl:**
   - Register a user
   - Create a checklist note with multiple items, verify items come back in order
   - Create a text note, verify `items` array is empty or omitted
   - Update a checklist note (modify/add/remove items), verify items change
   - Delete the note and verify it's gone (cascade deletes items automatically)
   - Try creating a note with invalid `type` (should fail validation in handler)
5. **Verify cascading delete** by checking items are gone when note is deleted:
   ```bash
   # After deleting a note, query the database directly:
   docker-compose exec postgres psql -U postgres -d gonotes -c "SELECT * FROM checklist_items WHERE note_id = <deleted_id>;"
   # Should return 0 rows
   ```

---

## Migration Sequence

1. Add `lib/pq` dependency: `go get github.com/lib/pq`
2. Create `docker-compose.yml` at project root
3. Write new PostgreSQL migration: `internal/db/migrations/001_init.sql`
4. Update `internal/config/config.go` to read `DATABASE_URL`
5. Update `internal/db/db.go` to use PostgreSQL driver
6. Update `internal/model/note.go` (remove JSON field if applicable)
7. Update `internal/repository/notes.go` to use relational queries
8. Test end-to-end with curl

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| `position` INT field | Explicit ordering is more reliable than relying on insertion order. Allows future reordering without updating all items. |
| Separate `checklist_items` table | Proper normalization; enables filtering, sorting, and bulk operations on items without touching notes. |
| No archive/soft-delete | Fresh start; checklist items are created/updated/deleted, not archived. |
| Position is just an INT, not a sequence with gaps | Simpler for this use case; sufficient for ordering. |

---

## Out of Scope

- Data migration from SQLite (fresh start)
- Item-level timestamps (not required)
- Bulk operations on items
- Full-text search on item text
- Versioning/audit trail for items

