package utils

import (
	"testing"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAccessToken(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret-key"

	token, err := CreateAccessToken(user, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseAccessToken_ValidToken(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret-key"

	token, err := CreateAccessToken(user, secret)
	require.NoError(t, err)

	claims, err := ParseAccessToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, Issuer, claims.Issuer)
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"

	token, err := CreateAccessToken(user, secret)
	require.NoError(t, err)

	claims, err := ParseAccessToken(token, wrongSecret)
	assert.Error(t, err)
	assert.Equal(t, ErrBadSignature, err)
	assert.NotNil(t, claims)
}

func TestParseAccessToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	invalidToken := "invalid.token.here"

	claims, err := ParseAccessToken(invalidToken, secret)
	assert.Error(t, err)
	assert.NotNil(t, claims)
}

func TestCreateRefreshToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"

	token, claims, err := CreateRefreshToken(userID, nil, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, Issuer, claims.Issuer)
}

func TestCreateRefreshToken_WithExistingExpiry(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"
	futureTime := time.Now().Add(24 * time.Hour)

	token, claims, err := CreateRefreshToken(userID, &futureTime, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, claims)
	
	// The expiry should be the provided time
	assert.WithinDuration(t, futureTime, claims.ExpiresAt.Time, time.Second)
}

func TestParseRefreshToken_ValidToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"

	token, createdClaims, err := CreateRefreshToken(userID, nil, secret)
	require.NoError(t, err)

	parsedClaims, err := ParseRefreshToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, parsedClaims.UserID)
	assert.Equal(t, createdClaims.ID, parsedClaims.ID)
	assert.Equal(t, Issuer, parsedClaims.Issuer)
}

func TestParseRefreshToken_WrongSecret(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"

	token, _, err := CreateRefreshToken(userID, nil, secret)
	require.NoError(t, err)

	claims, err := ParseRefreshToken(token, wrongSecret)
	assert.Error(t, err)
	assert.Equal(t, ErrBadSignature, err)
	assert.NotNil(t, claims)
}

func TestParseRefreshToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	invalidToken := "invalid.token.here"

	claims, err := ParseRefreshToken(invalidToken, secret)
	assert.Error(t, err)
	assert.NotNil(t, claims)
}

func TestAccessTokenExpiration(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret-key"

	token, err := CreateAccessToken(user, secret)
	require.NoError(t, err)

	claims, err := ParseAccessToken(token, secret)
	require.NoError(t, err)

	// Verify expiration is set correctly
	expectedExpiry := time.Now().Add(AccessExpiration)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 2*time.Second)
}

func TestRefreshTokenExpiration(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key"

	token, claims, err := CreateRefreshToken(userID, nil, secret)
	require.NoError(t, err)

	parsedClaims, err := ParseRefreshToken(token, secret)
	require.NoError(t, err)

	// Verify expiration is set correctly
	expectedExpiry := time.Now().Add(RefreshExpiration)
	assert.WithinDuration(t, expectedExpiry, parsedClaims.ExpiresAt.Time, 2*time.Second)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 2*time.Second)
}
