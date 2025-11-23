package main

import (
	"context"
	"log"
	"net/http"

	"time"

	"github.com/art-vbst/art-backend/internal/platform/assets"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/pooler"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/mailer"
	"github.com/art-vbst/art-backend/internal/platform/router"
	"github.com/art-vbst/art-backend/internal/platform/storage"
	"github.com/art-vbst/art-backend/internal/platform/utils"
)

func main() {
	ctx := context.Background()
	env := config.Load()

	pool := pooler.GetDbConnectionPool(ctx, env)
	defer pool.Close()
	store := store.New(pool)

	provider := storage.NewProvider(env)
	defer provider.Close()

	mailer := mailer.New(env)

	assets := assets.Load()
	defer assets.Close()

	r := router.New(store, provider, env, mailer, assets).CreateRouter()

	if config.IsDebug() {
		log.Printf("[WARNING] debug mode enabled")
	}

	srv := &http.Server{
		Addr:              ":" + env.Port,
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 * utils.MB,
	}

	log.Printf("Server starting on :%s", env.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
