package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/talmage89/art-backend/internal/api"
	"github.com/talmage89/art-backend/internal/api/payments"
	"github.com/talmage89/art-backend/internal/config"
	"github.com/talmage89/art-backend/internal/db"
)

func getDbConnectionPool(ctx context.Context, env *config.Config) (*db.Queries, *pgxpool.Pool) {
	poolConfig, err := pgxpool.ParseConfig(env.DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Minute * 30
	poolConfig.MaxConnIdleTime = time.Minute * 5
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	queries := db.New(pool)

	return queries, pool
}

func getRouter(env *config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Throttle(100))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.FrontendUrl, "https://checkout.stripe.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return r
}

func main() {
	env := config.Load()
	ctx := context.Background()

	queries, pool := getDbConnectionPool(ctx, env)
	defer pool.Close()

	r := getRouter(env)

	artworkHandler := api.NewArtworkHandler(queries)
	r.Mount("/artwork", artworkHandler.Routes())

	paymentsHandler := payments.NewPaymentsHandler(env, queries)
	r.Mount("/stripe", paymentsHandler.Routes())

	log.Printf("Server starting on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, r); err != nil {
		log.Fatal(err)
	}
}
