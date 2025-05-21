package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// NewSetWebhook creates a new webhook configuration
func NewSetWebhook(url string) tgbotapi.WebhookConfig {
	wh, _ := tgbotapi.NewWebhook(url)
	return wh
}

// SetupWebhook configures the webhook for the bot
func (b *Bot) SetupWebhook(webhookURL string) error {
	// Set webhook using the provided URL
	_, err := b.API.Request(NewSetWebhook(webhookURL))
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	// Verify webhook was set correctly
	info, err := b.API.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("failed to get webhook info: %w", err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	log.Printf("Webhook set to: %s", info.URL)
	return nil
}

type Bot struct {
	API *tgbotapi.BotAPI
}

// NewBot creates a new Telegram bot instance
func NewBot(token string, debug bool) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = debug
	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{API: api}, nil
}

// SendToChatID sends a message to a specific chat ID
func (b *Bot) SendToChatID(chatID int64, message string) error {
	if b.API == nil {
		return fmt.Errorf("bot is not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	_, err := b.API.Send(msg)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}
	return nil
}
