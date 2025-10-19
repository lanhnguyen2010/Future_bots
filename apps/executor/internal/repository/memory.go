package repository

import (
	"context"
	"sync"

	"github.com/future-bots/executor/internal/service"
)

// Memory stores orders in-memory for testing and local development.
type Memory struct {
	mu     sync.RWMutex
	orders map[string]service.Order
}

// NewMemory constructs an empty memory-backed repository.
func NewMemory() *Memory {
	return &Memory{
		orders: make(map[string]service.Order),
	}
}

// Create persists an order.
func (m *Memory) Create(_ context.Context, order service.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[order.ID] = order
	return nil
}

// Get retrieves an order by id.
func (m *Memory) Get(_ context.Context, id string) (service.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	order, ok := m.orders[id]
	if !ok {
		return service.Order{}, service.ErrOrderNotFound
	}
	return order, nil
}
