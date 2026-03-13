package repository

import (
	"database/sql"
	"fmt"

	"github.com/raworiginal/goNotes/internal/model"
)

func CreateNote(db *sql.DB, note *model.Note) (*model.Note, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	defer tx.Rollback()

	query := `
	INSERT INTO notes (user_id, title, type, body)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(query, note.UserID, note.Title, note.Type, note.Body).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	if note.Type == "checklist" {
		for i, item := range note.Items {
			query := `
			INSERT INTO checklist_items (note_id, text, completed, position)
			VALUES ($1, $2, $3, $4)
			RETURNING id
			`

			err = tx.QueryRow(query, note.ID, item.Text, item.Completed, i+1).Scan(&item.ID)
			if err != nil {
				return nil, err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}

	return note, nil
}

func FindNoteByID(db *sql.DB, noteID int) (*model.Note, error) {
	var note model.Note

	query := `
	SELECT id, user_id, title, type, body, created_at, updated_at 
	FROM notes 
	WHERE id = $1
	`
	err := db.QueryRow(query, noteID).Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, sql.ErrNoRows
		default:
			return nil, err
		}
	}

	if note.Type == "checklist" {
		note.Items, err = getCheckListItems(db, noteID)
		if err != nil {
			return nil, err
		}
	}
	return &note, nil
}

func FindAllNotesByUser(db *sql.DB, userID int) ([]*model.Note, error) {
	var notes []*model.Note
	query := `
	SELECT id, user_id, title, type, body, created_at, updated_at FROM notes Where user_id = $1 ORDER BY created_at DESC
	`
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
		err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		if note.Type == "checklist" {
			note.Items, err = getCheckListItems(db, note.ID)
			if err != nil {
				return nil, err
			}
		}

		notes = append(notes, &note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return notes, nil
}

func UpdateNote(db *sql.DB, note *model.Note) (*model.Note, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
	UPDATE notes 
	SET title = $1, type = $2, body = $3, updated_at = CURRENT_TIMESTAMP 
	WHERE id = $4
	`
	result, err := tx.Exec(query, note.Title, note.Type, note.Body, note.ID)
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	if note.Type == "checklist" {
		for i, item := range note.Items {
			query := `
			UPDATE checklist_items
			SET position = $1, text = $2, completed = $3
			WHERE id = $4
			`
			result, err := tx.Exec(query, i+1, item.Text, item.Completed, item.ID)
			if err != nil {
				return nil, fmt.Errorf("update checklist item: %w", err)
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return nil, fmt.Errorf("update checklist item: %w", err)
			}
			if rowsAffected == 0 {
				return nil, sql.ErrNoRows
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	note, err = FindNoteByID(db, note.ID)
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}

	return note, nil
}

func DeleteNote(db *sql.DB, noteID int) error {
	query := `
DELETE FROM notes WHERE id = $1
	`
	result, err := db.Exec(query, noteID)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func getCheckListItems(db *sql.DB, noteID int) ([]model.ChecklistItem, error) {
	var items []model.ChecklistItem
	checklistQuery := `
		SELECT id, position, text, completed
		FROM checklist_items
		WHERE note_id = $1
		ORDER BY position
		`
	rows, err := db.Query(checklistQuery, noteID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item model.ChecklistItem
		err := rows.Scan(&item.ID, &item.Position, &item.Text, &item.Completed)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, err
}
