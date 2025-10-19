package bots

import (
	"context"
	"fmt"
	"strings"
	"time"

	platformredis "github.com/future-bots/platform/redis"
)

// Telemetry captures bot state changes for external systems.
type Telemetry interface {
	RecordBotUpsert(ctx context.Context, bot Bot) error
}

// TelemetryFunc allows using bare functions as telemetry implementations.
type TelemetryFunc func(context.Context, Bot) error

// RecordBotUpsert implements Telemetry.
func (fn TelemetryFunc) RecordBotUpsert(ctx context.Context, bot Bot) error {
	if fn == nil {
		return nil
	}
	return fn(ctx, bot)
}

// TimeSeriesTelemetry persists bot configuration deltas to RedisTimeSeries.
type TimeSeriesTelemetry struct {
	series    *platformredis.TimeSeries
	retention time.Duration
	keyPrefix string
}

// NewTimeSeriesTelemetry wires RedisTimeSeries as the telemetry backend.
// If retention is zero, a 90 day retention window is applied.
func NewTimeSeriesTelemetry(series *platformredis.TimeSeries, retention time.Duration) *TimeSeriesTelemetry {
	if retention <= 0 {
		retention = 90 * 24 * time.Hour
	}
	return &TimeSeriesTelemetry{
		series:    series,
		retention: retention,
		keyPrefix: "bots",
	}
}

// RecordBotUpsert writes bot config revisions & enabled state to RedisTimeSeries.
func (t *TimeSeriesTelemetry) RecordBotUpsert(ctx context.Context, bot Bot) error {
	if t == nil || t.series == nil {
		return nil
	}

	now := bot.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	configKey := fmt.Sprintf("%s:%s:config_rev", t.keyPrefix, sanitizeKey(bot.ID))
	if err := t.series.Create(ctx, configKey, platformredis.SeriesOptions{
		Retention: t.retention,
		Labels: map[string]string{
			"bot_id":     bot.ID,
			"account_id": bot.AccountID,
			"metric":     "config_revision",
		},
	}); err != nil {
		return fmt.Errorf("create config series: %w", err)
	}
	if err := t.series.Add(ctx, configKey, now, float64(bot.ConfigRev), nil); err != nil {
		return fmt.Errorf("append config sample: %w", err)
	}

	enabledKey := fmt.Sprintf("%s:%s:enabled", t.keyPrefix, sanitizeKey(bot.ID))
	if err := t.series.Create(ctx, enabledKey, platformredis.SeriesOptions{
		Retention: t.retention,
		Labels: map[string]string{
			"bot_id":     bot.ID,
			"account_id": bot.AccountID,
			"metric":     "enabled",
		},
	}); err != nil {
		return fmt.Errorf("create enabled series: %w", err)
	}

	value := 0.0
	if bot.Enabled {
		value = 1
	}
	if err := t.series.Add(ctx, enabledKey, now, value, nil); err != nil {
		return fmt.Errorf("append enabled sample: %w", err)
	}

	return nil
}

func sanitizeKey(id string) string {
	if id == "" {
		return "unknown"
	}
	id = strings.ToLower(id)
	var b strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ':' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	out := b.String()
	out = strings.Trim(out, "_")
	if out == "" {
		return "unknown"
	}
	return out
}
