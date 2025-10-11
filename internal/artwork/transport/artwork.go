package transport

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talmage89/art-backend/internal/artwork/repo"
	"github.com/talmage89/art-backend/internal/artwork/service"
	"github.com/talmage89/art-backend/internal/platform/db/store"
	"github.com/talmage89/art-backend/internal/platform/utils"
)

type ArtworkHandler struct {
	service *service.ArtworkService
}

func NewArtworkHandler(db *store.Store) *ArtworkHandler {
	service := service.NewArtworkService(repo.New(db))
	return &ArtworkHandler{service: service}
}

func (h *ArtworkHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.Detail)
	return r
}

func (h *ArtworkHandler) List(w http.ResponseWriter, r *http.Request) {
	artworks, err := h.service.List(r.Context())
	if err != nil {
		handleArtworkServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, artworks)
}

func (h *ArtworkHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	artwork, err := h.service.Detail(r.Context(), id)
	if err != nil {
		handleArtworkServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, artwork)
}

func handleArtworkServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrArtworkNotFound):
		utils.RespondError(w, http.StatusNotFound, "Artwork not found")
	case errors.Is(err, service.ErrInvalidArtowrkUUID):
		utils.RespondError(w, http.StatusNotFound, "Invalid artwork ID format")
	default:
		log.Printf("artwork service error: %v", err)
		utils.RespondServerError(w)
	}
}
