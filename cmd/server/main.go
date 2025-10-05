package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	r.Get("/artworks/create", func(w http.ResponseWriter, r *http.Request) {
		artworkToCreate := db.CreateArtworkParams{
			Title:          "My Artwork",
			PaintingNumber: nil,
			PaintingYear:   nil,
			WidthInches:    pgtype.Numeric{},
			HeightInches:   pgtype.Numeric{},
			PriceCents:     nil,
			Paper:          nil,
			SortOrder:      nil,
			SoldAt:         pgtype.Timestamp{},
			Status:         db.ArtworkStatusAvailable,
			Medium:         db.ArtworkMediumOilPanel,
			Category:       db.ArtworkCategoryLandscape,
		}

		artwork, err := queries.CreateArtwork(r.Context(), artworkToCreate)
		if err != nil {
			http.Error(w, "Failed to create artwork", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(artwork); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	r.Get("/artworks", func(w http.ResponseWriter, r *http.Request) {
		artworks, err := queries.GetRecentArtworks(r.Context(), 10)
		if err != nil {
			http.Error(w, "Failed to fetch artworks", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(artworks); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	log.Printf("Server starting on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, r); err != nil {
		log.Fatal(err)
	}
}
