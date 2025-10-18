package service

import (
	"github.com/art-vbst/art-backend/internal/platform/mailer"
	"github.com/google/uuid"
)

type EmailService struct {
	mailer    mailer.Mailer
	signature string
}

func NewEmailService(mailer mailer.Mailer, signature string) *EmailService {
	return &EmailService{mailer: mailer, signature: signature}
}

func (s *EmailService) SendOrderRecieved(orderId uuid.UUID, to string) {
	subject := "Order Recieved!"

	body := "Thank you for your order!\n\n" +
		"I'll get started processing your order. I'll let you when it's shipped, and I'll provide you with a tracking link.\n\n" +
		"If you have any questions or comments feel free to reach out!\n\n" +
		s.signature + "\n\n" +
		"Order ID: " + orderId.String()

	s.mailer.SendEmail(to, subject, body)
}
