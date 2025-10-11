package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/talmage89/art-backend/internal/artwork/domain"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
	"github.com/talmage89/art-backend/internal/platform/db/store"
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

func (p *Postgres) GetArtworkDetail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error) {
	artworkRows, err := p.db.Queries().GetArtworkWithImages(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainArtworkDetailRows(artworkRows), nil
}

func (p *Postgres) GetArtworkCheckoutData(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error) {
	artworks, err := p.db.Queries().GetStripeDataByArtworkIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return toDomainArtworkCheckoutListRow(artworks), nil
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

func toDomainArtworkCheckoutListRow(rows []generated.GetStripeDataByArtworkIDsRow) []domain.Artwork {
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

// func (h *ArtworkHandler) Create(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		Title       string  `json:"title"`
// 		Description *string `json:"description"`
// 		Medium      *string `json:"medium"`
// 		Year        *int32  `json:"year"`
// 		ImageURL    string  `json:"image_url"`
// 		ImageWidth  *int32  `json:"image_width"`
// 		ImageHeight *int32  `json:"image_height"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	artwork, err := h.queries.CreateArtwork(r.Context(), generated.CreateArtworkParams{
// 		Title:          req.Title,
// 		PaintingNumber: nil,
// 		PaintingYear:   req.Year,
// 		WidthInches:    pgtype.Numeric{Int: big.NewInt(875), Exp: -2, Valid: true},
// 		HeightInches:   pgtype.Numeric{Int: big.NewInt(875), Exp: -2, Valid: true},
// 		PriceCents:     10_000,
// 		Paper:          nil,
// 		SortOrder:      nil,
// 		SoldAt:         pgtype.Timestamp{},
// 		Status:         generated.ArtworkStatusAvailable,
// 		Medium:         generated.ArtworkMediumAcrylicPanel,
// 		Category:       generated.ArtworkCategoryFigure,
// 	})
// 	if err != nil {
// 		http.Error(w, "Failed to create artwork", http.StatusInternalServerError)
// 		return
// 	}

// 	image, err := h.queries.CreateImage(r.Context(), generated.CreateImageParams{
// 		ArtworkID:   artwork.ID,
// 		ImageUrl:    req.ImageURL,
// 		ImageWidth:  req.ImageWidth,
// 		ImageHeight: req.ImageHeight,
// 	})
// 	if err != nil {
// 		http.Error(w, "Failed to create image", http.StatusInternalServerError)
// 		return
// 	}

// 	response := struct {
// 		Artwork generated.Artwork `json:"artwork"`
// 		Image   generated.Image   `json:"image"`
// 	}{
// 		Artwork: artwork,
// 		Image:   image,
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(response)
// }
