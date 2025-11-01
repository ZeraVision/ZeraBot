package telegram

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ZeraVision/ZeraBot/db"
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
	chatID := message.Chat.ID
	userID := message.From.ID
	command := message.Command()
	args := message.CommandArguments()

	// Check if the command requires admin privileges
	isRestrictedCommand := strings.ToLower(command) == "proposalsubscribe" ||
		strings.ToLower(command) == "proposalunsubscribe"

	if isRestrictedCommand {
		isAdmin, err := b.isGroupAdmin(chatID, userID)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			b.SendMessage(chatID, "❌ Failed to verify admin status. Please try again later.")
			return
		}
		if !isAdmin {
			b.SendMessage(chatID, "❌ This command is only available to group administrators.")
			return
		}
	}

	switch strings.ToLower(command) {
	case "start":
		b.sendHelpMessage(chatID)
	case "help":
		// Only respond to /help if it's addressed to this bot (private chat or /help@botname in groups)
		if !b.isCommandAddressedToBot(message) {
			return // Ignore the command if not addressed to this bot
		}
		b.sendHelpMessage(chatID)
	case "proposalsubscribe":
		if err := b.handleSubscribe(chatID, args); err != nil {
			log.Printf("Error handling subscribe command: %v", err)
			b.SendMessage(chatID, "❌ Failed to subscribe. Please try again later.")
		}
	case "proposalunsubscribe":
		if err := b.handleUnsubscribe(chatID, args); err != nil {
			log.Printf("Error handling unsubscribe command: %v", err)
			b.SendMessage(chatID, "❌ Failed to unsubscribe. Please try again later.")
		}
	case "mysubscriptions":
		if err := b.handleMySubscriptions(chatID); err != nil {
			log.Printf("Error handling my subscriptions command: %v", err)
			b.SendMessage(chatID, "❌ Failed to list subscriptions. Please try again later.")
		}
	default:
		b.SendMessage(chatID, "❌ Unknown command. Use /help to see available commands.")
	}
}

// isCommandAddressedToBot checks if a command was explicitly addressed to this bot
// Returns true for private chats or when command contains @botusername in groups
func (b *Bot) isCommandAddressedToBot(message *tgbotapi.Message) bool {
	chatID := message.Chat.ID
	userID := message.From.ID

	// Private chats are always addressed to the bot
	if chatID == userID {
		return true
	}

	// For groups, check if the message contains @botusername
	botInfo, err := b.API.GetMe()
	if err != nil {
		log.Printf("Error getting bot info: %v", err)
		// If we can't get bot info, allow the command (safer default)
		return true
	}

	botUsername := botInfo.UserName
	messageText := message.Text

	// Check if the message contains @botusername
	return strings.Contains(strings.ToLower(messageText), "@"+strings.ToLower(botUsername))
}

// sendHelpMessage sends the help message to the specified chat
func (b *Bot) sendHelpMessage(chatID int64) {
	helpText := `🤖 *Zera Bot Help* 🤖

*Available commands:*
/start - Start the bot
/help - Show this help message
/proposalSubscribe [symbols] - Subscribe to proposal updates
/proposalUnsubscribe [symbols] - Unsubscribe from proposal updates
/mySubscriptions - List all your current subscriptions

*Examples:*
- Subscribe to multiple tokens: /proposalSubscribe ZRA,ETH,BTC
- Unsubscribe from all: /proposalUnsubscribe all
- Unsubscribe from specific tokens: /proposalUnsubscribe ETH,BTC
- Check your subscriptions: /mySubscriptions

*Note:* Use 'all' to manage all subscriptions at once.`

	b.SendMessage(chatID, helpText)
}

// SendMessage sends a message to the specified chat with Markdown parsing
func (b *Bot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := b.API.Send(msg); err != nil {
		log.Printf("Markdown parsing failed for chat %d, retrying without parsing: %v", chatID, err)
		// If Markdown parsing fails, try without parsing
		msg.ParseMode = ""
		if _, err := b.API.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

// NotifySubscribers sends a notification to all subscribers of a specific symbol
func (b *Bot) NotifySubscribers(symbol string, message string) error {
	subscribers, err := b.subRepo.GetSubscribers(context.Background(), symbol, db.ProposalType)
	if err != nil {
		return fmt.Errorf("failed to get subscribers: %w", err)
	}

	// Only keep ZERABot test chat if in dev mode
	if os.Getenv("DEV") == "TRUE" {
		// Filter subscribers to only keep -4897181115 if it exists
		var filteredSubscribers []int64
		for _, subID := range subscribers {
			if subID == -4897181115 {
				filteredSubscribers = append(filteredSubscribers, subID)
			}
		}
		subscribers = filteredSubscribers
	}

	for _, chatID := range subscribers {

		if err := SendToChatID(chatID, message); err != nil {
			log.Printf("Failed to send notification to chat %d: %v", chatID, err)
		}
	}

	return nil
}
