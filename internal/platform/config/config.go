package config

import (
	"log"
	"os"
	"reflect"
	"slices"

	"github.com/joho/godotenv"
)

type Config struct {
	Debug               string
	Port                string
	DbUrl               string
	FrontendUrl         string
	StripeSecret        string
	StripeWebhookSecret string
	MailgunDomain       string
	MailgunApiKey       string
	TestEmail           string
	EmailFromName       string
	EmailSignature      string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := Config{
		Debug:               os.Getenv("DEBUG"),
		Port:                os.Getenv("PORT"),
		DbUrl:               os.Getenv("DB_URL"),
		FrontendUrl:         os.Getenv("FRONTEND_URL"),
		StripeSecret:        os.Getenv("STRIPE_SECRET"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		MailgunDomain:       os.Getenv("MAILGUN_DOMAIN"),
		MailgunApiKey:       os.Getenv("MAILGUN_API_KEY"),
		TestEmail:           os.Getenv("TEST_EMAIL"),
		EmailFromName:       os.Getenv("EMAIL_FROM_NAME"),
		EmailSignature:      os.Getenv("EMAIL_SIGNATURE"),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	ensureRequiredVars(&config)

	return &config
}

func ensureRequiredVars(config *Config) {
	optionalVars := []string{"Debug", "TestEmail"}

	typ := reflect.TypeOf(*config)
	val := reflect.ValueOf(*config)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		if slices.Contains(optionalVars, field.Name) {
			continue
		}
		if value.String() == "" {
			log.Fatalf("Missing required environment variable: %s", field.Name)
		}
	}
}
