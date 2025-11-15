package service

import (
	"errors"
	"fmt"

	"github.com/art-vbst/art-backend/internal/platform/mailer"
	"github.com/google/uuid"
)

var (
	ErrEmailSendFailed = errors.New("failed to send email")
)

type EmailService struct {
	mailer    mailer.Mailer
	signature string
}

func NewEmailService(mailer mailer.Mailer, signature string) *EmailService {
	return &EmailService{mailer: mailer, signature: signature}
}

func (s *EmailService) SendOrderReceived(orderID uuid.UUID, to string) error {
	subject := "Order Received!"

	body := "Thank you for your order!\n\n" +
		"Your order is being processed. You'll receive a notification when it ships, along with a tracking link if available.\n\n" +
		"If you have any questions or comments, feel free to reach out!\n\n" +
		s.signature + "\n\n" +
		"Order ID: " + orderID.String()

	if err := s.mailer.SendEmail(to, subject, body); err != nil {
		return fmt.Errorf("send order received email err: %w", ErrEmailSendFailed)
	}

	return nil
}

func (s *EmailService) SendOrderShipped(orderID uuid.UUID, to string, trackingLink *string) error {
	subject := "Order Shipped!"

	body := "Great news! Your order has been shipped.\n\n"

	if trackingLink != nil && *trackingLink != "" {
		body += "You can track your package using the following link:\n" +
			*trackingLink + "\n\n"
	}

	body += "If you have any questions or concerns, please don't hesitate to reach out!\n\n" +
		s.signature + "\n\n" +
		"Order ID: " + orderID.String()

	if err := s.mailer.SendEmail(to, subject, body); err != nil {
		return fmt.Errorf("send order shipped email err: %w", ErrEmailSendFailed)
	}

	return nil
}

func (s *EmailService) SendOrderDelivered(orderID uuid.UUID, to string) error {
	subject := "Order Delivered!"

	body := "Your order has been delivered!\n\n" +
		"I hope you enjoy your purchase. If you have any feedback or questions, I'd love to hear from you.\n\n" +
		"Thank you for your support!\n\n" +
		s.signature + "\n\n" +
		"Order ID: " + orderID.String()

	if err := s.mailer.SendEmail(to, subject, body); err != nil {
		return fmt.Errorf("send order delivered email err: %w", ErrEmailSendFailed)
	}

	return nil
}
