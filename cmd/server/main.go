package main

import (
	"context"
	"log"
	"net/http"

	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/pooler"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/mailer"
	"github.com/art-vbst/art-backend/internal/platform/router"
)

func main() {
	ctx := context.Background()
	env := config.Load()

	pool := pooler.GetDbConnectionPool(ctx, env)
	defer pool.Close()
	store := store.New(pool)

	mailer := mailer.New(env)

	r := router.New(store, env, mailer).CreateRouter()

	if config.IsDebug() {
		log.Printf("[WARNING] debug mode enabled")
	}

	log.Printf("Server starting on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, r); err != nil {
		log.Fatal(err)
	}
}
