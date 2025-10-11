package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/talmage89/art-backend/internal/artwork/transport"
	"github.com/talmage89/art-backend/internal/platform/config"
	"github.com/talmage89/art-backend/internal/platform/db/store"
)

type RouterService struct {
	db     *store.Store
	config *config.Config
}

func New(db *store.Store, config *config.Config) *RouterService {
	return &RouterService{
		db:     db,
		config: config,
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

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.config.FrontendUrl, "https://checkout.stripe.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *RouterService) registerRoutes(r *chi.Mux) {
	artworkHandler := transport.NewArtworkHandler(s.db)
	r.Mount("/artwork", artworkHandler.Routes())

	checkoutHandler := transport.NewCheckoutHandler(s.db, s.config)
	r.Mount("/checkout", checkoutHandler.Routes())
}
