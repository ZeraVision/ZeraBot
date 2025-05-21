package telegram

import (
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WebhookHandler handles incoming webhook requests
func (b *Bot) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	update, err := b.API.HandleUpdate(r)
	if err != nil {
		log.Printf("Error handling update: %v", err)
		return
	}

	if update.Message != nil {
		chatID := update.Message.Chat.ID
		username := update.Message.From.UserName
		log.Printf("Received message from chat ID: %d, username: @%s", chatID, username)
		
		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		}
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	switch message.Command() {
	case "start":
		msg.Text = "👋 Welcome to ZeraBot! I'm here to help. Use /help to see available commands."
	case "help":
		msg.Text = `🤖 *ZeraBot Help*

Available commands:
/start - Start the bot
/help - Show this help message
/hello - Say hello to the bot`
		msg.ParseMode = "Markdown"
	case "hello":
		msg.Text = "👋 Hello, " + message.From.FirstName + "!"
	default:
		msg.Text = "I don't know that command. Try /help to see available commands."
	}

	if _, err := b.API.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
