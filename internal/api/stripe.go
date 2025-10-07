package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/talmage89/art-backend/internal/config"
	"github.com/talmage89/art-backend/internal/db"
)

func NewStripeHandler(config *config.Config, queries db.Querier) *StripeHandler {
	return &StripeHandler{config: config, queries: queries}
}

type StripeHandler struct {
	config  *config.Config
	queries db.Querier
}

func (h *StripeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/checkout", h.handleCheckoutSessionCreate)
	return r
}

func UUIDToString(id pgtype.UUID) (string, error) {
	if !id.Valid {
		return "", fmt.Errorf("invalid UUID")
	}
	u, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (h *StripeHandler) handleCheckoutSessionCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ArtworkIds []string `json:"artwork_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ids := []pgtype.UUID{}

	for _, idString := range req.ArtworkIds {
		parsedUUID, err := uuid.Parse(idString)
		if err != nil {
			http.Error(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		id := pgtype.UUID{
			Bytes: parsedUUID,
			Valid: true,
		}

		ids = append(ids, id)
	}

	rows, err := h.queries.GetStripeDataByArtworkIDs(r.Context(), ids)

	if err != nil {
		http.Error(w, "Error retriving artworks", http.StatusInternalServerError)
		return
	}

	if len(rows) != len(req.ArtworkIds) {
		http.Error(w, "Artworks not found", http.StatusNotFound)
		return
	}

	stringifiedIdsBytes, err := json.Marshal(ids)
	if err != nil {
		http.Error(w, "An unknown error occured", http.StatusInternalServerError)
	}

	lineItems := []*stripe.CheckoutSessionLineItemParams{}

	for _, row := range rows {
		productData := stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name:   stripe.String(row.Title),
			Images: stripe.StringSlice([]string{row.ImageUrl}),
		}

		priceData := stripe.CheckoutSessionLineItemPriceDataParams{
			Currency:    stripe.String("usd"),
			UnitAmount:  stripe.Int64(int64(row.PriceCents)),
			ProductData: &productData,
		}

		lineItem := stripe.CheckoutSessionLineItemParams{
			PriceData: &priceData,
			Quantity:  stripe.Int64(1),
		}

		lineItems = append(lineItems, &lineItem)
	}

	stripe.Key = h.config.StripeSecretKey
	domain := h.config.FrontendUrl
	params := &stripe.CheckoutSessionParams{
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "?success=true"),
		CancelURL:  stripe.String(domain + "?canceled=true"),
		Metadata: map[string]string{
			"artwork_ids": string(stringifiedIdsBytes),
		},
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("session.New: %v", err)
		http.Error(w, "Error creating checkout session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": s.URL,
	})
}
