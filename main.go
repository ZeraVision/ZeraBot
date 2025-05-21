package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/acme/autocert"
)

type Config struct {
	BotToken   string
	Domain     string
	WebhookURL string
	Email      string
	Env        string // development or production
}

var (
	bot *tgbotapi.BotAPI
	cfg Config
)

func init() {
	godotenv.Load(".env")
}

func main() {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize bot
	if err := initBot(); err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	// Set up webhook
	if err := setupWebhook(); err != nil {
		log.Fatalf("Failed to set up webhook: %v", err)
	}

	// Set up and start server
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	server := setupServer()
	startServers(server)

	SendToChatID(-4897181115, "SUP")

	// Wait for interrupt signal
	<-stopChan
	shutdownServer(server)
}

func loadConfig() error {
	// Get configuration from environment
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	domain := os.Getenv("DOMAIN")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "production" // Default to production for safety
	}

	if botToken == "" || domain == "" || webhookSecret == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN, DOMAIN, and WEBHOOK_SECRET environment variables must be set")
	}

	webhookURL := fmt.Sprintf("https://%s/%s", domain, webhookSecret)
	if env == "development" {
		// In development, we'll use ngrok's URL directly
		ngrokURL := os.Getenv("NGROK_URL")
		if ngrokURL != "" {
			webhookURL = fmt.Sprintf("%s/%s", ngrokURL, webhookSecret)
		}
	} else {
		webhookURL = fmt.Sprintf("https://%s/%s", domain, webhookSecret)
	}

	cfg = Config{
		BotToken:   botToken,
		Domain:     domain,
		WebhookURL: webhookURL,
		Email:      os.Getenv("LETSENCRYPT_EMAIL"),
		Env:        env,
	}

	return nil
}

func initBot() error {
	var err error
	bot, err = tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return fmt.Errorf("creating bot: %w", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return nil
}

func setupWebhook() error {
	// Remove any existing webhook
	if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		return fmt.Errorf("removing existing webhook: %w", err)
	}

	// Set up new webhook
	wh, err := tgbotapi.NewWebhook(cfg.WebhookURL)
	if err != nil {
		return fmt.Errorf("creating webhook: %w", err)
	}

	if _, err = bot.Request(wh); err != nil {
		return fmt.Errorf("setting webhook: %w", err)
	}

	// Verify webhook
	info, err := bot.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("getting webhook info: %w", err)
	}

	log.Printf("Webhook set to: %s", info.URL)
	return nil
}

func setupServer() *http.Server {
	mux := http.NewServeMux()
	webhookPath := "/" + os.Getenv("WEBHOOK_SECRET")
	mux.HandleFunc(webhookPath, handleWebhook)

	// For development, use HTTP on port 8080
	if cfg.Env == "development" {
		log.Printf("Running in development mode on http://localhost:8080%s", webhookPath)
		return &http.Server{
			Addr:    ":8080",
			Handler: mux,
		}
	}

	// Production setup with Let's Encrypt
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domain),
		Cache:      autocert.DirCache("certs"),
		Email:      cfg.Email,
	}

	return &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}
}

func startServers(server *http.Server) {
	// Check if we're in development mode
	isDev := cfg.Env == "development"

	if !isDev {
		// In production, start HTTP server for Let's Encrypt HTTP-01 challenge
		go func() {
			http.HandleFunc("/.well-known/acme-challenge/", func(w http.ResponseWriter, r *http.Request) {
				http.FileServer(http.Dir(".well-known/acme-challenge")).ServeHTTP(w, r)
			})
			log.Fatal(http.ListenAndServe(":80", nil))
		}()

		// Start HTTPS server in production
		log.Println("Starting HTTPS server on :443")
		go func() {
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		}()
	} else {
		// In development, just start the HTTP server
		log.Printf("Starting development server on %s", server.Addr)
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}()
	}
}

func shutdownServer(server *http.Server) {
	log.Println("Shutting down...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
	log.Println("Server stopped")
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	update, err := bot.HandleUpdate(r)
	if err != nil {
		log.Printf("Error handling update: %v", err)
		return
	}

	if update.Message != nil && update.Message.IsCommand() {
		handleCommand(update.Message)
	}
}

func handleCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	fmt.Printf("Message from: %d\n", message.Chat.ID)

	switch message.Command() {
	case "start":
		msg.Text = "👋 Welcome to ZeraBot! I'm here to help. Use /help to see available commands."
	case "help":
		msg.Text = `🤖 *ZeraBot Help*

Available commands:
/start - Start the bot
/help - Show this help message
/hello - Say hello to the bot
`
		msg.ParseMode = "Markdown"
	case "hello":
		msg.Text = "👋 Hello, " + message.From.FirstName + "!"
	default:
		msg.Text = "I don't know that command. Try /help to see available commands."
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// SendToChatID sends a message to a specific chat ID
// chatID: The Telegram chat ID to send the message to
// message: The message text to send (supports Markdown formatting)
// Returns error if the message couldn't be sent
func SendToChatID(chatID int64, message string) error {
	if bot == nil {
		return fmt.Errorf("bot is not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}
	return nil
}
