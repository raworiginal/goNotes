package repository

import (
	"database/sql"
	"fmt"

	"github.com/raworiginal/goNotes/internal/model"
)

func CreateUser(db *sql.DB, username, hashedPassword string) (*model.User, error) {
	var user model.User

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	defer tx.Rollback()

	query := `
	INSERT INTO users (username, password) 
	VALUES ($1,$2)
	RETURNING id, username, password, created_at
	`
	err = tx.QueryRow(query, username, hashedPassword).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func FindByUsername(db *sql.DB, username string) (*model.User, error) {
	var user model.User

	query := `
	SELECT id, username, password, created_at 
	FROM users 
	WHERE username = $1
	`

	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, sql.ErrNoRows
		default:
			return nil, err
		}
	}
	return &user, nil
}
