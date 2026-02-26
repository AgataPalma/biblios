package users

import (
	"time"

	"github.com/AgataPalma/biblios/internal/apictx"
)

type User struct {
	ID           string      `json:"id"`
	Email        string      `json:"email"`
	Username     string      `json:"username"`
	PasswordHash string      `json:"-"`
	Role         apictx.Role `json:"role"`
	IsAdmin      bool        `json:"-"`
	Theme        string      `json:"theme"`
	DeletedAt    *time.Time  `json:"deleted_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
