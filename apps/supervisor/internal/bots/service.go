package bots

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Service orchestrates bot persistence alongside manifest generation.
type Service struct {
	repo     Repository
	writer   ManifestWriter
	logger   *slog.Logger
	timeFunc func() time.Time
}

// NewService constructs a bot service using the provided repository and manifest writer.
func NewService(repo Repository, writer ManifestWriter, logger *slog.Logger) *Service {
	return &Service{
		repo:     repo,
		writer:   writer,
		logger:   logger,
		timeFunc: func() time.Time { return time.Now().UTC() },
	}
}

// WithNow allows injecting deterministic time for tests.
func (s *Service) WithNow(now func() time.Time) *Service {
	s.timeFunc = now
	return s
}

// ListBots returns all bots in reverse chronological order.
func (s *Service) ListBots(ctx context.Context) ([]Bot, error) {
	return s.repo.List(ctx)
}

// UpsertBot creates or updates a bot entry and writes its manifest.
func (s *Service) UpsertBot(ctx context.Context, input UpsertInput) (Bot, error) {
	if err := validateInput(input); err != nil {
		return Bot{}, err
	}

	now := s.timeFunc()
	var existing Bot
	var err error
	configRev := 1
	if existing, err = s.repo.Get(ctx, input.ID); err == nil {
		configRev = existing.ConfigRev + 1
	} else if !errors.Is(err, ErrNotFound) {
		return Bot{}, err
	}

	bot := Bot{
		ID:          input.ID,
		AccountID:   input.AccountID,
		Name:        input.Name,
		Image:       input.Image,
		Enabled:     input.Enabled,
		Config:      cloneConfig(input.Config),
		ConfigRev:   configRev,
		Description: input.Description,
		Phase:       "stopped",
		UpdatedAt:   now,
	}

	if err == nil {
		bot.CreatedAt = existing.CreatedAt
		if existing.Description != "" && input.Description == "" {
			bot.Description = existing.Description
		}
	} else {
		bot.CreatedAt = now
	}

	if bot.Enabled {
		bot.Phase = "desired"
	}

	stored, err := s.repo.Save(ctx, bot)
	if err != nil {
		return Bot{}, err
	}

	if s.writer != nil {
		if _, err := s.writer.Write(ctx, stored); err != nil {
			s.logger.Error("failed to write bot manifest", "bot_id", stored.ID, "error", err)
			return Bot{}, fmt.Errorf("write manifest: %w", err)
		}
	}
	return stored, nil
}

var ErrValidation = errors.New("validation error")

func validateInput(input UpsertInput) error {
	if strings.TrimSpace(input.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrValidation)
	}
	if strings.TrimSpace(input.AccountID) == "" {
		return fmt.Errorf("%w: account_id is required", ErrValidation)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if strings.TrimSpace(input.Image) == "" {
		return fmt.Errorf("%w: image is required", ErrValidation)
	}
	if len(input.Config) == 0 {
		return fmt.Errorf("%w: config is required", ErrValidation)
	}
	return nil
}

func cloneConfig(cfg json.RawMessage) json.RawMessage {
	if cfg == nil {
		return nil
	}
	var buf bytes.Buffer
	buf.Write(cfg)
	return json.RawMessage(buf.Bytes())
}
