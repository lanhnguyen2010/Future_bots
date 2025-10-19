package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrInvalidQuantity is returned when the proposed quantity is non-positive.
var ErrInvalidQuantity = errors.New("proposed quantity must be greater than zero")

// RiskRepository represents the storage layer required by the risk service.
type RiskRepository interface {
	FetchLimits(ctx context.Context, botID, accountID, symbol string) (RiskLimits, error)
}

// RiskLimits encapsulates the exposure limits associated with a bot/account/symbol tuple.
type RiskLimits struct {
	MaxQuantity float64
}

// RiskCheckRequest represents a request to evaluate a risk exposure.
type RiskCheckRequest struct {
	BotID        string  `json:"bot_id"`
	AccountID    string  `json:"account_id"`
	Symbol       string  `json:"symbol"`
	ProposedSide string  `json:"proposed_side"`
	ProposedQty  float64 `json:"proposed_qty"`
}

// RiskCheckDecision holds the risk decision for a given request.
type RiskCheckDecision struct {
	Allowed   bool      `json:"allowed"`
	Reason    string    `json:"reason"`
	CheckedAt time.Time `json:"checked_at"`
}

// Service defines the risk evaluation contract.
type Service interface {
	Evaluate(ctx context.Context, req RiskCheckRequest) (RiskCheckDecision, error)
}

type service struct {
	repo RiskRepository
	now  func() time.Time
}

// New returns a risk service backed by the provided repository.
func New(repo RiskRepository, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return &service{
		repo: repo,
		now:  func() time.Time { return now().UTC() },
	}
}

func (s *service) Evaluate(ctx context.Context, req RiskCheckRequest) (RiskCheckDecision, error) {
	if req.ProposedQty <= 0 {
		return RiskCheckDecision{}, ErrInvalidQuantity
	}

	limits, err := s.repo.FetchLimits(ctx, req.BotID, req.AccountID, req.Symbol)
	if err != nil {
		return RiskCheckDecision{}, err
	}

	decision := RiskCheckDecision{
		Allowed:   req.ProposedQty <= limits.MaxQuantity,
		CheckedAt: s.now(),
	}

	if !decision.Allowed {
		decision.Reason = fmt.Sprintf("quantity %.2f exceeds maximum lot size %.2f", req.ProposedQty, limits.MaxQuantity)
	}

	return decision, nil
}
