package config

import (
	"log"
	"os"
	"reflect"
	"slices"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	FrontendUrl         string
	AdminUrl            string
	CookieDomain        string
	JwtSecret           string
	DbUrl               string
	GCSBucketName       string
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
		Port:                os.Getenv("PORT"),
		FrontendUrl:         os.Getenv("FRONTEND_URL"),
		AdminUrl:            os.Getenv("ADMIN_URL"),
		CookieDomain:        os.Getenv("COOKIE_DOMAIN"),
		JwtSecret:           os.Getenv("JWT_SECRET"),
		DbUrl:               os.Getenv("DB_URL"),
		GCSBucketName:       os.Getenv("GCS_BUCKET_NAME"),
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

func IsDebug() bool {
	return os.Getenv("DEBUG") == "true"
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
