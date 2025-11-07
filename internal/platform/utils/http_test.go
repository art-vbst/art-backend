package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{
		"message": "hello",
		"status":  "ok",
	}
	
	RespondJSON(w, http.StatusOK, data)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "hello", response["message"])
	assert.Equal(t, "ok", response["status"])
}

func TestRespondJSON_WithStruct(t *testing.T) {
	w := httptest.NewRecorder()
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	
	RespondJSON(w, http.StatusCreated, user)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var response domain.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Email, response.Email)
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	
	RespondError(w, http.StatusBadRequest, "invalid input")
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid input", response["error"])
}

func TestRespondServerError(t *testing.T) {
	w := httptest.NewRecorder()
	
	RespondServerError(w)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "an unknown error occurred", response["error"])
}

func TestSetAccessCookie(t *testing.T) {
	w := httptest.NewRecorder()
	token := "test-access-token"
	domain := "example.com"
	
	SetAccessCookie(w, token, domain)
	
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	
	cookie := cookies[0]
	assert.Equal(t, AccessCookieName, cookie.Name)
	assert.Equal(t, token, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, domain, cookie.Domain)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Greater(t, cookie.MaxAge, 0)
}

func TestSetAccessCookie_EmptyToken(t *testing.T) {
	w := httptest.NewRecorder()
	
	SetAccessCookie(w, "", "example.com")
	
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	
	cookie := cookies[0]
	assert.Equal(t, AccessCookieName, cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge) // Should delete the cookie
}

func TestSetRefreshCookie(t *testing.T) {
	w := httptest.NewRecorder()
	token := "test-refresh-token"
	domain := "example.com"
	
	SetRefreshCookie(w, token, domain)
	
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	
	cookie := cookies[0]
	assert.Equal(t, RefreshCookieName, cookie.Name)
	assert.Equal(t, token, cookie.Value)
	assert.Equal(t, "/auth/refresh", cookie.Path)
	assert.Equal(t, domain, cookie.Domain)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Greater(t, cookie.MaxAge, 0)
}

func TestSetRefreshCookie_EmptyToken(t *testing.T) {
	w := httptest.NewRecorder()
	
	SetRefreshCookie(w, "", "example.com")
	
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	
	cookie := cookies[0]
	assert.Equal(t, RefreshCookieName, cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge) // Should delete the cookie
}

func TestGetSessionCookie_Success(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "test-cookie",
		Value: "test-value",
	})
	
	value, err := GetSessionCookie(w, r, "test-cookie")
	require.NoError(t, err)
	assert.Equal(t, "test-value", value)
}

func TestGetSessionCookie_Missing(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	
	value, err := GetSessionCookie(w, r, "missing-cookie")
	assert.Error(t, err)
	assert.Empty(t, value)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetAccessCookie_Success(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  AccessCookieName,
		Value: "test-access-token",
	})
	
	value, err := GetAccessCookie(w, r)
	require.NoError(t, err)
	assert.Equal(t, "test-access-token", value)
}

func TestGetRefreshCookie_Success(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  RefreshCookieName,
		Value: "test-refresh-token",
	})
	
	value, err := GetRefreshCookie(w, r)
	require.NoError(t, err)
	assert.Equal(t, "test-refresh-token", value)
}

func TestAuthenticate_ValidToken(t *testing.T) {
	secret := "test-secret"
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	
	token, err := CreateAccessToken(user, secret)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  AccessCookieName,
		Value: token,
	})
	
	claims, err := Authenticate(w, r, secret)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
}

func TestAuthenticate_MissingCookie(t *testing.T) {
	secret := "test-secret"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	
	claims, err := Authenticate(w, r, secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	secret := "test-secret"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  AccessCookieName,
		Value: "invalid-token",
	})
	
	claims, err := Authenticate(w, r, secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_WrongSecret(t *testing.T) {
	secret := "test-secret"
	wrongSecret := "wrong-secret"
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	
	token, err := CreateAccessToken(user, secret)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  AccessCookieName,
		Value: token,
	})
	
	claims, err := Authenticate(w, r, wrongSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
