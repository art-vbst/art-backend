package transport

import (
	"encoding/json"
	"net/http"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/auth/repo"
	"github.com/art-vbst/art-backend/internal/auth/service"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	service *service.AuthService
}

func New(db *store.Store) *AuthHandler {
	repo := repo.New(db)
	service := service.New(repo)
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/me", h.me)
	r.Get("/refresh", h.refresh)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	return r
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
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

	utils.SetAccessCookie(w, data.AccessToken)
	utils.SetRefreshCookie(w, data.RefreshToken)
	utils.RespondJSON(w, http.StatusOK, data.User)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetAccessCookie(w, r)
	if err != nil {
		return
	}

	claims, err := utils.ParseAccessToken(token)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.Logout(r.Context(), claims.UserID); err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.SetAccessCookie(w, "")
	utils.SetRefreshCookie(w, "")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetAccessCookie(w, r)
	if err != nil {
		return
	}

	claims, err := utils.ParseAccessToken(token)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	utils.RespondJSON(w, http.StatusOK, &domain.User{ID: claims.UserID, Email: claims.Email})
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetRefreshCookie(w, r)
	if err != nil {
		return
	}

	data, err := h.service.Refresh(r.Context(), token)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.SetAccessCookie(w, data.AccessToken)
	utils.SetRefreshCookie(w, data.RefreshToken)
	utils.RespondJSON(w, http.StatusOK, data.User)
}

func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	default:
		utils.RespondServerError(w)
	}
}
