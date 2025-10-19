package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ValidationError represents a user error on the order intent payload.
type ValidationError struct {
	Reason string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return e.Reason
}

// ErrOrderNotFound indicates an order could not be located.
var ErrOrderNotFound = errors.New("order not found")

// OrderRepository describes the persistence contract for orders.
type OrderRepository interface {
	Create(ctx context.Context, order Order) error
	Get(ctx context.Context, id string) (Order, error)
}

// OrderIntent represents the payload required to submit an order from a bot.
type OrderIntent struct {
	BotID    string  `json:"bot_id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
}

// Order describes the status of an order after processing.
type Order struct {
	ID        string    `json:"id"`
	BotID     string    `json:"bot_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Service exposes the executor operations.
type Service interface {
	SubmitOrder(ctx context.Context, intent OrderIntent) (Order, error)
	GetOrder(ctx context.Context, id string) (Order, error)
}

type service struct {
	repo OrderRepository
	now  func() time.Time
}

// New constructs an executor service.
func New(repo OrderRepository, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return &service{
		repo: repo,
		now:  func() time.Time { return now().UTC() },
	}
}

func (s *service) SubmitOrder(ctx context.Context, intent OrderIntent) (Order, error) {
	if err := validateIntent(intent); err != nil {
		return Order{}, err
	}

	now := s.now()
	order := Order{
		ID:        fmt.Sprintf("ord-%s", now.Format("20060102150405")),
		BotID:     strings.TrimSpace(intent.BotID),
		Symbol:    strings.TrimSpace(intent.Symbol),
		Side:      strings.ToLower(strings.TrimSpace(intent.Side)),
		Quantity:  intent.Quantity,
		Price:     intent.Price,
		Status:    "accepted",
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (s *service) GetOrder(ctx context.Context, id string) (Order, error) {
	return s.repo.Get(ctx, id)
}

func validateIntent(intent OrderIntent) error {
	if strings.TrimSpace(intent.BotID) == "" {
		return ValidationError{Reason: "bot_id is required"}
	}
	if strings.TrimSpace(intent.Symbol) == "" {
		return ValidationError{Reason: "symbol is required"}
	}
	if intent.Quantity <= 0 {
		return ValidationError{Reason: "quantity must be greater than zero"}
	}
	side := strings.ToLower(strings.TrimSpace(intent.Side))
	if side != "buy" && side != "sell" {
		return ValidationError{Reason: "side must be buy or sell"}
	}
	return nil
}
