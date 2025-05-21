package db

import (
	"context"
	"errors"
	"fmt"
)

// SubscriptionType represents the type of subscription
type SubscriptionType string

const (
	// ProposalType is a subscription for proposal updates
	ProposalType SubscriptionType = "proposal"
)

// Subscription represents a user's subscription to a specific symbol and type
type Subscription struct {
	ID        string
	ChatID    int64
	Symbol    string
	Type      SubscriptionType
	CreatedAt string
	UpdatedAt string
}

// SubscriptionRepository handles database operations for subscriptions
type SubscriptionRepository struct {
	db *Database
}

func NewSubscriptionRepository(db *Database) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Subscribe adds a new subscription or returns existing one
func (r *SubscriptionRepository) Subscribe(ctx context.Context, chatID int64, subType SubscriptionType, symbol string) (*Subscription, error) {
	const query = `
		INSERT INTO subscriptions (chat_id, symbol, type)
		VALUES ($1, $2, $3)
		ON CONFLICT (chat_id, symbol, type) DO UPDATE
		SET updated_at = NOW()
		RETURNING id, chat_id, symbol, type, created_at, updated_at
	`

	sub := &Subscription{}
	err := r.db.DB().QueryRowContext(
		ctx,
		query,
		chatID,
		symbol,
		subType,
	).Scan(
		&sub.ID,
		&sub.ChatID,
		&sub.Symbol,
		&sub.Type,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	return sub, nil
}

// Unsubscribe removes a subscription
func (r *SubscriptionRepository) Unsubscribe(ctx context.Context, chatID int64, subType SubscriptionType, symbol string) error {
	const query = `DELETE FROM subscriptions WHERE chat_id = $1 AND symbol = $2 AND type = $3`

	result, err := r.db.DB().ExecContext(ctx, query, chatID, symbol, subType)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("subscription not found")
	}

	return nil
}

// GetSubscribers returns all chat IDs subscribed to a specific symbol and type
func (r *SubscriptionRepository) GetSubscribers(ctx context.Context, symbol string, subType SubscriptionType) ([]int64, error) {
	const query = `SELECT chat_id FROM subscriptions WHERE symbol = $1 OR symbol = 'all' AND type = $2`

	rows, err := r.db.DB().QueryContext(ctx, query, symbol, subType)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers: %w", err)
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("failed to scan chat ID: %w", err)
		}
		chatIDs = append(chatIDs, chatID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return chatIDs, nil
}

// GetUserSubscriptions returns all subscriptions for a specific chat ID
func (r *SubscriptionRepository) GetUserSubscriptions(ctx context.Context, chatID int64) ([]*Subscription, error) {
	const query = `
		SELECT id, chat_id, symbol, type, created_at, updated_at
		FROM subscriptions
		WHERE chat_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.DB().QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		sub := &Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.ChatID,
			&sub.Symbol,
			&sub.Type,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	return subscriptions, nil
}

// UnsubscribeAll removes all subscriptions for a specific chat ID and subscription type
func (r *SubscriptionRepository) UnsubscribeAll(ctx context.Context, chatID int64, subType SubscriptionType) error {
	const query = `
		DELETE FROM subscriptions
		WHERE chat_id = $1 AND type = $2
	`

	_, err := r.db.DB().ExecContext(ctx, query, chatID, subType)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe all: %w", err)
	}

	return nil
}
