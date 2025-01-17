package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password,omitempty"`
	Token          string    `json:"token,omitempty"`
	RefreshToken   string    `json:"refresh_token,omitempty"`
	IsChirpyRed    bool      `json:"is_chirpy_red,omitempty"`
}
