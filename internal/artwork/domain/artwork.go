package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
)

type ArtworkStatus = generated.ArtworkStatus

const (
	ArtworkStatusAvailable   ArtworkStatus = "available"
	ArtworkStatusPending     ArtworkStatus = "pending"
	ArtworkStatusSold        ArtworkStatus = "sold"
	ArtworkStatusNotForSale  ArtworkStatus = "not_for_sale"
	ArtworkStatusUnavailable ArtworkStatus = "unavailable"
	ArtworkStatusComingSoon  ArtworkStatus = "coming_soon"
)

type ArtworkMedium = generated.ArtworkMedium

const (
	ArtworkMediumOilPanel     ArtworkMedium = "oil_panel"
	ArtworkMediumAcrylicPanel ArtworkMedium = "acrylic_panel"
	ArtworkMediumOilMdf       ArtworkMedium = "oil_mdf"
	ArtworkMediumOilPaper     ArtworkMedium = "oil_paper"
	ArtworkMediumUnknown      ArtworkMedium = "unknown"
)

type ArtworkCategory = generated.ArtworkCategory

const (
	ArtworkCategoryFigure      ArtworkCategory = "figure"
	ArtworkCategoryLandscape   ArtworkCategory = "landscape"
	ArtworkCategoryMultiFigure ArtworkCategory = "multi_figure"
	ArtworkCategoryOther       ArtworkCategory = "other"
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
	Status         ArtworkStatus
	Medium         ArtworkMedium
	Category       ArtworkCategory
	Images         []Image
	CreatedAt      time.Time
	OrderId        *uuid.UUID
}
