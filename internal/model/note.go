package model

import (
	"time"
)

type ChecklistItem struct {
	Completed bool   `json:"completed"`
	Text      string `json:"text"`
}
type Note struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	Title     string          `json:"title"`
	Type      string          `json:"type"`
	Body      string          `json:"body"`
	Items     []ChecklistItem `json:"items"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
