package domain

import (
	"time"

	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
)

type ArtworkStatus = generated.ArtworkStatus

const (
	ArtworkStatusAvailable   ArtworkStatus = "available"
	ArtworkStatusSold        ArtworkStatus = "sold"
	ArtworkStatusNotForSale  ArtworkStatus = "not_for_sale"
	ArtworkStatusUnavailable ArtworkStatus = "unavailable"
	ArtworkStatusComingSoon  ArtworkStatus = "coming_soon"
)

type ArtworkMedium = generated.ArtworkMedium

const (
	ArtworkMediumOilOnPanel        ArtworkMedium = "oil_on_panel"
	ArtworkMediumAcrylicOnPanel    ArtworkMedium = "acrylic_on_panel"
	ArtworkMediumOilOnMdf          ArtworkMedium = "oil_on_mdf"
	ArtworkMediumOilOnOilPaper     ArtworkMedium = "oil_on_oil_paper"
	ArtworkMediumClaySculpture     ArtworkMedium = "clay_sculpture"
	ArtworkMediumPlasterSculpture  ArtworkMedium = "plaster_sculpture"
	ArtworkMediumInkOnPaper        ArtworkMedium = "ink_on_paper"
	ArtworkMediumMixedMediaOnPaper ArtworkMedium = "mixed_media_on_paper"
	ArtworkMediumUnknown           ArtworkMedium = "unknown"
)

type ArtworkCategory = generated.ArtworkCategory

const (
	ArtworkCategoryFigure      ArtworkCategory = "figure"
	ArtworkCategoryLandscape   ArtworkCategory = "landscape"
	ArtworkCategoryMultiFigure ArtworkCategory = "multi_figure"
	ArtworkCategoryOther       ArtworkCategory = "other"
)

type Artwork struct {
	ID             uuid.UUID       `json:"id"`
	Title          string          `json:"title"`
	PaintingNumber *int32          `json:"painting_number"`
	PaintingYear   *int32          `json:"painting_year"`
	WidthInches    float64         `json:"width_inches"`
	HeightInches   float64         `json:"height_inches"`
	PriceCents     int32           `json:"price_cents"`
	Description    string          `json:"description"`
	Paper          *bool           `json:"paper"`
	SortOrder      int32           `json:"sort_order"`
	SoldAt         *time.Time      `json:"sold_at"`
	Status         ArtworkStatus   `json:"status"`
	Medium         ArtworkMedium   `json:"medium"`
	Category       ArtworkCategory `json:"category"`
	Images         []Image         `json:"images"`
	CreatedAt      time.Time       `json:"created_at"`
	OrderId        *uuid.UUID      `json:"order_id"`
}
