package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (p *Postgres) UpdateArtwork(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error) {
	var artwork *domain.Artwork

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params, err := p.toUpdateArtworkParams(id, payload)
		if err != nil {
			return err
		}

		row, err := q.UpdateArtwork(ctx, *params)
		if err != nil {
			return err
		}

		artwork, err = toDomainArtwork(&row)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return artwork, nil
}

func (p *Postgres) UpdateImage(ctx context.Context, id uuid.UUID, isMainImage bool) (*domain.Image, error) {
	var image *domain.Image

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params := p.toUpdateImageParams(id, isMainImage)

		row, err := q.UpdateImage(ctx, *params)
		if err != nil {
			return err
		}

		image = toDomainImage(&row)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return image, nil
}

func (p *Postgres) SetImageAsMain(ctx context.Context, artID, id uuid.UUID) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params := generated.SetMainImageParams{
			ArtworkID: pgtype.UUID{Bytes: artID, Valid: true},
			ID:        id,
		}
		return q.SetMainImage(ctx, params)
	})
}

func (p *Postgres) UpdateArtworksAsPurchased(
	ctx context.Context,
	ids []uuid.UUID,
	orderID uuid.UUID,
	callback func(selectedIDs []uuid.UUID) error,
) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		rows, err := q.SelectArtworksForUpdate(ctx, ids)
		if err != nil {
			return err
		}

		rowIDs := make([]uuid.UUID, len(rows))
		for i, row := range rows {
			rowIDs[i] = row.ID
		}

		if err := callback(rowIDs); err != nil {
			return err
		}

		params := generated.UpdateArtworksAsPurchasedParams{
			Column1: ids,
			OrderID: pgtype.UUID{Bytes: orderID, Valid: true},
		}

		if _, err := q.UpdateArtworksAsPurchased(ctx, params); err != nil {
			return err
		}

		return nil
	})
}

func (p *Postgres) toUpdateArtworkParams(id uuid.UUID, payload *domain.ArtworkPayload) (*generated.UpdateArtworkParams, error) {
	widthInches, err := utils.NumericFromFloat(payload.WidthInches)
	if err != nil {
		return nil, err
	}

	heightInches, err := utils.NumericFromFloat(payload.HeightInches)
	if err != nil {
		return nil, err
	}

	return &generated.UpdateArtworkParams{
		ID:             id,
		Title:          payload.Title,
		PaintingNumber: payload.PaintingNumber,
		PaintingYear:   payload.PaintingYear,
		WidthInches:    widthInches,
		HeightInches:   heightInches,
		PriceCents:     int32(payload.PriceCents),
		Description:    &payload.Description,
		Paper:          &payload.Paper,
		SortOrder:      payload.SortOrder,
		Status:         payload.Status,
		Medium:         payload.Medium,
		Category:       payload.Category,
	}, nil
}

func (p *Postgres) toUpdateImageParams(id uuid.UUID, isMainImage bool) *generated.UpdateImageParams {
	return &generated.UpdateImageParams{
		ID:          id,
		IsMainImage: isMainImage,
	}
}
