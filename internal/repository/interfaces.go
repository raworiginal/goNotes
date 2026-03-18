package repository

import (
	"time"

	"github.com/raworiginal/goNotes/internal/model"
)

type NoteRepository interface {
	ListNotes(userID int) ([]*model.Note, error)
	FindNoteByID(noteID int) (*model.Note, error)
	CreateNote(note *model.Note) (*model.Note, error)
	UpdateNote(note *model.Note) (*model.Note, error)
	DeleteNote(noteID int) error
}

type UserRepository interface {
	CreateUser(username, hashedPassword string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	ListUsers() ([]*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(userID int) error
}

type TokenRepository interface {
	CreateToken(userID int, token string, expiresAt time.Time) error
	FindUserIDByToken(token string) (int, error)
	DeleteToken(token string) error
	DeleteExpiredToken() error
}
