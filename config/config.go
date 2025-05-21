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
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	domain := os.Getenv("DOMAIN")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "production" // Default to production for safety
	}

	if botToken == "" || domain == "" || webhookSecret == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN, DOMAIN, and WEBHOOK_SECRET environment variables must be set")
	}

	webhookURL := fmt.Sprintf("https://%s/%s", domain, webhookSecret)

	return &Config{
		BotToken:   botToken,
		Domain:     domain,
		WebhookURL: webhookURL,
		Email:      os.Getenv("EMAIL"),
		Env:        env,
	}, nil
}
