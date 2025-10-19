package bots

import (
	"context"
	"errors"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("bot not found")

// Repository defines storage operations for bots.
type Repository interface {
	List(ctx context.Context) ([]Bot, error)
	Get(ctx context.Context, id string) (Bot, error)
	Save(ctx context.Context, bot Bot) (Bot, error)
}

// MemoryRepository keeps bot state in-memory.
type MemoryRepository struct {
	mu   sync.RWMutex
	bots map[string]Bot
}

// NewMemoryRepository returns an empty bot repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		bots: make(map[string]Bot),
	}
}

func (r *MemoryRepository) List(_ context.Context) ([]Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Bot, 0, len(r.bots))
	for _, bot := range r.bots {
		items = append(items, bot)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items, nil
}

func (r *MemoryRepository) Get(_ context.Context, id string) (Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if bot, ok := r.bots[id]; ok {
		return bot, nil
	}
	return Bot{}, ErrNotFound
}

func (r *MemoryRepository) Save(_ context.Context, bot Bot) (Bot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bots[bot.ID] = bot
	return bot, nil
}
