package postgres

import (
	"time"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
)

func toDomainArtwork(row *generated.Artwork) (*domain.Artwork, error) {
	widthInches, err := row.WidthInches.Float64Value()
	if err != nil {
		return nil, err
	}

	heightInches, err := row.HeightInches.Float64Value()
	if err != nil {
		return nil, err
	}

	var soldAt *time.Time
	if row.SoldAt.Valid {
		soldAt = &row.SoldAt.Time
	}

	return &domain.Artwork{
		ID:             row.ID,
		Title:          row.Title,
		PaintingNumber: row.PaintingNumber,
		PaintingYear:   row.PaintingYear,
		WidthInches:    widthInches.Float64,
		HeightInches:   heightInches.Float64,
		PriceCents:     row.PriceCents,
		Paper:          row.Paper,
		SortOrder:      row.SortOrder,
		SoldAt:         soldAt,
		Status:         row.Status,
		Medium:         row.Medium,
		Category:       row.Category,
		CreatedAt:      row.CreatedAt.Time,
	}, nil
}
