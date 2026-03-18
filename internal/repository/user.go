package repository

import (
	"database/sql"
	"fmt"

	"github.com/raworiginal/goNotes/internal/model"
)

type PGUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *PGUserRepository {
	return &PGUserRepository{db: db}
}

func (r *PGUserRepository) CreateUser(db *sql.DB, username, hashedPassword string) (*model.User, error) {
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

func (r *PGUserRepository) FindUserByUsername(db *sql.DB, username string) (*model.User, error) {
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

func (r *PGUserRepository) ListUsers(db *sql.DB) ([]*model.User, error) {
	var users []*model.User
	query := `
	SELECT id, username, password, role, created_at, updated_at
	FROM users
	ORDER BY id ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list all users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

func (r *PGUserRepository) UpdateUser(db *sql.DB, user *model.User) error {
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

func (r *PGUserRepository) DeleteUser(db *sql.DB, userID int) error {
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
