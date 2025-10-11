package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
)

type Artwork struct {
	ID             uuid.UUID
	Title          string
	PaintingNumber *int32
	PaintingYear   *int32
	WidthInches    float64
	HeightInches   float64
	PriceCents     int32
	Paper          *bool
	SortOrder      *int32
	SoldAt         *time.Time
	Status         generated.ArtworkStatus
	Medium         generated.ArtworkMedium
	Category       generated.ArtworkCategory
	Images         []Image
	CreatedAt      time.Time
}
