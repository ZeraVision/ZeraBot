package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jfederk/ZeraBot/config"
	"github.com/jfederk/ZeraBot/server"
	"github.com/jfederk/ZeraBot/telegram"
	"github.com/joho/godotenv"
)

var (
	bot *telegram.Bot
	cfg *config.Config
	srv *server.Server
)

func init() {
	godotenv.Load(".env")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize bot
	bot, err = telegram.NewBot(cfg.BotToken, cfg.Env == "development")
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	// Set up server
	webhookPath := "/" + os.Getenv("WEBHOOK_SECRET")
	srv, err = server.New(bot.API, cfg.Domain, webhookPath, cfg.Env == "production")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Set up webhook
	if err := bot.SetupWebhook(srv.WebhookURL()); err != nil {
		log.Fatalf("Failed to set up webhook: %v", err)
	}

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Send a test message
	if err := bot.SendToChatID(-4897181115, "🤖 *Bot started successfully!*"); err != nil {
		log.Printf("Failed to send startup message: %v", err)
	}

	// Wait for interrupt signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
}
