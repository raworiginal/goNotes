package handler

import (
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
	store *repo.Store
	cfg   *config.Config
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

func NewNotesHandler(store *repo.Store, cfg *config.Config) *NotesHandler {
	return &NotesHandler{store, cfg}
}

func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	notes, err := h.store.Notes.ListNotes(userID)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req *model.Note
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid json request", http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.Type != "text" && req.Type != "checklist" {
		jsonError(w, "title and type required", http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	req.UserID = userID

	created, err := h.store.Notes.CreateNote(req)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *NotesHandler) GetNoteByID(w http.ResponseWriter, r *http.Request) {
	noteIDString := chi.URLParam(r, "id")
	noteID, err := strconv.Atoi(noteIDString)
	if err != nil {
		jsonError(w, "invalid noteID", http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	note, err := h.store.Notes.FindNoteByID(noteID)
	if err != nil {
		if err == repo.ErrNotFound {
			jsonError(w, "note not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if note.UserID != userID {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)
}

func (h *NotesHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	noteIDstring := chi.URLParam(r, "id")
	noteID, err := strconv.Atoi(noteIDstring)
	if err != nil {
		jsonError(w, "invalid note id", http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	note, err := h.store.Notes.FindNoteByID(noteID)
	if err != nil {
		if err == repo.ErrNotFound {
			jsonError(w, "note not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if note.UserID != userID {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req *model.Note
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.Type != "text" && req.Type != "checklist" {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	req.ID = noteID

	updatedNote, err := h.store.Notes.UpdateNote(req)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
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
		jsonError(w, "invalid note id", http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	note, err := h.store.Notes.FindNoteByID(noteID)
	if err != nil {
		if err == repo.ErrNotFound {
			jsonError(w, "note not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if note.UserID != userID {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req PatchNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
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
		note.Items = *req.Items
	}

	updatedNote, err := h.store.Notes.UpdateNote(note)
	if err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
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
		jsonError(w, "invalid note id", http.StatusBadRequest)
		return
	}
	userID, err := am.UserIDFromContext(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	note, err := h.store.Notes.FindNoteByID(noteID)
	if err != nil {
		if err == repo.ErrNotFound {
			jsonError(w, "note not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if note.UserID != userID {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.store.Notes.DeleteNote(noteID); err != nil {
		jsonError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
