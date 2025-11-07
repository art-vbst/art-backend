package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/google/uuid"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		wantStatusCode int
	}{
		{
			name:           "simple map",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "struct data",
			status:         http.StatusCreated,
			data:           domain.User{ID: uuid.New(), Email: "test@example.com"},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "array data",
			status:         http.StatusOK,
			data:           []string{"item1", "item2"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "nil data",
			status:         http.StatusNoContent,
			data:           nil,
			wantStatusCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			RespondJSON(w, tt.status, tt.data)

			if w.Code != tt.wantStatusCode {
				t.Errorf("RespondJSON() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("RespondJSON() Content-Type = %v, want application/json", contentType)
			}

			// Verify JSON can be decoded (if data is not nil)
			if tt.data != nil {
				var result map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
					// Try decoding as array
					var arrResult []interface{}
					w2 := httptest.NewRecorder()
					RespondJSON(w2, tt.status, tt.data)
					if err2 := json.NewDecoder(w2.Body).Decode(&arrResult); err2 != nil {
						// Some types might not decode to map or array, that's ok
						// Just verify it's valid JSON
						if !json.Valid(w.Body.Bytes()) {
							t.Errorf("RespondJSON() produced invalid JSON: %v", err)
						}
					}
				}
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		wantStatusCode int
		wantMessage    string
	}{
		{
			name:           "bad request",
			status:         http.StatusBadRequest,
			message:        "Invalid input",
			wantStatusCode: http.StatusBadRequest,
			wantMessage:    "Invalid input",
		},
		{
			name:           "unauthorized",
			status:         http.StatusUnauthorized,
			message:        "unauthorized",
			wantStatusCode: http.StatusUnauthorized,
			wantMessage:    "unauthorized",
		},
		{
			name:           "not found",
			status:         http.StatusNotFound,
			message:        "Resource not found",
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "Resource not found",
		},
		{
			name:           "internal server error",
			status:         http.StatusInternalServerError,
			message:        "Something went wrong",
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			RespondError(w, tt.status, tt.message)

			if w.Code != tt.wantStatusCode {
				t.Errorf("RespondError() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("RespondError() Content-Type = %v, want application/json", contentType)
			}

			var result map[string]string
			if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if result["error"] != tt.wantMessage {
				t.Errorf("RespondError() message = %v, want %v", result["error"], tt.wantMessage)
			}
		})
	}
}

func TestRespondServerError(t *testing.T) {
	w := httptest.NewRecorder()

	RespondServerError(w)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("RespondServerError() status = %v, want %v", w.Code, http.StatusInternalServerError)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if result["error"] != "an unknown error occurred" {
		t.Errorf("RespondServerError() message = %v, want 'an unknown error occurred'", result["error"])
	}
}

func TestSetAccessCookie(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		domain   string
		wantName string
	}{
		{
			name:     "with token",
			token:    "test-access-token",
			domain:   "example.com",
			wantName: AccessCookieName,
		},
		{
			name:     "empty token for deletion",
			token:    "",
			domain:   "example.com",
			wantName: AccessCookieName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			SetAccessCookie(w, tt.token, tt.domain)

			cookies := w.Result().Cookies()
			if len(cookies) == 0 {
				t.Fatal("SetAccessCookie() did not set a cookie")
			}

			cookie := cookies[0]
			if cookie.Name != tt.wantName {
				t.Errorf("SetAccessCookie() cookie name = %v, want %v", cookie.Name, tt.wantName)
			}

			if cookie.Value != tt.token {
				t.Errorf("SetAccessCookie() cookie value = %v, want %v", cookie.Value, tt.token)
			}

			if cookie.Path != "/" {
				t.Errorf("SetAccessCookie() cookie path = %v, want /", cookie.Path)
			}

			if cookie.Domain != tt.domain {
				t.Errorf("SetAccessCookie() cookie domain = %v, want %v", cookie.Domain, tt.domain)
			}

			if !cookie.HttpOnly {
				t.Error("SetAccessCookie() cookie should be HttpOnly")
			}

			if cookie.SameSite != http.SameSiteStrictMode {
				t.Errorf("SetAccessCookie() cookie SameSite = %v, want %v", cookie.SameSite, http.SameSiteStrictMode)
			}

			if tt.token == "" {
				if cookie.MaxAge != -1 {
					t.Errorf("SetAccessCookie() with empty token should set MaxAge = -1, got %v", cookie.MaxAge)
				}
			} else {
				if cookie.MaxAge <= 0 {
					t.Errorf("SetAccessCookie() with token should set positive MaxAge, got %v", cookie.MaxAge)
				}
			}
		})
	}
}

