package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/art-vbst/art-backend/internal/platform/config"
)

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func RespondServerError(w http.ResponseWriter) {
	RespondError(w, http.StatusInternalServerError, "an unknown error occurred")
}

const (
	AccessCookieName  = "access_token"
	RefreshCookieName = "refresh_token"
)

type AuthCookieParams struct {
	name   string
	token  string
	path   string
	maxAge int
}

func SetAccessCookie(w http.ResponseWriter, token string) {
	params := &AuthCookieParams{
		name:   AccessCookieName,
		token:  token,
		path:   "/",
		maxAge: int(AccessExpiration.Seconds()),
	}

	if token == "" {
		params.maxAge = -1
	}

	SetAuthCookie(w, params)
}

func SetRefreshCookie(w http.ResponseWriter, token string) {
	params := &AuthCookieParams{
		name:   RefreshCookieName,
		token:  token,
		path:   "/auth/refresh",
		maxAge: int(RefreshExpiration.Seconds()),
	}

	if token == "" {
		params.maxAge = -1
	}

	SetAuthCookie(w, params)
}

func SetAuthCookie(w http.ResponseWriter, params *AuthCookieParams) {
	env := config.Load()

	cookie := &http.Cookie{
		Name:     params.name,
		Value:    params.token,
		Path:     params.path,
		Domain:   env.CookieDomain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   params.maxAge,
	}

	if config.IsDebug() {
		cookie.Secure = false
	}

	http.SetCookie(w, cookie)
}

func GetAccessCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	return GetSessionCookie(w, r, AccessCookieName)
}

func GetRefreshCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	return GetSessionCookie(w, r, RefreshCookieName)
}

func GetSessionCookie(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			RespondError(w, http.StatusUnauthorized, "unauthorized")
			return "", err
		}

		RespondError(w, http.StatusBadRequest, "bad request")
		return "", err
	}

	return cookie.Value, nil
}

func Authenticate(w http.ResponseWriter, r *http.Request) (*AccessClaims, error) {
	token, err := GetAccessCookie(w, r)
	if err != nil {
		return nil, err
	}

	claims, err := ParseAccessToken(token)
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "unauthorized")
		return nil, err
	}

	return claims, nil
}
