package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockMailer is a mock implementation of mailer.Mailer
type mockMailer struct {
	sendEmailFunc func(to, subject, body string) error
}

func (m *mockMailer) SendEmail(to, subject, body string) error {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(to, subject, body)
	}
	return nil
}

func TestNewEmailService(t *testing.T) {
	mailer := &mockMailer{}
	signature := "Best regards,\nTest"
	
	service := NewEmailService(mailer, signature)

	if service == nil {
		t.Fatal("NewEmailService() returned nil")
	}
	if service.mailer == nil {
		t.Error("NewEmailService() service.mailer is nil")
	}
	if service.signature != signature {
		t.Errorf("NewEmailService() signature = %v, want %v", service.signature, signature)
	}
}

func TestEmailService_SendOrderReceived(t *testing.T) {
	orderID := uuid.New()
	email := "customer@example.com"
	signature := "Best regards,\nArt Store"

	tests := []struct {
		name     string
		orderID  uuid.UUID
		to       string
		mockFunc func(to, subject, body string) error
		wantErr  bool
	}{
		{
			name:    "successful send",
			orderID: orderID,
			to:      email,
			mockFunc: func(to, subject, body string) error {
				if to != email {
					t.Errorf("SendEmail called with wrong to: %v, want %v", to, email)
				}
				if subject != "Order Received!" {
					t.Errorf("SendEmail called with wrong subject: %v", subject)
				}
				if !strings.Contains(body, orderID.String()) {
					t.Error("Email body should contain order ID")
				}
				if !strings.Contains(body, signature) {
					t.Error("Email body should contain signature")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:    "mailer error",
			orderID: orderID,
			to:      email,
			mockFunc: func(to, subject, body string) error {
				return errors.New("SMTP error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &mockMailer{
				sendEmailFunc: tt.mockFunc,
			}
			service := NewEmailService(mailer, signature)

			err := service.SendOrderReceived(tt.orderID, tt.to)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendOrderReceived() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr && !errors.Is(err, ErrEmailSendFailed) {
				// Check that error is wrapped with ErrEmailSendFailed
				if !strings.Contains(err.Error(), "send order received email err") {
					t.Errorf("SendOrderReceived() error should wrap ErrEmailSendFailed")
				}
			}
		})
	}
}

func TestEmailService_SendOrderShipped(t *testing.T) {
	orderID := uuid.New()
	email := "customer@example.com"
	signature := "Best regards,\nArt Store"
	trackingLink := "https://tracking.example.com/12345"

	tests := []struct {
		name         string
		orderID      uuid.UUID
		to           string
		trackingLink *string
		mockFunc     func(to, subject, body string) error
		wantErr      bool
	}{
		{
			name:         "successful send with tracking link",
			orderID:      orderID,
			to:           email,
			trackingLink: &trackingLink,
			mockFunc: func(to, subject, body string) error {
				if to != email {
					t.Errorf("SendEmail called with wrong to: %v, want %v", to, email)
				}
				if subject != "Order Shipped!" {
					t.Errorf("SendEmail called with wrong subject: %v", subject)
				}
				if !strings.Contains(body, orderID.String()) {
					t.Error("Email body should contain order ID")
				}
				if !strings.Contains(body, trackingLink) {
					t.Error("Email body should contain tracking link")
				}
				if !strings.Contains(body, signature) {
					t.Error("Email body should contain signature")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:         "successful send without tracking link",
			orderID:      orderID,
			to:           email,
			trackingLink: nil,
			mockFunc: func(to, subject, body string) error {
				if strings.Contains(body, "tracking") && strings.Contains(body, "http") {
					t.Error("Email body should not contain tracking link when not provided")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:         "successful send with empty tracking link",
			orderID:      orderID,
			to:           email,
			trackingLink: func() *string { s := ""; return &s }(),
			mockFunc: func(to, subject, body string) error {
				// Empty tracking link should be treated like no tracking link
				return nil
			},
			wantErr: false,
		},
		{
			name:         "mailer error",
			orderID:      orderID,
			to:           email,
			trackingLink: &trackingLink,
			mockFunc: func(to, subject, body string) error {
				return errors.New("SMTP error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &mockMailer{
				sendEmailFunc: tt.mockFunc,
			}
			service := NewEmailService(mailer, signature)

			err := service.SendOrderShipped(tt.orderID, tt.to, tt.trackingLink)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendOrderShipped() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr && !errors.Is(err, ErrEmailSendFailed) {
				if !strings.Contains(err.Error(), "send order shipped email err") {
					t.Errorf("SendOrderShipped() error should wrap ErrEmailSendFailed")
				}
			}
		})
	}
}

func TestEmailService_SendOrderDelivered(t *testing.T) {
	orderID := uuid.New()
	email := "customer@example.com"
	signature := "Best regards,\nArt Store"

	tests := []struct {
		name     string
		orderID  uuid.UUID
		to       string
		mockFunc func(to, subject, body string) error
		wantErr  bool
	}{
		{
			name:    "successful send",
			orderID: orderID,
			to:      email,
			mockFunc: func(to, subject, body string) error {
				if to != email {
					t.Errorf("SendEmail called with wrong to: %v, want %v", to, email)
				}
				if subject != "Order Delivered!" {
					t.Errorf("SendEmail called with wrong subject: %v", subject)
				}
				if !strings.Contains(body, orderID.String()) {
					t.Error("Email body should contain order ID")
				}
				if !strings.Contains(body, signature) {
					t.Error("Email body should contain signature")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:    "mailer error",
			orderID: orderID,
			to:      email,
			mockFunc: func(to, subject, body string) error {
				return errors.New("SMTP error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := &mockMailer{
				sendEmailFunc: tt.mockFunc,
			}
			service := NewEmailService(mailer, signature)

			err := service.SendOrderDelivered(tt.orderID, tt.to)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendOrderDelivered() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr && !errors.Is(err, ErrEmailSendFailed) {
				if !strings.Contains(err.Error(), "send order delivered email err") {
					t.Errorf("SendOrderDelivered() error should wrap ErrEmailSendFailed")
				}
			}
		})
	}
}

func TestEmailService_AllEmailsContainSignature(t *testing.T) {
	// Test that all email types contain the signature
	orderID := uuid.New()
	email := "customer@example.com"
	signature := "Unique Signature 12345"
	trackingLink := "https://example.com/track"

	var capturedBodies []string
	
	mailer := &mockMailer{
		sendEmailFunc: func(to, subject, body string) error {
			capturedBodies = append(capturedBodies, body)
			return nil
		},
	}
	service := NewEmailService(mailer, signature)

	// Send all email types
	service.SendOrderReceived(orderID, email)
	service.SendOrderShipped(orderID, email, &trackingLink)
	service.SendOrderDelivered(orderID, email)

	// Verify all contain signature
	for i, body := range capturedBodies {
		if !strings.Contains(body, signature) {
			t.Errorf("Email %d does not contain signature", i)
		}
	}

	if len(capturedBodies) != 3 {
		t.Errorf("Expected 3 emails, got %d", len(capturedBodies))
	}
}

func TestEmailService_AllEmailsContainOrderID(t *testing.T) {
	// Test that all email types contain the order ID
	orderID := uuid.New()
	email := "customer@example.com"
	signature := "Test Signature"
	trackingLink := "https://example.com/track"

	var capturedBodies []string
	
	mailer := &mockMailer{
		sendEmailFunc: func(to, subject, body string) error {
			capturedBodies = append(capturedBodies, body)
			return nil
		},
	}
	service := NewEmailService(mailer, signature)

	// Send all email types
	service.SendOrderReceived(orderID, email)
	service.SendOrderShipped(orderID, email, &trackingLink)
	service.SendOrderDelivered(orderID, email)

	// Verify all contain order ID
	orderIDStr := orderID.String()
	for i, body := range capturedBodies {
		if !strings.Contains(body, orderIDStr) {
			t.Errorf("Email %d does not contain order ID", i)
		}
	}
}
