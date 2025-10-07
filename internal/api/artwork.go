package api

import (
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/talmage89/art-backend/internal/db"
)

func NewArtworkHandler(queries db.Querier) *ArtworkHandler {
	return &ArtworkHandler{queries: queries}
}

type ArtworkHandler struct {
	queries db.Querier
}

func (h *ArtworkHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Detail)

	return r
}

type Image struct {
	ImageID        pgtype.UUID `json:"image_id"`
	ImageURL       string      `json:"image_url"`
	ImageWidth     int32       `json:"image_width"`
	ImageHeight    int32       `json:"image_height"`
	ImageCreatedAt time.Time   `json:"image_created_at"`
}

type ArtworkListResponse struct {
	db.ListArtworksRow
	Images []Image `json:"images"`
}

func (h *ArtworkHandler) List(w http.ResponseWriter, r *http.Request) {
	artworks, err := h.queries.ListArtworks(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch artworks", http.StatusInternalServerError)
		return
	}

	imagesByArtwork := make(map[pgtype.UUID][]Image)

	for _, row := range artworks {
		if row.ImageID.Valid {
			image := Image{
				ImageID:        row.ImageID,
				ImageURL:       row.ImageUrl,
				ImageWidth:     *row.ImageWidth,
				ImageHeight:    *row.ImageHeight,
				ImageCreatedAt: row.ImageCreatedAt.Time,
			}
			imagesByArtwork[row.ID] = append(imagesByArtwork[row.ID], image)
		}
	}

	seen := make(map[pgtype.UUID]bool)
	response := make([]ArtworkListResponse, 0)

	for _, row := range artworks {
		if !seen[row.ID] {
			seen[row.ID] = true
			response = append(response, ArtworkListResponse{
				ListArtworksRow: row,
				Images:          imagesByArtwork[row.ID],
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ArtworkDetailResponse struct {
	db.GetArtworkWithImagesRow
	Images []Image `json:"images"`
}

func (h *ArtworkHandler) Detail(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	parsedUUID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	id := pgtype.UUID{
		Bytes: parsedUUID,
		Valid: true,
	}

	rows, err := h.queries.GetArtworkWithImages(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get Artworks", http.StatusInternalServerError)
		return
	}

	if len(rows) < 1 {
		http.Error(w, "Artwork not found", http.StatusNotFound)
		return
	}

	var images []Image

	for _, row := range rows {
		if !row.ImageID.Valid {
			continue
		}

		image := Image{
			ImageID:        row.ImageID,
			ImageURL:       *row.ImageUrl,
			ImageWidth:     *row.ImageWidth,
			ImageHeight:    *row.ImageHeight,
			ImageCreatedAt: row.ImageCreatedAt.Time,
		}

		images = append(images, image)
	}

	response := ArtworkDetailResponse{
		GetArtworkWithImagesRow: rows[0],
		Images:                  images,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ArtworkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string  `json:"title"`
		Description *string `json:"description"`
		Medium      *string `json:"medium"`
		Year        *int32  `json:"year"`
		ImageURL    string  `json:"image_url"`
		ImageWidth  *int32  `json:"image_width"`
		ImageHeight *int32  `json:"image_height"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	artwork, err := h.queries.CreateArtwork(r.Context(), db.CreateArtworkParams{
		Title:          req.Title,
		PaintingNumber: nil,
		PaintingYear:   req.Year,
		WidthInches:    pgtype.Numeric{Int: big.NewInt(875), Exp: -2, Valid: true},
		HeightInches:   pgtype.Numeric{Int: big.NewInt(875), Exp: -2, Valid: true},
		PriceCents:     10_000,
		Paper:          nil,
		SortOrder:      nil,
		SoldAt:         pgtype.Timestamp{},
		Status:         db.ArtworkStatusAvailable,
		Medium:         db.ArtworkMediumAcrylicPanel,
		Category:       db.ArtworkCategoryFigure,
	})
	if err != nil {
		http.Error(w, "Failed to create artwork", http.StatusInternalServerError)
		return
	}

	image, err := h.queries.CreateImage(r.Context(), db.CreateImageParams{
		ArtworkID:   artwork.ID,
		ImageUrl:    req.ImageURL,
		ImageWidth:  req.ImageWidth,
		ImageHeight: req.ImageHeight,
	})
	if err != nil {
		http.Error(w, "Failed to create image", http.StatusInternalServerError)
		return
	}

	response := struct {
		Artwork db.Artwork `json:"artwork"`
		Image   db.Image   `json:"image"`
	}{
		Artwork: artwork,
		Image:   image,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
