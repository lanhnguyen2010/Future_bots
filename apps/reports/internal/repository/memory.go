package repository

import (
	"context"

	"github.com/future-bots/reports/internal/service"
)

// Memory implements service.PnLRepository returning static data for now.
type Memory struct {
	realized   float64
	unrealized float64
}

// NewMemory constructs an in-memory repository with optional preset values.
func NewMemory(realized, unrealized float64) *Memory {
	if realized == 0 && unrealized == 0 {
		realized = 1525.23
		unrealized = 210.42
	}
	return &Memory{realized: realized, unrealized: unrealized}
}

// FetchPnL returns the configured PnL snapshot.
func (m *Memory) FetchPnL(_ context.Context, _, _, _ string) (service.PnLData, error) {
	return service.PnLData{Realized: m.realized, Unrealized: m.unrealized}, nil
}
