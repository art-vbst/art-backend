package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	Revoked   bool
}
