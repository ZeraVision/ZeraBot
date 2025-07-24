package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/ZeraVision/ZeraBot/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *Bot

func GetBot() *Bot {
	return bot
}

func SetBot(b *Bot) {
	bot = b
}

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
	API      *tgbotapi.BotAPI
	database *db.Database
	subRepo  *db.SubscriptionRepository
}

// NewBot creates a new Telegram bot instance with the provided token and database
func NewBot(token string, debug bool, database *db.Database) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = debug

	bot := &Bot{
		API:      api,
		database: database,
		subRepo:  db.NewSubscriptionRepository(database),
	}

	SetBot(bot)
	return bot, nil
}

// SendToChatID sends a message to a specific chat ID
func SendToChatID(chatID int64, message string) error {
	if bot == nil {
		return fmt.Errorf("bot not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	_, err := bot.API.Send(msg)
	return err
}

// isGroupAdmin checks if a user is an admin in a group
func (b *Bot) isGroupAdmin(chatID int64, userID int64) (bool, error) {
	// If it's a private chat, consider the user as admin
	if chatID == userID {
		return true, nil
	}

	// Get chat member info
	member, err := b.API.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to get chat member: %w", err)
	}

	// Check if the member is an admin or creator
	return member.IsAdministrator() || member.IsCreator(), nil
}

// isValidSymbolFormat checks if a symbol follows the format $LETTERS+4DIGITS
func isValidSymbolFormat(symbol string) bool {
	if len(symbol) < 6 { // Minimum length: $A+0000 (6 chars)
		return false
	}

	// Check for leading $
	if symbol[0] != '$' {
		return false
	}

	// Find the + separator
	plusIndex := strings.Index(symbol, "+")
	if plusIndex == -1 || plusIndex == 1 { // Must have letters before +
		return false
	}

	// Check letters part (between $ and +)
	letters := symbol[1:plusIndex]
	if len(letters) == 0 {
		return false
	}
	for _, c := range letters {
		if !unicode.IsLetter(c) {
			return false
		}
	}

	// Check numbers part (after +)
	numbers := symbol[plusIndex+1:]
	if len(numbers) != 4 {
		return false
	}
	for _, c := range numbers {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	return true
}

// processSymbols processes a comma-separated string of symbols, removes any whitespace,
// and ensures the symbol part is uppercase (e.g., $zra+0000 becomes $ZRA+0000)
func processSymbols(input string) []string {
	symbols := strings.Split(input, ",")
	var result []string
	for _, s := range symbols {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		// Convert to uppercase if it's not the "all" keyword
		if strings.ToLower(s) != "all" {
			// Split into parts before and after +
			parts := strings.SplitN(s, "+", 2)
			if len(parts) == 2 {
				// Convert the symbol part (after $) to uppercase, keep the numbers as is
				symbolPart := strings.TrimPrefix(parts[0], "$")
				s = "$" + strings.ToUpper(symbolPart) + "+" + parts[1]
			}
		}
		result = append(result, s)
	}
	return result
}

// handleSubscribe handles the /subscribe command
func (b *Bot) handleSubscribe(chatID int64, args string) error {
	symbolsInput := strings.TrimSpace(args)
	if symbolsInput == "" {
		return SendToChatID(chatID, "Please provide a symbol to subscribe to (e.g., /proposalSubscribe $ZRA+0000 or /proposalSubscribe $ZRA+0000,$ZIP+0000)")
	}

	symbols := processSymbols(symbolsInput)

	// Validate symbol format
	for _, s := range symbols {
		if strings.ToLower(s) != "all" && !isValidSymbolFormat(s) {
			return SendToChatID(chatID, "❌ Invalid symbol format. Please use format $SYMBOL+NNNN (e.g., $ZRA+0000) or 'all'")
		}
	}

	var resultMsgs []string

	// Check for "all" keyword
	hasAll := false
	for _, s := range symbols {
		if strings.ToLower(s) == "all" {
			hasAll = true
			break
		}
	}

	if hasAll {
		// Unsubscribe from all existing subscriptions first
		err := b.subRepo.UnsubscribeAll(context.Background(), chatID, db.ProposalType)
		if err != nil {
			return fmt.Errorf("failed to clear existing subscriptions: %w", err)
		}

		// Subscribe to "all" (special handling might be needed here)
		_, err = b.subRepo.Subscribe(context.Background(), chatID, db.ProposalType, "all")
		if err != nil {
			return fmt.Errorf("failed to subscribe to all: %w", err)
		}

		return SendToChatID(chatID, "✅ Subscribed to all proposals.")
	}

	// Process individual symbols
	var successCount int
	for _, symbol := range symbols {
		_, err := b.subRepo.Subscribe(context.Background(), chatID, db.ProposalType, symbol)
		if err != nil {
			resultMsgs = append(resultMsgs, fmt.Sprintf("❌ Failed to subscribe to %s: %v", symbol, err))
		} else {
			successCount++
		}
	}

	if successCount > 0 {
		msg := fmt.Sprintf("✅ Successfully subscribed to %s", symbols[0])
		if successCount > 1 {
			msg = fmt.Sprintf("✅ Successfully subscribed to %d symbols", successCount)
		}
		resultMsgs = append([]string{msg}, resultMsgs...)
	}

	return SendToChatID(chatID, strings.Join(resultMsgs, "\n"))
}

// handleUnsubscribe handles the /unsubscribe command
func (b *Bot) handleUnsubscribe(chatID int64, args string) error {
	symbolsInput := strings.TrimSpace(args)
	if symbolsInput == "" {
		return SendToChatID(chatID, "Please provide a symbol to unsubscribe from (e.g., /proposalUnsubscribe $ZRA+0000 or /proposalUnsubscribe $ZRA+0000,$ZIP+0000)")
	}

	symbols := processSymbols(symbolsInput)

	// Validate symbol format
	for _, s := range symbols {
		if strings.ToLower(s) != "all" && !isValidSymbolFormat(s) {
			return SendToChatID(chatID, "❌ Invalid symbol format. Please use format $SYMBOL+NNNN (e.g., $ZRA+0000) or 'all'")
		}
	}

	var resultMsgs []string

	// Check for "all" keyword
	hasAll := false
	for _, s := range symbols {
		if strings.ToLower(s) == "all" {
			hasAll = true
			break
		}
	}

	if hasAll {
		err := b.subRepo.UnsubscribeAll(context.Background(), chatID, db.ProposalType)
		if err != nil {
			return fmt.Errorf("failed to unsubscribe from all: %w", err)
		}
		return SendToChatID(chatID, "✅ Unsubscribed from all proposals")
	}

	// Process individual symbols
	var successCount int
	for _, symbol := range symbols {
		err := b.subRepo.Unsubscribe(context.Background(), chatID, db.ProposalType, symbol)
		if err != nil {
			resultMsgs = append(resultMsgs, fmt.Sprintf("❌ Failed to unsubscribe from %s: %v", symbol, err))
		} else {
			successCount++
		}
	}

	if successCount > 0 {
		msg := fmt.Sprintf("✅ Successfully unsubscribed from %s", symbols[0])
		if successCount > 1 {
			msg = fmt.Sprintf("✅ Successfully unsubscribed from %d symbols", successCount)
		}
		resultMsgs = append([]string{msg}, resultMsgs...)
	}

	if len(resultMsgs) == 0 {
		return SendToChatID(chatID, "No valid symbols provided to unsubscribe from.")
	}

	return SendToChatID(chatID, strings.Join(resultMsgs, "\n"))
}

// handleMySubscriptions handles the /mysubscriptions command
func (b *Bot) handleMySubscriptions(chatID int64) error {
	subs, err := b.subRepo.GetUserSubscriptions(context.Background(), chatID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(subs) == 0 {
		return SendToChatID(chatID, "You are not subscribed to any proposals yet.\nUse /subscribe [symbol] to subscribe.")
	}

	var subList []string
	for _, sub := range subs {
		subList = append(subList, fmt.Sprintf("• %s (%s)", sub.Symbol, sub.Type))
	}

	message := fmt.Sprintf("📋 Your subscriptions (%d):\n%s",
		len(subs),
		strings.Join(subList, "\n"))

	return SendToChatID(chatID, message)
}
