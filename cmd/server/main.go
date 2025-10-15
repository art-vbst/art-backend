package main

import (
	"context"
	"log"
	"net/http"

	"github.com/talmage89/art-backend/internal/platform/config"
	"github.com/talmage89/art-backend/internal/platform/db/pooler"
	"github.com/talmage89/art-backend/internal/platform/db/store"
	"github.com/talmage89/art-backend/internal/platform/router"
)

func main() {
	ctx := context.Background()
	config := config.Load()

	pool := pooler.GetDbConnectionPool(ctx, config)
	defer pool.Close()

	store := store.New(pool)

	r := router.New(store, config).CreateRouter()

	log.Printf("Server starting on :%s", config.Port)
	log.Printf("included for ci retrigger...")
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		log.Fatal(err)
	}
}
