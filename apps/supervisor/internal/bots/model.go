package bots

import (
	"encoding/json"
	"time"
)

// Bot represents the desired state of a trading bot.
type Bot struct {
	ID          string          `json:"id"`
	AccountID   string          `json:"account_id"`
	Name        string          `json:"name"`
	Image       string          `json:"image"`
	Enabled     bool            `json:"enabled"`
	Config      json.RawMessage `json:"config"`
	ConfigRev   int             `json:"config_rev"`
	Description string          `json:"description"`
	Phase       string          `json:"phase"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// UpsertInput captures the payload required to create or update a bot.
type UpsertInput struct {
	ID          string
	AccountID   string
	Name        string
	Image       string
	Enabled     bool
	Config      json.RawMessage
	Description string
}
