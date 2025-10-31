package postgres

import (
	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
)

func toDomainArtwork(row *generated.Artwork) *domain.Artwork {
	widthInches, _ := row.WidthInches.Float64Value()
	heightInches, _ := row.HeightInches.Float64Value()

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
		SoldAt:         &row.SoldAt.Time,
		Status:         row.Status,
		Medium:         row.Medium,
		Category:       row.Category,
		CreatedAt:      row.CreatedAt.Time,
	}
}
