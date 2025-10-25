package transport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/art-vbst/art-backend/internal/auth/repo"
	"github.com/art-vbst/art-backend/internal/auth/service"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	service *service.AuthService
}

func New(db *store.Store, env *config.Config) *AuthHandler {
	repo := repo.New(db)
	service := service.New(repo, env)
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	r.Get("/me", h.me)
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

	loginData, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.SetAccessCookie(w, loginData.AccessToken)
	utils.SetRefreshCookie(w, loginData.RefreshToken)

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GetAccessCookie(w, r)
	if err != nil {
		return
	}

	user, err := h.service.Authenticate(r.Context(), token)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	if err := h.service.Logout(r.Context(), user.ID); err != nil {
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

	user, err := h.service.Authenticate(r.Context(), token)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	default:
		log.Printf("auth error: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "An unknown error occurred")
	}
}
