package transport

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	artrepo "github.com/art-vbst/art-backend/internal/artwork/repo"
	payrepo "github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/art-vbst/art-backend/internal/payments/service"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
)

type CheckoutHandler struct {
	service *service.CheckoutService
}

func NewCheckoutHandler(db *store.Store, config *config.Config) *CheckoutHandler {
	service := service.NewCheckoutService(artrepo.New(db), payrepo.New(db), config)
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

type CheckoutResponse struct {
	Url string `json:"url"`
}

func (h *CheckoutHandler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	url, err := h.service.CreateCheckoutSession(r.Context(), req.ArtworkIds)
	if err != nil {
		handleCheckoutServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, &CheckoutResponse{
		Url: *url,
	})
}

func handleCheckoutServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidUUIDs):
		utils.RespondError(w, http.StatusBadRequest, "Invalid artwork ID format")
	case errors.Is(err, service.ErrArtworksNotFound):
		utils.RespondError(w, http.StatusNotFound, "One or more artworks not found or unavailable")
	case errors.Is(err, service.ErrEmptyRequest):
		utils.RespondError(w, http.StatusBadRequest, "Artwork IDs cannot be empty")
	case errors.Is(err, service.ErrTooManyItems):
		utils.RespondError(w, http.StatusBadRequest, "Too many items in cart")
	default:
		log.Printf("checkout error: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create checkout session")
	}
}
