package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/art-vbst/art-backend/internal/platform/utils"
)

func (p *Postgres) CreateArtwork(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error) {
	var created *domain.Artwork

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params, err := p.toCreateArtworkParams(body)
		if err != nil {
			return err
		}

		row, err := q.CreateArtwork(ctx, *params)
		if err != nil {
			return err
		}

		created, err = toDomainArtwork(&row)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return created, nil
}

func (p *Postgres) toCreateArtworkParams(body *domain.ArtworkPayload) (*generated.CreateArtworkParams, error) {
	width, err := utils.NumericFromFloat(body.WidthInches)
	if err != nil {
		return nil, err
	}

	height, err := utils.NumericFromFloat(body.HeightInches)
	if err != nil {
		return nil, err
	}

	params := generated.CreateArtworkParams{
		Title:          body.Title,
		PaintingNumber: body.PaintingNumber,
		PaintingYear:   body.PaintingYear,
		WidthInches:    width,
		HeightInches:   height,
		PriceCents:     int32(body.PriceCents),
		Paper:          &body.Paper,
		Status:         body.Status,
		Medium:         body.Medium,
		Category:       body.Category,
	}

	return &params, nil
}
