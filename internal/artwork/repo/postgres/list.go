package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
)

func (p *Postgres) ListArtworks(ctx context.Context) ([]domain.Artwork, error) {
	artworks, err := p.db.Queries().ListArtworks(ctx)
	if err != nil {
		return nil, err
	}
	return p.toDomainArtworkListRow(artworks), nil
}

func (p *Postgres) GetArtworkCheckoutData(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error) {
	artworks, err := p.db.Queries().ListArtworkStripeData(ctx, ids)
	if err != nil {
		return nil, err
	}
	return toDomainArtworkCheckoutListRow(artworks), nil
}

func (p *Postgres) toDomainArtworkListRow(rows []generated.ListArtworksRow) []domain.Artwork {
	artworks := []domain.Artwork{}

	for _, row := range rows {
		widthInches, _ := row.WidthInches.Float64Value()
		heightInches, _ := row.HeightInches.Float64Value()

		image := domain.Image{}
		if row.ImageID != uuid.Nil {
			image = domain.Image{
				ID:          row.ImageID,
				ArtworkID:   row.ID,
				IsMainImage: true,
				ImageURL:    row.ImageUrl,
				ImageWidth:  row.ImageWidth,
				ImageHeight: row.ImageHeight,
				CreatedAt:   row.ImageCreatedAt.Time,
			}
		}

		artwork := domain.Artwork{
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
			Images:         []domain.Image{image},
		}

		artworks = append(artworks, artwork)
	}

	return artworks
}

func toDomainArtworkCheckoutListRow(rows []generated.ListArtworkStripeDataRow) []domain.Artwork {
	artworks := []domain.Artwork{}

	for _, row := range rows {
		image := domain.Image{
			ID:          row.ImageID,
			ArtworkID:   row.ID,
			IsMainImage: true,
			ImageURL:    row.ImageUrl,
		}

		artwork := domain.Artwork{
			ID:         row.ID,
			Title:      row.Title,
			PriceCents: row.PriceCents,
			Images:     []domain.Image{image},
		}

		artworks = append(artworks, artwork)
	}

	return artworks
}
