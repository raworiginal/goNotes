# goNotes

A REST API for managing notes, built as a learning exercise in writing idiomatic Go. The goal was to explore Go conventions: the `cmd/internal` project layout, repository pattern with interfaces, JWT-based auth, database migrations, and clean separation of concerns across handlers, middleware, and data layers.

## Tech Stack

| Layer          | Technology                                                     |
| -------------- | -------------------------------------------------------------- |
| Language       | Go 1.25                                                        |
| Router         | [chi](https://github.com/go-chi/chi) v5                        |
| Database       | PostgreSQL 18                                                  |
| Migrations     | [Goose](https://github.com/pressly/goose) v3                   |
| Auth           | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt |
| CORS           | [go-chi/cors](https://github.com/go-chi/cors)                  |
| Config         | [godotenv](https://github.com/joho/godotenv)                   |
| Infrastructure | Docker Compose                                                 |

---

## Endpoints

### Health

#### `GET /health`

```bash
curl http://localhost:8080/health
```

```json
{ "status": "ok" }
```

---

### Auth

#### `POST /auth/register`

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "hunter2"}'
```

```json
{
  "id": 1,
  "role": "user",
  "username": "alice",
  "created_at": "2026-03-19T10:00:00Z",
  "updated_at": "2026-03-19T10:00:00Z"
}
```

#### `POST /auth/login`

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "hunter2"}'
```

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "d3f1a2b4c5e6..."
}
```

#### `POST /auth/refresh`

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "d3f1a2b4c5e6..."}'
```

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### `POST /auth/logout`

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "d3f1a2b4c5e6..."}'
```

`204 No Content`

---

### Notes

All notes endpoints require a valid JWT in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

#### `GET /notes/`

```bash
curl http://localhost:8080/notes/ \
  -H "Authorization: Bearer <access_token>"
```

```json
[
  {
    "id": 1,
    "user_id": 1,
    "title": "Shopping list",
    "type": "checklist",
    "body": "",
    "items": [
      { "id": 1, "completed": false, "text": "Milk", "position": 0 },
      { "id": 2, "completed": true, "text": "Eggs", "position": 1 }
    ],
    "created_at": "2026-03-19T10:05:00Z",
    "updated_at": "2026-03-19T10:05:00Z"
  }
]
```

#### `POST /notes/`

```bash
curl -X POST http://localhost:8080/notes/ \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Meeting notes",
    "type": "text",
    "body": "Discuss Q2 roadmap."
  }'
```

```json
{
  "id": 2,
  "user_id": 1,
  "title": "Meeting notes",
  "type": "text",
  "body": "Discuss Q2 roadmap.",
  "items": null,
  "created_at": "2026-03-19T11:00:00Z",
  "updated_at": "2026-03-19T11:00:00Z"
}
```

#### `GET /notes/{id}`

```bash
curl http://localhost:8080/notes/2 \
  -H "Authorization: Bearer <access_token>"
```

```json
{
  "id": 2,
  "user_id": 1,
  "title": "Meeting notes",
  "type": "text",
  "body": "Discuss Q2 roadmap.",
  "items": null,
  "created_at": "2026-03-19T11:00:00Z",
  "updated_at": "2026-03-19T11:00:00Z"
}
```

#### `PUT /notes/{id}`

Full replacement of a note.

```bash
curl -X PUT http://localhost:8080/notes/2 \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Meeting notes (updated)",
    "type": "text",
    "body": "Discuss Q2 roadmap and staffing."
  }'
```

```json
{
  "id": 2,
  "user_id": 1,
  "title": "Meeting notes (updated)",
  "type": "text",
  "body": "Discuss Q2 roadmap and staffing.",
  "items": null,
  "created_at": "2026-03-19T11:00:00Z",
  "updated_at": "2026-03-19T11:30:00Z"
}
```

#### `PATCH /notes/{id}`

Partial update — only provided fields are changed.

```bash
curl -X PATCH http://localhost:8080/notes/2 \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "Q2 Planning"}'
```

```json
{
  "id": 2,
  "user_id": 1,
  "title": "Q2 Planning",
  "type": "text",
  "body": "Discuss Q2 roadmap and staffing.",
  "items": null,
  "created_at": "2026-03-19T11:00:00Z",
  "updated_at": "2026-03-19T11:45:00Z"
}
```

#### `DELETE /notes/{id}`

```bash
curl -X DELETE http://localhost:8080/notes/2 \
  -H "Authorization: Bearer <access_token>"
```

`204 No Content`

---

## Setup

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

### 1. Clone and configure

```bash
git clone https://github.com/raworiginal/goNotes.git
cd goNotes
cp .env.example .env
```

Edit `.env` and set secure values for `POSTGRES_PASSWORD` and `JWT_SECRET`.

### 2. Start the database

```bash
docker-compose up -d
```

### 3. Run the API

```bash
make run
# or: go run ./cmd/api
```

Database migrations run automatically on startup. The server listens on the port specified in `.env` (default: `8080`).

### Environment Variables

| Variable            | Description                     | Default                                       |
| ------------------- | ------------------------------- | --------------------------------------------- |
| `PORT`              | Server port                     | `8080`                                        |
| `DATABASE_URL`      | PostgreSQL connection string    | —                                             |
| `JWT_SECRET`        | Secret for signing JWTs         | —                                             |
| `ACCESS_TOKEN_TTL`  | Access token lifetime           | `15m`                                         |
| `REFRESH_TOKEN_TTL` | Refresh token lifetime          | `168h`                                        |
| `CORS_ORIGINS`      | Comma-separated allowed origins | `http://localhost:3000,http://localhost:5173` |

### Makefile

```bash
make build   # Compile to bin/goNotes
make run     # Run via go run
make clean   # Remove bin/
```

---
