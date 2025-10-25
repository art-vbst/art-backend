package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	log.Printf("%d %s\n", status, message)
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
	cookie := &http.Cookie{
		Name:     params.name,
		Value:    params.token,
		Path:     params.path,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   params.maxAge,
	}

	if IsDebug() {
		cookie.Secure = false
	}

	http.SetCookie(w, cookie)
}

func GetAccessCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie(AccessCookieName)
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
