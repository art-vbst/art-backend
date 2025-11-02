package postgres

import (
	"time"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
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

func toDomainImage(row *generated.Image) *domain.Image {
	return &domain.Image{
		ArtworkID:   uuid.UUID(row.ArtworkID.Bytes),
		ID:          row.ID,
		ObjectName:  row.ObjectName,
		ImageURL:    row.ImageUrl,
		IsMainImage: row.IsMainImage,
		ImageWidth:  row.ImageWidth,
		ImageHeight: row.ImageHeight,
	}
}
