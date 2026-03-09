package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func Open(dbPath string) (*sql.DB, error) {
	// Open returns a connection pool, not a single connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify  the connection actually works
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
