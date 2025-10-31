package transport

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo"
	"github.com/art-vbst/art-backend/internal/artwork/service"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
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
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.detail)
	return r
}

func (h *ArtworkHandler) list(w http.ResponseWriter, r *http.Request) {
	artworks, err := h.service.List(r.Context())
	if err != nil {
		handleArtworkServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, artworks)
}

func (h *ArtworkHandler) create(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r); err != nil {
		return
	}

	var body domain.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	artwork, err := h.service.Create(r.Context(), &body)
	if err != nil {
		handleArtworkServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, artwork)
}

func (h *ArtworkHandler) detail(w http.ResponseWriter, r *http.Request) {
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
