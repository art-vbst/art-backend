package config

import (
	"log"
	"os"
	"reflect"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DbUrl               string
	FrontendUrl         string
	StripeSecret        string
	StripeWebhookSecret string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := Config{
		Port:                os.Getenv("PORT"),
		DbUrl:               os.Getenv("DB_URL"),
		FrontendUrl:         os.Getenv("FRONTEND_URL"),
		StripeSecret:        os.Getenv("STRIPE_SECRET"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}

	typ := reflect.TypeOf(config)
	val := reflect.ValueOf(config)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		if field.Name == "Port" {
			continue
		}
		if value.String() == "" {
			log.Fatalf("Missing required environment variable: %s", field.Name)
		}
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	return &config
}
