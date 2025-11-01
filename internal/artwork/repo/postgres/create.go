package postgres

import (
	"context"
	"log"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/jackc/pgx/v5/pgtype"
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

func (p *Postgres) CreateImage(ctx context.Context, data *domain.CreateImagePayload) (*domain.Image, error) {
	var image *domain.Image

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params := p.toCreateImageParams(data)

		row, err := q.CreateImage(ctx, *params)
		if err != nil {
			log.Print("create image", err)
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

func (p *Postgres) toCreateImageParams(data *domain.CreateImagePayload) *generated.CreateImageParams {
	return &generated.CreateImageParams{
		ArtworkID:   pgtype.UUID{Bytes: data.ArtworkID, Valid: true},
		ImageUrl:    data.ImageURL,
		IsMainImage: data.IsMainImage,
		ImageWidth:  data.ImageWidth,
		ImageHeight: data.ImageHeight,
	}
}
