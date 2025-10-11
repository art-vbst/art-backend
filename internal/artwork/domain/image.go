package domain

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID          uuid.UUID
	ArtworkID   uuid.UUID
	IsMainImage bool
	ImageURL    string
	ImageWidth  *int32
	ImageHeight *int32
	CreatedAt   time.Time
}
