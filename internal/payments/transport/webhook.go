package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	artrepo "github.com/art-vbst/art-backend/internal/artwork/repo"
	payrepo "github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/art-vbst/art-backend/internal/payments/service"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/mailer"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"
)

type WebhookHandler struct {
	service *service.WebhookService
	env     *config.Config
}

const (
	CheckoutComplete stripe.EventType = "checkout.session.completed"
	CheckoutExpired  stripe.EventType = "checkout.session.expired"
)

func NewWebhookHandler(db *store.Store, env *config.Config, mailer mailer.Mailer) *WebhookHandler {
	emails := service.NewEmailService(mailer, env.EmailSignature)
	service := service.NewWebhookService(payrepo.New(db), artrepo.New(db), emails, env)
	return &WebhookHandler{service: service, env: env}
}

func (h *WebhookHandler) Routes() *chi.Mux {
	r := chi.NewRouter()
	// Per-IP rate limit webhook hits
	limiter := utils.NewIPRateLimiter(60, time.Minute)
	r.With(limiter.Middleware).Post("/", h.post)
	return r
}

func (h *WebhookHandler) post(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1*utils.MB)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to read request body")
		return
	}
	defer r.Body.Close()

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.env.StripeWebhookSecret)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "failed to verify webhook signature")
		return
	}

	if event.Type != CheckoutComplete && event.Type != CheckoutExpired {
		w.WriteHeader(http.StatusOK)
		return
	}

	session, err := parseCheckoutSession(event)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "error parsing webhook JSON to checkout session")
		return
	}

	switch event.Type {
	case CheckoutComplete:
		if err := h.service.HandleCheckoutComplete(r.Context(), session); err != nil {
			handleServiceError(w, err)
			return
		}
	case CheckoutExpired:
		if err := h.service.HandleCheckoutExpired(r.Context(), session); err != nil {
			handleServiceError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func parseCheckoutSession(event stripe.Event) (*stripe.CheckoutSession, error) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return nil, fmt.Errorf("unmarshal checkout session err: %w", err)
	}

	return &session, nil
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrIntentAlreadyProcessed):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrEmailSendFailed):
		log.Printf("email send failed: %v", err)
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrOrderNotFound):
		utils.RespondError(w, http.StatusNotFound, "Order not found")
	case errors.Is(err, service.ErrArtworksNotAvailable):
		utils.RespondError(w, http.StatusNotFound, "One or more artworks is not available for purchase")
	case errors.Is(err, service.ErrMetadataParse):
		utils.RespondError(w, http.StatusInternalServerError, "Metadata parse error")
	case errors.Is(err, service.ErrBadIntentStatus):
		utils.RespondError(w, http.StatusInternalServerError, "Unexpected payment intent status")
	case errors.Is(err, service.ErrIntentProcessingFailure):
		utils.RespondError(w, http.StatusInternalServerError, "Failed to process payment intent")
	default:
		log.Printf("webhook service error: %v", err)
		utils.RespondServerError(w)
	}
}
