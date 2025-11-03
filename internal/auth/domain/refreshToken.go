package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	Jti       uuid.UUID
	SessionID uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
	Revoked   bool
}
