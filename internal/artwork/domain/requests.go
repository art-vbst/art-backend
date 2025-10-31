package domain

type CreateRequest struct {
	Title          string          `json:"title"`
	PaintingNumber *int32          `json:"painting_number"`
	PaintingYear   *int32          `json:"painting_year"`
	WidthInches    float64         `json:"width_inches"`
	HeightInches   float64         `json:"height_inches"`
	PriceCents     int             `json:"price_cents"`
	Paper          bool            `json:"paper"`
	Status         ArtworkStatus   `json:"status"`
	Medium         ArtworkMedium   `json:"medium"`
	Category       ArtworkCategory `json:"category"`
}
