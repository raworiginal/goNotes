package model

import "time"

type OAuthAccount struct {
	ID          int
	UserID      int
	Provider    string
	ProviderUID string
	Email       *string
	CreatedAt   time.Time
}
