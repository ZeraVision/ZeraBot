package config

import (
	"fmt"
	"os"
)

type Config struct {
	BotToken   string
	Domain     string
	WebhookURL string
	Email      string
	Env        string
	DatabaseURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	domain := os.Getenv("DOMAIN")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	databaseURL := os.Getenv("DATABASE_URL")
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "production" // Default to production for safety
	}

	if botToken == "" || domain == "" || webhookSecret == "" || databaseURL == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN, DOMAIN, WEBHOOK_SECRET, and DATABASE_URL environment variables must be set")
	}

	webhookURL := fmt.Sprintf("https://%s/%s", domain, webhookSecret)

	return &Config{
		BotToken:   botToken,
		Domain:     domain,
		WebhookURL: webhookURL,
		Email:      os.Getenv("EMAIL"),
		Env:        env,
		DatabaseURL: databaseURL,
	}, nil
}
