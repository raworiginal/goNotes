package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type PGTokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *PGTokenRepository {
	return &PGTokenRepository{db}
}

func (r *PGTokenRepository) CreateToken(userID int, token string, expiresAt time.Time) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	defer tx.Rollback()

	query := `
	INSERT INTO refresh_tokens (user_id, token, expires_at) 
	VALUES ($1,$2,$3)
	`
	_, err = tx.Exec(query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	return nil
}

func (r *PGTokenRepository) FindUserIDByToken(token string) (int, error) {
	var userID int
	query := `
	SELECT user_id FROM refresh_tokens 
	WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP
	`
	err := r.db.QueryRow(query, token).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("find token: %w", err)
	}
	return userID, nil
}

func (r *PGTokenRepository) DeleteToken(token string) error {
	query := `
	DELETE FROM refresh_tokens 
	WHERE token = $1
	`
	_, err := r.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}

func (r *PGTokenRepository) DeleteExpiredToken() error {
	query := `
	DELETE FROM refresh_tokens 
	WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("delete expired token: %w", err)
	}
	return nil
}
