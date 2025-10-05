package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port  string
	DbUrl string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := Config{
		Port:  os.Getenv("PORT"),
		DbUrl: os.Getenv("DB_URL"),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	return &config
}
