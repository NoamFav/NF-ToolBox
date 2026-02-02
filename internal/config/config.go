package config

import "os"

type Config struct {
	Port        string
	Env         string
	DatabaseURL string

	JWTSecret  string
	PrivateKey string
	PublicKey  string

	StripeSecretKey    string
	StripeWebhookSecret string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Env:         getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/license_db?sslmode=disable"),

		JWTSecret:  getEnv("JWT_SECRET", ""),
		PrivateKey: getEnv("PRIVATE_KEY", ""),
		PublicKey:  getEnv("PUBLIC_KEY", ""),

		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
