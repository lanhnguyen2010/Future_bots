package service

import (
	"context"
	"time"
)

// PnLRepository represents the storage dependency for PnL generation.
type PnLRepository interface {
	FetchPnL(ctx context.Context, accountID, botID, window string) (PnLData, error)
}

// PnLData carries the raw values aggregated by the repository.
type PnLData struct {
	Realized   float64
	Unrealized float64
}

// PnLQuery represents the filters used for generating a PnL report.
type PnLQuery struct {
	AccountID string
	BotID     string
	Window    string
}

// PnLReport describes the response payload for the PnL endpoint.
type PnLReport struct {
	AccountID  string    `json:"account_id"`
	BotID      string    `json:"bot_id"`
	Realized   float64   `json:"realized"`
	Unrealized float64   `json:"unrealized"`
	Window     string    `json:"window"`
	Generated  time.Time `json:"generated_at"`
}

// Service defines the reports contract.
type Service interface {
	GeneratePnLReport(ctx context.Context, query PnLQuery) (PnLReport, error)
}

type service struct {
	repo             PnLRepository
	defaultAccountID string
	defaultWindow    string
	now              func() time.Time
}

// New builds a reports service with sane defaults.
func New(repo PnLRepository, defaultAccountID string, now func() time.Time) Service {
	if defaultAccountID == "" {
		defaultAccountID = "acct-123"
	}
	if now == nil {
		now = time.Now
	}
	return &service{
		repo:             repo,
		defaultAccountID: defaultAccountID,
		defaultWindow:    "1d",
		now:              func() time.Time { return now().UTC() },
	}
}

func (s *service) GeneratePnLReport(ctx context.Context, query PnLQuery) (PnLReport, error) {
	accountID := query.AccountID
	if accountID == "" {
		accountID = s.defaultAccountID
	}

	window := query.Window
	if window == "" {
		window = s.defaultWindow
	}

	data, err := s.repo.FetchPnL(ctx, accountID, query.BotID, window)
	if err != nil {
		return PnLReport{}, err
	}

	return PnLReport{
		AccountID:  accountID,
		BotID:      query.BotID,
		Realized:   data.Realized,
		Unrealized: data.Unrealized,
		Window:     window,
		Generated:  s.now(),
	}, nil
}
