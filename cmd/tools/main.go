package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/db/pooler"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/tools"
)

func main() {
	ctx := context.Background()
	config := config.Load()

	pool := pooler.GetDbConnectionPool(ctx, config)
	defer pool.Close()
	store := store.New(pool)

	if len(os.Args[1:]) == 0 {
		fmt.Println("A command must be specified")
		fmt.Println("Available commands: createuser")
		return
	}

	switch os.Args[1] {
	case "createuser":
		if err := tools.CreateUser(ctx, store); err != nil {
			log.Fatal(err)
		}
	}
}
