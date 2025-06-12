package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jfederk/ZeraBot/telegram"
)

type Server struct {
	httpServer *http.Server
	bot        *telegram.Bot
	webhookURL string
}

// New creates a new server instance
func New(bot *telegram.Bot, domain, webhookPath string, isProduction bool) (*Server, error) {
	s := &Server{
		bot: bot,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(webhookPath, s.webhookHandler)

	server := &http.Server{
		Handler: mux,
	}

	server.Addr = ":8080"
	if isProduction {
		s.webhookURL = fmt.Sprintf("https://%s%s", domain, webhookPath)
		log.Printf("Running in production mode on %s", server.Addr)
	} else {
		if ngrokURL := os.Getenv("NGROK_URL"); ngrokURL != "" {
			s.webhookURL = fmt.Sprintf("%s%s", ngrokURL, webhookPath)
		}
		log.Printf("Running in development mode on http://localhost%s", server.Addr)
	}

	s.httpServer = server
	return s, nil
}

// Start starts the server
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)

	if s.httpServer.TLSConfig != nil {
		return s.httpServer.ListenAndServeTLS("", "")
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error during server shutdown: %w", err)
	}
	log.Println("Server stopped")
	return nil
}

// WebhookURL returns the webhook URL that was set
func (s *Server) WebhookURL() string {
	return s.webhookURL
}

func (s *Server) webhookHandler(w http.ResponseWriter, r *http.Request) {
	s.bot.WebhookHandler(w, r)
}
