package repository

import (
	"database/sql"
	"fmt"

	"github.com/raworiginal/goNotes/internal/model"
)

type PGNoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *PGNoteRepository {
	return &PGNoteRepository{db: db}
}

func (r *PGNoteRepository) CreateNote(note *model.Note) (*model.Note, error) {
	tx, err := r.db.Begin()
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

func (r *PGNoteRepository) FindNoteByID(noteID int) (*model.Note, error) {
	var note model.Note

	query := `
	SELECT id, user_id, title, type, body, created_at, updated_at 
	FROM notes 
	WHERE id = $1
	`
	err := r.db.QueryRow(query, noteID).Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find note by id: %w", err)
	}

	if note.Type == "checklist" {
		note.Items, err = r.getCheckListItems(noteID)
		if err != nil {
			return nil, err
		}
	}
	return &note, nil
}

func (r *PGNoteRepository) ListNotes(userID int) ([]*model.Note, error) {
	var notes []*model.Note
	query := `
	SELECT id, user_id, title, type, body, created_at, updated_at FROM notes Where user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("find all notes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var note model.Note
		err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Type, &note.Body, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		if note.Type == "checklist" {
			note.Items, err = r.getCheckListItems(note.ID)
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

func (r *PGNoteRepository) UpdateNote(note *model.Note) (*model.Note, error) {
	tx, err := r.db.Begin()
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
		return nil, ErrNotFound
	}

	_, err = tx.Exec(`DELETE FROM checklist_items WHERE note_id = $1`, note.ID)
	if err != nil {
		return nil, fmt.Errorf("delete checklist items: %w", err)
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
				return nil, fmt.Errorf("insert checklist item: %w", err)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	note, err = r.FindNoteByID(note.ID)
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}

	return note, nil
}

func (r *PGNoteRepository) DeleteNote(noteID int) error {
	query := `
DELETE FROM notes WHERE id = $1
	`
	result, err := r.db.Exec(query, noteID)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PGNoteRepository) getCheckListItems(noteID int) ([]model.ChecklistItem, error) {
	var items []model.ChecklistItem
	checklistQuery := `
		SELECT id, position, text, completed
		FROM checklist_items
		WHERE note_id = $1
		ORDER BY position
		`
	rows, err := r.db.Query(checklistQuery, noteID)
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
