package payments

import (
	"encoding/json"
	"errors"

	"github.com/stripe/stripe-go/v83"
	"github.com/talmage89/art-backend/internal/config"
	"github.com/talmage89/art-backend/internal/db"
)

var (
	ErrInvalidRequestBody     = errors.New("invalid request body")
	ErrFailedBodyUnmarshaling = errors.New("request body unmarshalling failed")
)

type WebhookService struct {
	queries db.Querier
	config  *config.Config
}

func NewWebhokService(queries db.Querier, config *config.Config) *WebhookService {
	return &WebhookService{
		queries: queries,
		config:  config,
	}
}

func (s *WebhookService) Webhook(event stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return ErrFailedBodyUnmarshaling
		}
		s.handleCheckoutSessionCompleted(&session)
	}

	return nil

	// // Unmarshal the event data into an appropriate struct depending on its Type
	// switch event.Type {
	// case "payment_intent.succeeded":
	// 	var paymentIntent stripe.PaymentIntent
	// 	err := json.Unmarshal(event.Data.Raw, &paymentIntent)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	// Then define and call a func to handle the successful payment intent.
	// 	// handlePaymentIntentSucceeded(paymentIntent)
	// case "payment_method.attached":
	// 	var paymentMethod stripe.PaymentMethod
	// 	err := json.Unmarshal(event.Data.Raw, &paymentMethod)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	// Then define and call a func to handle the successful attachment of a PaymentMethod.
	// 	// handlePaymentMethodAttached(paymentMethod)
	// // ... handle other event types
	// default:
	// 	fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	// }
}

func (s *WebhookService) handleCheckoutSessionCompleted(session *stripe.CheckoutSession) {
	// update
	// s.queries.CreateOrder()
}
