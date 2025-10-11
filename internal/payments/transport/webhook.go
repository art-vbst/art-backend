package transport

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"
	"github.com/talmage89/art-backend/internal/payments/repo"
	"github.com/talmage89/art-backend/internal/payments/service"
	"github.com/talmage89/art-backend/internal/platform/config"
	"github.com/talmage89/art-backend/internal/platform/db/store"
	"github.com/talmage89/art-backend/internal/platform/utils"
)

type WebhookHandler struct {
	service *service.WebhookService
	env     *config.Config
}

func NewWebhookHandler(db *store.Store, env *config.Config) *WebhookHandler {
	service := service.NewWebhookService(repo.New(db))
	return &WebhookHandler{service: service, env: env}
}

func (h *WebhookHandler) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.post)
	return r
}

func (h *WebhookHandler) post(w http.ResponseWriter, r *http.Request) {
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

	if event.Type != "checkout.session.completed" {
		w.WriteHeader(http.StatusOK)
		return
	}

	session, err := parseCheckoutSession(event)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "error parsing webhook JSON to checkout session")
		return
	}

	h.service.HandleCheckoutComplete(r.Context(), session)
	w.WriteHeader(http.StatusOK)
}

func parseCheckoutSession(event stripe.Event) (*stripe.CheckoutSession, error) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return nil, err
	}

	return &session, nil
}
