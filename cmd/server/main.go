package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/talmage89/art-backend/internal/api"
	"github.com/talmage89/art-backend/internal/config"
	"github.com/talmage89/art-backend/internal/db"
)

func getDbConnection(ctx context.Context, env *config.Config) *db.Queries {
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

	return db.New(pool)
}

func main() {
	env := config.Load()
	ctx := context.Background()

	queries := getDbConnection(ctx, env)
	artworkHandler := api.NewArtworkHandler(queries)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Throttle(100))

	r.Mount("/artwork", artworkHandler.Routes())

	log.Printf("Server starting on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, r); err != nil {
		log.Fatal(err)
	}
}
