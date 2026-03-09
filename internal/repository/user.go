package repository

import (
	"database/sql"
	"fmt"

	"github.com/raworiginal/goNotes/internal/model"
)

func Create(db *sql.DB, username, hashedPassword string) (*model.User, error) {
	var user model.User
	query := "INSERT INTO users (username, password) VALUES (?,?)"
	result, err := db.Exec(query, username, hashedPassword)
	if err != nil {
		return nil, err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("lastInseretId: %w", err)
	}
	query = "SELECT * FROM users where id = ?"
	err = db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
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

func FindByUsername(db *sql.DB, username string) (*model.User, error) {
	var user model.User
	query := "Select id, username, password, created_at FROM users Where username = ?"

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
