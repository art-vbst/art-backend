package main

import (
	"log"

	"github.com/talmage89/art-backend/internal/config"
)

func main() {
	env := config.Load()

	log.Println((env.Port))
	log.Println(env.DbUrl)
}
