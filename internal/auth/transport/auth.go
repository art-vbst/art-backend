package transport

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/auth/repo"
	"github.com/art-vbst/art-backend/internal/auth/service"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	service *service.AuthService
	env     *config.Config
}

func New(db *store.Store, env *config.Config) *AuthHandler {
	repo := repo.New(db)
	service := service.New(repo, env)
	return &AuthHandler{service: service, env: env}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/me", h.me)
	r.Post("/refresh", h.refresh)
	r.Post("/logout", h.logout)

	limiter := utils.NewIPRateLimiter(10, time.Minute)
	r.With(limiter.Middleware).Post("/login", h.login)
	r.With(limiter.Middleware).Post("/totp", h.totp)

	return r
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1*utils.MB)
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	data, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	returnData := map[string]string{}
	if data.QRCodeBytes != nil {
		returnData["qr_code"] = base64.StdEncoding.EncodeToString(*data.QRCodeBytes)
	}

	utils.SetTOTPCookie(w, data.TOTPToken, h.env.CookieDomain)
	utils.RespondJSON(w, http.StatusOK, returnData)
}

type TOTPRequest struct {
	PresentedTOTP string `json:"totp"`
}

func (h *AuthHandler) totp(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetTOTPCookie(w, r)
	if err != nil {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1*utils.MB)
	var req TOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	data, err := h.service.ValidateTOTP(r.Context(), token, req.PresentedTOTP)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.SetAccessCookie(w, data.AccessToken, h.env.CookieDomain)
	utils.SetRefreshCookie(w, data.RefreshToken, h.env.CookieDomain)
	utils.RespondJSON(w, http.StatusOK, data.User)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetAccessCookie(w, r)
	if err != nil {
		return
	}

	claims, err := utils.ParseAccessToken(token, h.env.JwtSecret)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	if err := h.service.Logout(r.Context(), claims.UserID); err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.SetAccessCookie(w, "", h.env.CookieDomain)
	utils.SetRefreshCookie(w, "", h.env.CookieDomain)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetAccessCookie(w, r)
	if err != nil {
		return
	}

	claims, err := utils.ParseAccessToken(token, h.env.JwtSecret)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, &domain.User{ID: claims.UserID, Email: claims.Email})
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1*utils.MB)
	token, err := utils.GetRefreshCookie(w, r)
	if err != nil {
		return
	}

	data, err := h.service.Refresh(r.Context(), token)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.SetAccessCookie(w, data.AccessToken, h.env.CookieDomain)
	utils.SetRefreshCookie(w, data.RefreshToken, h.env.CookieDomain)
	utils.RespondJSON(w, http.StatusOK, data.User)
}

func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPassword),
		errors.Is(err, service.ErrInvalidTOTP),
		errors.Is(err, service.ErrTokenMismatch),
		errors.Is(err, service.ErrUserMismatch),
		errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrTokenNotFound),
		errors.Is(err, service.ErrTokenExpired),
		errors.Is(err, service.ErrInvalidToken),
		errors.Is(err, utils.ErrTokenExpired),
		errors.Is(err, utils.ErrInvalidToken),
		errors.Is(err, utils.ErrBadAlgorithm),
		errors.Is(err, utils.ErrBadSignature):
		log.Println(err)
		utils.RespondError(w, http.StatusBadRequest, "bad request")
	default:
		log.Println(err)
		utils.RespondServerError(w)
	}
}
