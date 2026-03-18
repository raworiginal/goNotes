package repository

type Store struct {
	Notes  NoteRepository
	Users  UserRepository
	Tokens TokenRepository
}
