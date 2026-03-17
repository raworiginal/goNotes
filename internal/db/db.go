package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/raworiginal/goNotes/internal/db/migrations"
)

func Open(databaseURL string) (*sql.DB, error) {
	// Open returns a connection pool, not a single connection
	connStr := databaseURL + "?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify  the connection actually works
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}
