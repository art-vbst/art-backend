package transport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/art-vbst/art-backend/internal/payments/service"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/mailer"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrdersHandler struct {
	service *service.OrdersService
	env     *config.Config
}

func NewOrdersHandler(db *store.Store, env *config.Config, mailer mailer.Mailer) *OrdersHandler {
	emails := service.NewEmailService(mailer, env.EmailSignature)
	service := service.NewOrderService(repo.New(db), emails)
	return &OrdersHandler{service: service, env: env}
}

func (h *OrdersHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Get("/{id}", h.detail)
	r.Get("/public/{id}", h.detailPublic)
	r.Put("/{id}", h.updateStatus)
	return r
}

func (h *OrdersHandler) list(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r, h.env.JwtSecret); err != nil {
		return
	}

	statuses, err := parseOrderStatuses(r.URL.Query()["status"])
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid order status provided")
		return
	}

	orders, err := h.service.List(r.Context(), statuses)
	if err != nil {
		handleOrdersServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, orders)
}

func (h *OrdersHandler) detail(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r, h.env.JwtSecret); err != nil {
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "bad order uuid")
		return
	}

	order, err := h.service.Detail(r.Context(), id)
	if err != nil {
		handleOrdersServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, order)
}

type GetOrderPublicParams struct {
	StripeSessionID *string `json:"stripe_session_id"`
}

func (h *OrdersHandler) detailPublic(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "bad order uuid")
		return
	}

	var params GetOrderPublicParams
	stripeSessionID := r.URL.Query().Get("stripe_session_id")
	if stripeSessionID != "" {
		params.StripeSessionID = &stripeSessionID
	}

	order, err := h.service.DetailPublic(r.Context(), id, params.StripeSessionID)
	if err != nil {
		handleOrdersServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, order)
}

type updateStatusPayload struct {
	Status       string  `json:"status"`
	TrackingLink *string `json:"tracking_link"`
}

func (h *OrdersHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r, h.env.JwtSecret); err != nil {
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "bad order uuid")
		return
	}

	var payload updateStatusPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "bad request body")
		return
	}

	parsedStatuses, err := parseOrderStatuses([]string{payload.Status})
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid order status")
		return
	}
	if len(parsedStatuses) == 0 {
		utils.RespondError(w, http.StatusBadRequest, "invalid order status")
		return
	}

	status := parsedStatuses[0]
	if status != domain.OrderStatusShipped && status != domain.OrderStatusCompleted {
		utils.RespondError(w, http.StatusBadRequest, "invalid order status: only 'shipped' or 'completed' are allowed")
		return
	}

	switch status {
	case domain.OrderStatusShipped:
		if err := h.service.MarkAsShipped(r.Context(), id, payload.TrackingLink); err != nil {
			handleOrdersServiceError(w, err)
			return
		}
	case domain.OrderStatusCompleted:
		if err := h.service.MarkAsDelivered(r.Context(), id); err != nil {
			handleOrdersServiceError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleOrdersServiceError(w http.ResponseWriter, err error) {
	switch {
	default:
		log.Printf("orders service error: %v", err)
		utils.RespondServerError(w)
	}
}
