package main

import (
	"context"
	"log"
	"net/http"

	"github.com/talmage89/art-backend/internal/platform/config"
	"github.com/talmage89/art-backend/internal/platform/db/pooler"
	"github.com/talmage89/art-backend/internal/platform/router"
)

func main() {
	ctx := context.Background()
	config := config.Load()

	queries, pool := pooler.GetDbConnectionPool(ctx, config)
	defer pool.Close()

	r := router.NewRouterService(config, queries).CreateRouter()

	log.Printf("Server starting on :%s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		log.Fatal(err)
	}
}
