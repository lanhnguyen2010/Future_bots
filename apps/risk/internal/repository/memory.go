package repository

import (
	"context"

	"github.com/future-bots/risk/internal/service"
)

// Memory implements service.RiskRepository returning a static exposure limit.
type Memory struct {
	maxQuantity float64
}

// NewMemory creates a repository that always returns the provided max quantity.
// When maxQuantity is not supplied or invalid, a default of 10 lots is used.
func NewMemory(maxQuantity float64) *Memory {
	if maxQuantity <= 0 {
		maxQuantity = 10
	}
	return &Memory{maxQuantity: maxQuantity}
}

// FetchLimits returns the configured exposure limits.
func (m *Memory) FetchLimits(_ context.Context, _, _, _ string) (service.RiskLimits, error) {
	return service.RiskLimits{MaxQuantity: m.maxQuantity}, nil
}
