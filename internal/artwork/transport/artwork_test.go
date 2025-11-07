package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authdomain "github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These are integration-style tests that test the HTTP handlers.
// They don't mock the database, so they're more like end-to-end tests.
// For production, you'd want to either mock the service layer or use a test database.

func TestArtworkHandler_ListRoute(t *testing.T) {
	// This test just verifies the route is set up correctly
	// In a real scenario, you'd mock the database or use a test database
	env := &config.Config{JwtSecret: "test-secret"}
	handler := &ArtworkHandler{env: env}
	
	router := handler.Routes()
	assert.NotNil(t, router)
}

func TestParseArtworkStatuses_ValidStatuses(t *testing.T) {
	statuses := []string{"available", "sold"}
	result, err := parseArtworkStatuses(statuses)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, domain.ArtworkStatusAvailable)
	assert.Contains(t, result, domain.ArtworkStatusSold)
}

func TestParseArtworkStatuses_InvalidStatus(t *testing.T) {
	statuses := []string{"invalid_status"}
	result, err := parseArtworkStatuses(statuses)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseArtworkStatuses_EmptyList(t *testing.T) {
	statuses := []string{}
	result, err := parseArtworkStatuses(statuses)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestParseArtworkStatuses_MixedValidInvalid(t *testing.T) {
	statuses := []string{"available", "invalid"}
	result, err := parseArtworkStatuses(statuses)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseArtworkStatuses_AllValidStatuses(t *testing.T) {
	statuses := []string{
		"available",
		"sold",
		"not_for_sale",
		"unavailable",
		"coming_soon",
	}
	result, err := parseArtworkStatuses(statuses)
	require.NoError(t, err)
	assert.Len(t, result, 5)
	assert.Contains(t, result, domain.ArtworkStatusAvailable)
	assert.Contains(t, result, domain.ArtworkStatusSold)
	assert.Contains(t, result, domain.ArtworkStatusNotForSale)
	assert.Contains(t, result, domain.ArtworkStatusUnavailable)
	assert.Contains(t, result, domain.ArtworkStatusComingSoon)
}

// Helper to create a test JWT token
func createTestAccessToken(t *testing.T, secret string) string {
	t.Helper()
	user := &authdomain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	
	token, err := utils.CreateAccessToken(user, secret)
	require.NoError(t, err)
	return token
}

// Test authentication middleware behavior
func TestArtworkHandler_CreateRequiresAuth(t *testing.T) {
	env := &config.Config{JwtSecret: "test-secret"}
	// We can't easily test without a real store, but we can test that
	// the route requires authentication by checking for 401 without a token
	
	handler := &ArtworkHandler{env: env}
	router := handler.Routes()
	
	payload := domain.ArtworkPayload{
		Title:        "Test",
		WidthInches:  10,
		HeightInches: 10,
		PriceCents:   1000,
		Status:       domain.ArtworkStatusAvailable,
		Medium:       domain.ArtworkMediumOilPanel,
		Category:     domain.ArtworkCategoryFigure,
	}
	
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should return 401 Unauthorized without a valid token
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestArtworkHandler_UpdateRequiresAuth(t *testing.T) {
	env := &config.Config{JwtSecret: "test-secret"}
	handler := &ArtworkHandler{env: env}
	router := handler.Routes()
	
	artworkID := uuid.New()
	payload := domain.ArtworkPayload{
		Title: "Updated",
	}
	
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPut, "/"+artworkID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	// Set up chi context for URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", artworkID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 401 Unauthorized without a valid token
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestArtworkHandler_DeleteRequiresAuth(t *testing.T) {
	env := &config.Config{JwtSecret: "test-secret"}
	handler := &ArtworkHandler{env: env}
	router := handler.Routes()
	
	artworkID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/"+artworkID.String(), nil)
	
	// Set up chi context for URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", artworkID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return 401 Unauthorized without a valid token
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Note: Testing actual endpoint behavior would require mocking the service layer
// or setting up a test database. The tests above verify the route setup and
// authentication requirements, which is the core transport layer responsibility.