func TestSetRefreshCookie(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		domain   string
		wantName string
	}{
		{
			name:     "with token",
			token:    "test-refresh-token",
			domain:   "example.com",
			wantName: RefreshCookieName,
		},
		{
			name:     "empty token for deletion",
			token:    "",
			domain:   "example.com",
			wantName: RefreshCookieName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			SetRefreshCookie(w, tt.token, tt.domain)

			cookies := w.Result().Cookies()
			if len(cookies) == 0 {
				t.Fatal("SetRefreshCookie() did not set a cookie")
			}

			cookie := cookies[0]
			if cookie.Name != tt.wantName {
				t.Errorf("SetRefreshCookie() cookie name = %v, want %v", cookie.Name, tt.wantName)
			}

			if cookie.Value != tt.token {
				t.Errorf("SetRefreshCookie() cookie value = %v, want %v", cookie.Value, tt.token)
			}

			if cookie.Path != "/auth/refresh" {
				t.Errorf("SetRefreshCookie() cookie path = %v, want /auth/refresh", cookie.Path)
			}

			if cookie.Domain != tt.domain {
				t.Errorf("SetRefreshCookie() cookie domain = %v, want %v", cookie.Domain, tt.domain)
			}

			if !cookie.HttpOnly {
				t.Error("SetRefreshCookie() cookie should be HttpOnly")
			}

			if tt.token == "" {
				if cookie.MaxAge != -1 {
					t.Errorf("SetRefreshCookie() with empty token should set MaxAge = -1, got %v", cookie.MaxAge)
				}
			} else {
				if cookie.MaxAge <= 0 {
					t.Errorf("SetRefreshCookie() with token should set positive MaxAge, got %v", cookie.MaxAge)
				}
			}
		})
	}
}

func TestGetAccessCookie(t *testing.T) {
	tests := []struct {
		name      string
		setCookie bool
		cookieVal string
		wantErr   bool
	}{
		{
			name:      "cookie present",
			setCookie: true,
			cookieVal: "test-token",
			wantErr:   false,
		},
		{
			name:      "cookie missing",
			setCookie: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  AccessCookieName,
					Value: tt.cookieVal,
				})
			}

			w := httptest.NewRecorder()
			got, err := GetAccessCookie(w, req)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccessCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.cookieVal {
				t.Errorf("GetAccessCookie() = %v, want %v", got, tt.cookieVal)
			}
		})
	}
}

func TestGetRefreshCookie(t *testing.T) {
	tests := []struct {
		name      string
		setCookie bool
		cookieVal string
		wantErr   bool
	}{
		{
			name:      "cookie present",
			setCookie: true,
			cookieVal: "test-refresh-token",
			wantErr:   false,
		},
		{
			name:      "cookie missing",
			setCookie: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  RefreshCookieName,
					Value: tt.cookieVal,
				})
			}

			w := httptest.NewRecorder()
			got, err := GetRefreshCookie(w, req)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRefreshCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.cookieVal {
				t.Errorf("GetRefreshCookie() = %v, want %v", got, tt.cookieVal)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	secret := "test-secret-key"

	validToken, err := CreateAccessToken(user, secret)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	tests := []struct {
		name      string
		setCookie bool
		token     string
		secret    string
		wantErr   bool
		wantEmail string
	}{
		{
			name:      "valid token",
			setCookie: true,
			token:     validToken,
			secret:    secret,
			wantErr:   false,
			wantEmail: user.Email,
		},
		{
			name:      "missing cookie",
			setCookie: false,
			token:     "",
			secret:    secret,
			wantErr:   true,
		},
		{
			name:      "invalid token",
			setCookie: true,
			token:     "invalid-token",
			secret:    secret,
			wantErr:   true,
		},
		{
			name:      "wrong secret",
			setCookie: true,
			token:     validToken,
			secret:    "wrong-secret",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  AccessCookieName,
					Value: tt.token,
				})
			}

			w := httptest.NewRecorder()
			claims, err := Authenticate(w, req, tt.secret)

			if (err != nil) != tt.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims == nil {
					t.Fatal("Authenticate() returned nil claims")
				}
				if claims.Email != tt.wantEmail {
					t.Errorf("Authenticate() email = %v, want %v", claims.Email, tt.wantEmail)
				}
			}
		})
	}
}
