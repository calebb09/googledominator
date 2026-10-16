package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	Env                 string
	DatabaseURL         string
	FrontendURL         string
	BaseURL             string
	PublicAPIURL        string
	StripeSecretKey     string
	StripeWebhookSecret string
	JWTSecret           string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("[INFO] No .env file found or error loading, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@localhost:5432/googledominator?schema=public"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://googledominator.co"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = frontendURL
	}

	publicAPIURL := os.Getenv("PUBLIC_API_URL")

	stripeSecret := os.Getenv("STRIPE_SECRET_KEY")
	if stripeSecret == "" {
		stripeSecret = "sk_test_51MockStripeSecretKeyGoogleDominatorKey"
	}

	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if stripeWebhookSecret == "" {
		stripeWebhookSecret = "whsec_mock_stripe_webhook_secret_key"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "googledominator-super-secret-jwt-key-2026"
	}

	return &Config{
		Port:                port,
		Env:                 env,
		DatabaseURL:         dbURL,
		FrontendURL:         frontendURL,
		BaseURL:             baseURL,
		PublicAPIURL:        publicAPIURL,
		StripeSecretKey:     stripeSecret,
		StripeWebhookSecret: stripeWebhookSecret,
		JWTSecret:           jwtSecret,
	}
}
