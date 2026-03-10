package repository

import (
	"database/sql"
	"fmt"

	"github.com/raworiginal/goNotes/internal/model"
)

func CreateNote(db *sql.DB, userID int, title, noteType, body, items string) (*model.Note, error) {
	var note model.Note
	query := "INSERT INTO notes (user_id, title, type, body, items) VALUES (?,?,?,?,?)"
	result, err := db.Exec(query, userID, title, noteType, body, items)
	if err != nil {
		return nil, err
	}
	noteID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("LastInsertId: %w", err)
	}
	query = "SELECT id, user_id, title, type, body, items, created_at, updated_at FROM notes Where id = ?"
	err = db.QueryRow(query, noteID).Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.Items, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, sql.ErrNoRows
		default:
			return nil, err
		}
	}

	return &note, nil
}

func FindNoteByID(db *sql.DB, userID, noteID int) (*model.Note, error) {
	var note model.Note
	query := "SELECT id, user_id, title, type, body, items, created_at, updated_at FROM notes Where id = ? AND user_id = ?"
	err := db.QueryRow(query, noteID, userID).Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.Items, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, sql.ErrNoRows
		default:
			return nil, err
		}
	}
	return &note, nil
}

func FindAllNotesByUser(db *sql.DB, userID int) ([]*model.Note, error) {
	var notes []*model.Note
	query := "SELECT id, user_id, title, type, body, items, created_at, updated_at FROM notes Where user_id = ? ORDER BY created_at DESC"
	rows, err := db.Query(query, userID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, sql.ErrNoRows
		default:
			return nil, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var note model.Note
		err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.Items, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, &note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return notes, nil
}

func UpdateNote(db *sql.DB, userID, noteID int, title, noteType, body, items string) error {
	query := "UPDATE notes SET title = ?, type = ?, body = ?, items = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND id = ?"
	_, err := db.Exec(query, title, noteType, body, items, userID, noteID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteNote(db *sql.DB, userID, noteID int) error {
	query := "DELETE FROM notes WHERE user_id = ? AND id = ?"
	_, err := db.Exec(query, userID, noteID)
	if err != nil {
		return err
	}
	return nil
}
