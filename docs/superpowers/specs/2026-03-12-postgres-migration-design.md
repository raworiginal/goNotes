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

**File:** `docker-compose.yml`

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

**Usage:**
- `docker-compose up -d` to start PostgreSQL
- `docker-compose down` to stop
- Default connection string: `postgres://postgres:postgres@localhost:5432/gonotes`

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
- Remove `Items *string` field (was holding JSON)
- Keep `Items []ChecklistItem` in memory only (not a DB field, loaded separately)
- Add a struct field to distinguish from the DB struct if needed

**No changes needed:**
- `user.go`
- `token.go`

### 5. Repository Layer (`internal/repository/notes.go`)

**Current approach:**
- `Create()`: Marshals checklist items to JSON, stores in `items` field
- `FindByID()`: Unmarshals JSON into items
- `Update()`: Re-marshals JSON

**Target approach:**
- `Create()`: Inserts note, then loops through items to insert each one with position
- `FindByID()`: Queries note, then queries `checklist_items` ORDER BY position
- `Update()`: Updates note fields, then handles items (delete old, insert new)
- `Delete()`: Cascades automatically via FK constraint

**New queries to write:**
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

### 6. Handlers (`internal/handler/notes.go`)

**No logic changes needed.** Handlers still work with `Note` structs that have `Items []ChecklistItem`. The repository handles converting between the request/response format and the relational schema.

---

## API Compatibility

Request/response shapes remain unchanged:

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

The refactoring is purely internal to the database layer. Handlers and API contracts stay the same.

---

## Testing Strategy

1. **Spin up PostgreSQL:** `docker-compose up -d`
2. **Update `.env`:** Point `DATABASE_URL` to the running container
3. **Start the API:** `go run ./cmd/api`
4. **Test manually with curl:**
   - Register a user
   - Create a checklist note with multiple items
   - Verify items come back in order
   - Update the note (modify/add/remove items)
   - Delete the note (verify cascade)

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

