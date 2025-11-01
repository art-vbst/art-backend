package postgres

import (
	"context"
	"errors"

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

func (p *Postgres) UpdateArtworksForPendingOrder(ctx context.Context, orderId uuid.UUID, ids []uuid.UUID) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		rows, err := q.UpdateArtworksForOrder(ctx, generated.UpdateArtworksForOrderParams{
			OrderID: pgtype.UUID{Bytes: orderId, Valid: true},
			Column2: ids,
		})
		if len(rows) != len(ids) {
			return errors.New("one or more artworks not found")
		}
		return err
	})
}

func (p *Postgres) UpdateArtworkStatuses(ctx context.Context, orderID uuid.UUID, status domain.ArtworkStatus) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params := generated.UpdateArtworkStatusParams{
			OrderID: pgtype.UUID{Bytes: orderID, Valid: true},
			Status:  status,
		}

		if _, err := q.UpdateArtworkStatus(ctx, params); err != nil {
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
		Paper:          &payload.Paper,
		SortOrder:      payload.SortOrder,
		Status:         payload.Status,
		Medium:         payload.Medium,
		Category:       payload.Category,
	}, nil
}
