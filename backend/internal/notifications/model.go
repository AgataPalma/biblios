package notifications

import (
	"encoding/json"
	"time"
)

type Notification struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Data      json.RawMessage `json:"data,omitempty"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// Well-known notification types
const (
	TypeLibraryInvitation = "library_invitation"
	TypeReviewLike        = "review_like"
	TypeLibraryActivity   = "library_activity"
)
