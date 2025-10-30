package repo

import (
	"context"
	"errors"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Postgres struct {
	db *store.Store
}

func (p *Postgres) ListArtworks(ctx context.Context) ([]domain.Artwork, error) {
	artworks, err := p.db.Queries().ListArtworks(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainArtworkListRow(artworks), nil
}

func (p *Postgres) CreateArtwork(ctx context.Context, body *domain.CreateRequest) (*domain.Artwork, error) {
	var created *domain.Artwork

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		width, err := utils.NumericFromFloat(body.WidthInches)
		if err != nil {
			return err
		}

		height, err := utils.NumericFromFloat(body.HeightInches)
		if err != nil {
			return err
		}

		payload := generated.CreateArtworkParams{
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

		row, err := q.CreateArtwork(ctx, payload)
		if err != nil {
			return err
		}

		created = toDomainArtwork(&row)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return created, nil
}

func (p *Postgres) GetArtworkDetail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
	artworkRows, err := p.db.Queries().GetArtworkWithImages(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainArtworkDetailRows(artworkRows), nil
}

func (p *Postgres) GetArtworkCheckoutData(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error) {
	artworks, err := p.db.Queries().ListArtworkStripeData(ctx, ids)
	if err != nil {
		return nil, err
	}
	return toDomainArtworkCheckoutListRow(artworks), nil
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

func toDomainArtworkListRow(rows []generated.ListArtworksRow) []domain.Artwork {
	artworks := []domain.Artwork{}

	for _, row := range rows {
		widthInches, _ := row.WidthInches.Float64Value()
		heightInches, _ := row.HeightInches.Float64Value()

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
			Images: []domain.Image{
				{
					ID:          row.ImageID,
					ArtworkID:   row.ID,
					IsMainImage: true,
					ImageURL:    row.ImageUrl,
					ImageWidth:  row.ImageWidth,
					ImageHeight: row.ImageHeight,
					CreatedAt:   row.ImageCreatedAt.Time,
				},
			},
		}

		artworks = append(artworks, artwork)
	}

	return artworks
}

func toDomainArtworkCheckoutListRow(rows []generated.ListArtworkStripeDataRow) []domain.Artwork {
	artworks := []domain.Artwork{}

	for _, row := range rows {
		artwork := domain.Artwork{
			ID:         row.ID,
			Title:      row.Title,
			PriceCents: row.PriceCents,
			Images: []domain.Image{
				{
					ID:          row.ImageID,
					ArtworkID:   row.ID,
					IsMainImage: true,
					ImageURL:    row.ImageUrl,
				},
			},
		}

		artworks = append(artworks, artwork)
	}

	return artworks
}

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

func toDomainArtworkDetailRows(rows []generated.GetArtworkWithImagesRow) *domain.Artwork {
	if len(rows) == 0 {
		return nil
	}

	artworkRow := rows[0]

	widthInches, _ := artworkRow.WidthInches.Float64Value()
	heightInches, _ := artworkRow.HeightInches.Float64Value()

	artwork := &domain.Artwork{
		ID:             artworkRow.ID,
		Title:          artworkRow.Title,
		PaintingNumber: artworkRow.PaintingNumber,
		PaintingYear:   artworkRow.PaintingYear,
		WidthInches:    widthInches.Float64,
		HeightInches:   heightInches.Float64,
		PriceCents:     artworkRow.PriceCents,
		Paper:          artworkRow.Paper,
		SortOrder:      artworkRow.SortOrder,
		SoldAt:         &artworkRow.SoldAt.Time,
		Status:         artworkRow.Status,
		Medium:         artworkRow.Medium,
		Category:       artworkRow.Category,
		CreatedAt:      artworkRow.CreatedAt.Time,
	}

	artwork.Images = toDomainImages(rows)

	return artwork
}

func toDomainImages(rows []generated.GetArtworkWithImagesRow) []domain.Image {
	images := []domain.Image{}

	if len(rows) == 0 || !rows[0].ImageID.Valid {
		return images
	}

	for _, row := range rows {
		imageID, _ := uuid.FromBytes(row.ImageID.Bytes[:])

		imageURL := ""
		if row.ImageUrl != nil {
			imageURL = *row.ImageUrl
		}

		isMainImage := false
		if row.IsMainImage != nil {
			isMainImage = *row.IsMainImage
		}

		images = append(images, domain.Image{
			ID:          imageID,
			ArtworkID:   row.ID,
			IsMainImage: isMainImage,
			ImageURL:    imageURL,
			ImageWidth:  row.ImageWidth,
			ImageHeight: row.ImageHeight,
			CreatedAt:   row.ImageCreatedAt.Time,
		})
	}

	return images
}
