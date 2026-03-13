package model

import (
	"time"
)

type ChecklistItem struct {
	ID        int    `json:"id"`
	Completed bool   `json:"completed"`
	Text      string `json:"text"`
	Position  int    `json:"position"`
}
type Note struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	Title     string          `json:"title"`
	Type      string          `json:"type"`
	Body      string          `json:"body"`
	Items     []ChecklistItem `json:"items,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
