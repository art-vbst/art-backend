package router

import (
	"fmt"
	"net/http"
	"time"

	artwork "github.com/art-vbst/art-backend/internal/artwork/transport"
	auth "github.com/art-vbst/art-backend/internal/auth/transport"
	payments "github.com/art-vbst/art-backend/internal/payments/transport"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/mailer"
	"github.com/art-vbst/art-backend/internal/platform/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouterService struct {
	db       *store.Store
	provider storage.Provider
	config   *config.Config
	mailer   mailer.Mailer
}

func New(db *store.Store, provider storage.Provider, config *config.Config, mailer mailer.Mailer) *RouterService {
	return &RouterService{
		db:       db,
		provider: provider,
		config:   config,
		mailer:   mailer,
	}
}

func (s *RouterService) CreateRouter() *chi.Mux {
	r := chi.NewRouter()
	s.registerMiddleware(r)
	s.registerRoutes(r)
	return r
}

func (s *RouterService) registerMiddleware(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Throttle(100))

	allowedOrigins := []string{s.config.FrontendUrl, s.config.AdminUrl, "https://checkout.stripe.com"}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(securityHeaders)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func (s *RouterService) registerRoutes(r *chi.Mux) {
	authHandler := auth.New(s.db, s.config)
	r.Mount("/auth", authHandler.Routes())

	artworkHandler := artwork.NewArtworkHandler(s.db, s.provider, s.config)
	r.Mount("/artworks", artworkHandler.Routes())

	imageHandler := artwork.NewImageHandler(s.db, s.provider, s.config)
	imagesRoute := fmt.Sprintf("/artworks/{%s}/images", artwork.ArtworkIDParam)
	r.Mount(imagesRoute, imageHandler.Routes())

	ordersHandler := payments.NewOrdersHandler(s.db, s.config, s.mailer)
	r.Mount("/orders", ordersHandler.Routes())

	checkoutHandler := payments.NewCheckoutHandler(s.db, s.config)
	r.Mount("/checkout", checkoutHandler.Routes())

	webhookHandler := payments.NewWebhookHandler(s.db, s.config, s.mailer)
	r.Mount("/stripe/webhook", webhookHandler.Routes())

	if config.IsDebug() {
		if localStorage, ok := s.provider.(*storage.LocalStorage); ok {
			r.Mount("/uploads", serveLocalStorage(localStorage))
		}
	}
}
