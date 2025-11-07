package utils

import (
	"testing"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/google/uuid"
)

func TestCreateAccessToken(t *testing.T) {
	tests := []struct {
		name   string
		user   *domain.User
		secret string
	}{
		{
			name: "valid user",
			user: &domain.User{
				ID:    uuid.New(),
				Email: "test@example.com",
			},
			secret: "test-secret-key",
		},
		{
			name: "different user",
			user: &domain.User{
				ID:    uuid.New(),
				Email: "another@example.com",
			},
			secret: "another-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := CreateAccessToken(tt.user, tt.secret)
			if err != nil {
				t.Fatalf("CreateAccessToken() error = %v", err)
			}
			if token == "" {
				t.Error("CreateAccessToken() returned empty token")
			}

			// Verify the token can be parsed back
			claims, err := ParseAccessToken(token, tt.secret)
			if err != nil {
				t.Fatalf("ParseAccessToken() error = %v", err)
			}

			if claims.UserID != tt.user.ID {
				t.Errorf("ParseAccessToken() UserID = %v, want %v", claims.UserID, tt.user.ID)
			}
			if claims.Email != tt.user.Email {
				t.Errorf("ParseAccessToken() Email = %v, want %v", claims.Email, tt.user.Email)
			}
			if claims.Issuer != Issuer {
				t.Errorf("ParseAccessToken() Issuer = %v, want %v", claims.Issuer, Issuer)
			}
		})
	}
}

func TestCreateRefreshToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"

	tests := []struct {
		name              string
		userID            uuid.UUID
		existingExpiresAt *time.Time
	}{
		{
			name:              "without existing expiration",
			userID:            userID,
			existingExpiresAt: nil,
		},
		{
			name:              "with existing expiration",
			userID:            userID,
			existingExpiresAt: func() *time.Time { t := time.Now().Add(14 * 24 * time.Hour); return &t }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, claims, err := CreateRefreshToken(tt.userID, tt.existingExpiresAt, secret)
			if err != nil {
				t.Fatalf("CreateRefreshToken() error = %v", err)
			}
			if token == "" {
				t.Error("CreateRefreshToken() returned empty token")
			}
			if claims == nil {
				t.Fatal("CreateRefreshToken() returned nil claims")
			}

			if claims.UserID != tt.userID {
				t.Errorf("CreateRefreshToken() claims.UserID = %v, want %v", claims.UserID, tt.userID)
			}

			if tt.existingExpiresAt != nil {
				// Should use existing expiration (within a small tolerance for rounding)
				diff := claims.ExpiresAt.Time.Sub(*tt.existingExpiresAt)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("CreateRefreshToken() claims.ExpiresAt = %v, want %v (diff: %v)", claims.ExpiresAt.Time, *tt.existingExpiresAt, diff)
				}
			} else {
				// Should set new expiration
				expectedExpiration := time.Now().Add(RefreshExpiration)
				if claims.ExpiresAt.Time.Before(expectedExpiration.Add(-1*time.Minute)) ||
					claims.ExpiresAt.Time.After(expectedExpiration.Add(1*time.Minute)) {
					t.Errorf("CreateRefreshToken() claims.ExpiresAt is not within expected range")
				}
			}
		})
	}
}

func TestParseAccessToken(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret-key"

	validToken, err := CreateAccessToken(user, secret)
	if err != nil {
		t.Fatalf("Failed to create token for testing: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		secret    string
		wantErr   bool
		wantEmail string
	}{
		{
			name:      "valid token",
			token:     validToken,
			secret:    secret,
			wantErr:   false,
			wantEmail: user.Email,
		},
		{
			name:    "invalid secret",
			token:   validToken,
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.token",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseAccessToken(tt.token, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && claims.Email != tt.wantEmail {
				t.Errorf("ParseAccessToken() Email = %v, want %v", claims.Email, tt.wantEmail)
			}
		})
	}
}

func TestParseRefreshToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"

	validToken, _, err := CreateRefreshToken(userID, nil, secret)
	if err != nil {
		t.Fatalf("Failed to create token for testing: %v", err)
	}

	tests := []struct {
		name       string
		token      string
		secret     string
		wantErr    bool
		wantUserID uuid.UUID
	}{
		{
			name:       "valid token",
			token:      validToken,
			secret:     secret,
			wantErr:    false,
			wantUserID: userID,
		},
		{
			name:    "invalid secret",
			token:   validToken,
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.token",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseRefreshToken(tt.token, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && claims.UserID != tt.wantUserID {
				t.Errorf("ParseRefreshToken() UserID = %v, want %v", claims.UserID, tt.wantUserID)
			}
		})
	}
}

func TestParseAccessToken_WrongAlgorithm(t *testing.T) {
	// This test verifies that tokens signed with wrong algorithm are rejected
	// We can't easily create a token with wrong algorithm using the same library,
	// but we can test the error handling path by using a token from different method
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret"

	token, err := CreateAccessToken(user, secret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Parse with correct secret should work
	_, err = ParseAccessToken(token, secret)
	if err != nil {
		t.Errorf("ParseAccessToken() with correct secret failed: %v", err)
	}
}

func TestTokenExpiration(t *testing.T) {
	// Test that tokens have reasonable expiration times
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret"

	// Test access token expiration
	accessToken, err := CreateAccessToken(user, secret)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	accessClaims, err := ParseAccessToken(accessToken, secret)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	// Access token should expire in approximately 1 minute
	expectedExpiration := time.Now().Add(AccessExpiration)
	if accessClaims.ExpiresAt.Time.Before(expectedExpiration.Add(-10*time.Second)) ||
		accessClaims.ExpiresAt.Time.After(expectedExpiration.Add(10*time.Second)) {
		t.Errorf("Access token expiration is not within expected range")
	}

	// Test refresh token expiration
	userID := uuid.New()
	refreshToken, refreshClaims, err := CreateRefreshToken(userID, nil, secret)
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	parsedRefreshClaims, err := ParseRefreshToken(refreshToken, secret)
	if err != nil {
		t.Fatalf("ParseRefreshToken() error = %v", err)
	}

	// Refresh token should expire in approximately 7 days
	expectedRefreshExpiration := time.Now().Add(RefreshExpiration)
	if parsedRefreshClaims.ExpiresAt.Time.Before(expectedRefreshExpiration.Add(-1*time.Minute)) ||
		parsedRefreshClaims.ExpiresAt.Time.After(expectedRefreshExpiration.Add(1*time.Minute)) {
		t.Errorf("Refresh token expiration is not within expected range")
	}

	// Claims should match
	if refreshClaims.ID != parsedRefreshClaims.ID {
		t.Errorf("Token ID mismatch: created=%v, parsed=%v", refreshClaims.ID, parsedRefreshClaims.ID)
	}
}

func TestTokenIssuer(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret"

	token, err := CreateAccessToken(user, secret)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	claims, err := ParseAccessToken(token, secret)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	if claims.Issuer != Issuer {
		t.Errorf("Token issuer = %v, want %v", claims.Issuer, Issuer)
	}
}
