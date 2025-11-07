package transport

import (
	"log"
	"net/http"

	"github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/art-vbst/art-backend/internal/payments/service"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrdersHandler struct {
	service *service.OrdersService
}

func NewOrdersHandler(db *store.Store) *OrdersHandler {
	service := service.NewOrderService(repo.New(db))
	return &OrdersHandler{service: service}
}

func (h *OrdersHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Get("/{id}", h.detail)
	r.Get("/public/{id}", h.getPublic)
	return r
}

func (h *OrdersHandler) list(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r); err != nil {
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
	if _, err := utils.Authenticate(w, r); err != nil {
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

func (h *OrdersHandler) getPublic(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.service.GetPublic(r.Context(), id, params.StripeSessionID)
	if err != nil {
		handleOrdersServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, order)
}

func handleOrdersServiceError(w http.ResponseWriter, err error) {
	switch {
	default:
		log.Printf("orders service error: %v", err)
		utils.RespondServerError(w)
	}
}
