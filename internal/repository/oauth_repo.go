package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/raworiginal/goNotes/internal/model"
)

type PGOAuthRepository struct {
	db *sql.DB
}

func NewOAuthRepository(db *sql.DB) *PGOAuthRepository {
	return &PGOAuthRepository{db: db}
}

func (r *PGOAuthRepository) FindUserByProvider(provider, providerUID string) (*model.User, error) {
	var user model.User
	query := `
	SELECT u.id, u.username, u.email, u.created_at
	FROM users u
	JOIN oauth_accounts oa ON oa.user_id = u.id
	WHERE oa.provider = $1 AND oa.provider_uid = $2
	`
	err := r.db.QueryRow(query, provider, providerUID).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by provider: %w", err)
	}
	return &user, nil
}

func (r *PGOAuthRepository) CreateOAuthUser(username string, email *string, provider, providerUID string) (*model.User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}
	defer tx.Rollback()

	var user model.User
	err = tx.QueryRow(
		`INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id, username, email, created_at`,
		username, email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO oauth_accounts (user_id, provider, provider_uid, email) VALUES ($1, $2, $3, $4)`,
		user.ID, provider, providerUID, email,
	)
	if err != nil {
		return nil, fmt.Errorf("create oauth account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}
	return &user, nil
}

func (r *PGOAuthRepository) CreateOAuthCode(userID int, code string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO oauth_codes (user_id, code, expires_at) VALUES ($1, $2, $3)`,
		userID, code, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("create oauth code: %w", err)
	}
	return nil
}

func (r *PGOAuthRepository) ExchangeOAuthCode(code string) (*model.User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("exchange oauth code: %w", err)
	}
	defer tx.Rollback()

	var userID int
	var expiresAt time.Time
	err = tx.QueryRow(
		`DELETE FROM oauth_codes WHERE code = $1 RETURNING user_id, expires_at`,
		code,
	).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("exchange oauth code: %w", err)
	}

	if time.Now().After(expiresAt) {
		return nil, ErrNotFound
	}

	var user model.User
	err = tx.QueryRow(
		`SELECT id, username, email, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("exchange oauth code: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("exchange oauth code: %w", err)
	}
	return &user, nil
}
