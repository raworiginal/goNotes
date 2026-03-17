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
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &user, nil
}

func UpdateUser(db *sql.DB, user *model.User) error {
	query := `
	UPDATE users
	SET username = $1, password = $2, role = $3, updated_at = CURRENT_TIMESTAMP
	WHERE id = $4
	`
	result, err := db.Exec(query, user.Username, user.Password, user.Role)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func DeleteUser(db *sql.DB, userID int) error {
	query := `
	DELETE FROM users
	WHERE id = $1
	`
	result, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
