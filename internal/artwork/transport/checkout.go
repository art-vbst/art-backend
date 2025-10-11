package transport

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talmage89/art-backend/internal/artwork/repo"
	"github.com/talmage89/art-backend/internal/artwork/service"
	"github.com/talmage89/art-backend/internal/platform/config"
	"github.com/talmage89/art-backend/internal/platform/db/store"
	"github.com/talmage89/art-backend/internal/platform/utils"
)

type CheckoutHandler struct {
	service *service.CheckoutService
}

func NewCheckoutHandler(db *store.Store, config *config.Config) *CheckoutHandler {
	service := service.NewCheckoutService(repo.New(db), config)
	return &CheckoutHandler{service: service}
}

func (h *CheckoutHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.handleCheckout)
	return r
}

type CheckoutRequest struct {
	ArtworkIds []string `json:"artwork_ids"`
}

func (h *CheckoutHandler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.CreateCheckoutSession(r.Context(), req.ArtworkIds)
	if err != nil {
		handleCheckoutServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, result)
}

func handleCheckoutServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidUUIDs):
		utils.RespondError(w, http.StatusBadRequest, "Invalid artwork ID format")
	case errors.Is(err, service.ErrArtworksNotFound):
		utils.RespondError(w, http.StatusNotFound, "One or more artworks not found")
	case errors.Is(err, service.ErrEmptyRequest):
		utils.RespondError(w, http.StatusBadRequest, "Artwork IDs cannot be empty")
	case errors.Is(err, service.ErrTooManyItems):
		utils.RespondError(w, http.StatusBadRequest, "Too many items in cart")
	default:
		log.Printf("checkout error: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create checkout session")
	}
}
