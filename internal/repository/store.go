package repository

import "database/sql"

type Store struct {
	Notes  NoteRepository
	Users  UserRepository
	Tokens TokenRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Notes:  NewNoteRepository(db),
		Users:  NewUserRepository(db),
		Tokens: NewTokenRepository(db),
	}
}
