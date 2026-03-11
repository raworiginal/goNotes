package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/raworiginal/goNotes/internal/config"
	am "github.com/raworiginal/goNotes/internal/middleware"
	"github.com/raworiginal/goNotes/internal/model"
	repo "github.com/raworiginal/goNotes/internal/repository"
)

type NotesHandler struct {
	db  *sql.DB
	cfg *config.Config
}

type NewNoteRequest struct {
	Title string                `json:"title"`
	Type  string                `json:"type"`
	Body  string                `json:"body"`
	Items []model.ChecklistItem `json:"items,omitempty"`
}

type PatchNoteRequest struct {
	Title *string                `json:"title"`
	Type  *string                `json:"type"`
	Body  *string                `json:"body"`
	Items *[]model.ChecklistItem `json:"items,omitempty"`
}

func NewNotesHandler(db *sql.DB, cfg *config.Config) *NotesHandler {
	return &NotesHandler{db: db, cfg: cfg}
}

func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}
	notes, err := repo.FindAllNotesByUser(h.db, userID)
	if err != nil {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req NewNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.Type != "text" && req.Type != "checklist" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}
	// Marshal items if checklist
	var itemsJSON string
	if req.Type == "checklist" {
		itemsBytes, err := json.Marshal(req.Items)
		if err != nil {
			http.Error(w, `{"error":"invalid items"}`, http.StatusBadRequest)
			return
		}
		itemsJSON = string(itemsBytes)
	}
	// if req.Type == "text", itemsJSON stays empty

	note := &model.Note{
		UserID: userID,
		Title:  req.Title,
		Type:   req.Type,
		Body:   req.Body,
		Items:  itemsJSON,
	}

	// Create the note
	created, err := repo.CreateNote(h.db, note)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *NotesHandler) GetNoteByID(w http.ResponseWriter, r *http.Request) {
	noteIDString := chi.URLParam(r, "id")
	noteID, err := strconv.Atoi(noteIDString)
	if err != nil {
		http.Error(w, `{"error":"invalid noteID"}`, http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	note, err := repo.FindNoteByID(h.db, userID, noteID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		} else {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)
}

func (h *NotesHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	noteIDstring := chi.URLParam(r, "id")
	noteID, err := strconv.Atoi(noteIDstring)
	if err != nil {
		http.Error(w, `{"error":"invalid note id"}`, http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	note, err := repo.FindNoteByID(h.db, userID, noteID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		} else {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
	}
	var req NewNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.Type != "text" && req.Type != "checklist" {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	// Marshal items if checklist
	var itemsJSON string
	if req.Type == "checklist" {
		itemsBytes, err := json.Marshal(req.Items)
		if err != nil {
			http.Error(w, `{"error":"invalid items"}`, http.StatusBadRequest)
			return
		}
		itemsJSON = string(itemsBytes)
	}
	// if req.Type == "text", itemsJSON stays empty

	note = &model.Note{
		ID:     noteID,
		UserID: userID,
		Title:  req.Title,
		Type:   req.Type,
		Body:   req.Body,
		Items:  itemsJSON,
	}

	updatedNote, err := repo.UpdateNote(h.db, note)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedNote)
}

func (h *NotesHandler) PatchNote(w http.ResponseWriter, r *http.Request) {
	noteIDstring := chi.URLParam(r, "id")
	noteID, err := strconv.Atoi(noteIDstring)
	if err != nil {
		http.Error(w, `{"error":"invalid note id"}`, http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	note, err := repo.FindNoteByID(h.db, userID, noteID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		} else {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
	}
	var req PatchNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Title != nil {
		note.Title = *req.Title
	}
	if req.Type != nil {
		note.Type = *req.Type
	}
	if req.Body != nil {
		note.Body = *req.Body
	}
	if req.Items != nil {
		itemsBytes, err := json.Marshal(req.Items)
		if err != nil {
			http.Error(w, `{"error":"invalid items"}`, http.StatusBadRequest)
			return
		}
		note.Items = string(itemsBytes)
	}
	updatedNote, err := repo.UpdateNote(h.db, note)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedNote)
}

func (h *NotesHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	noteIDstring := chi.URLParam(r, "id")
	noteID, err := strconv.Atoi(noteIDstring)
	if err != nil {
		http.Error(w, `{"error":"invalid note id"}`, http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	_, err = repo.FindNoteByID(h.db, userID, noteID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		} else {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
	}
	if err := repo.DeleteNote(h.db, userID, noteID); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
