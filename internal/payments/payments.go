package payments

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/talmage89/art-backend/internal/platform/config"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
)

type PaymentsHandler struct {
	queries generated.Querier
	config  *config.Config
}

func NewPaymentsHandler(config *config.Config, queries generated.Querier) *PaymentsHandler {
	return &PaymentsHandler{
		config:  config,
		queries: queries,
	}
}

func (h *PaymentsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/checkout", h.handleCheckout)
	return r
}

func (h *PaymentsHandler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	service := NewCheckoutService(h.queries, h.config)
	result, err := service.CreateCheckoutSession(r.Context(), req.ArtworkIds)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidUUIDs):
		respondError(w, http.StatusBadRequest, "Invalid artwork ID format")
	case errors.Is(err, ErrArtworksNotFound):
		respondError(w, http.StatusNotFound, "One or more artworks not found")
	case errors.Is(err, ErrEmptyRequest):
		respondError(w, http.StatusBadRequest, "Artwork IDs cannot be empty")
	case errors.Is(err, ErrTooManyItems):
		respondError(w, http.StatusBadRequest, "Too many items in cart")
	default:
		log.Printf("checkout error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create checkout session")
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
