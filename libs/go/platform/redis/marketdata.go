package redis

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MarketTick captures a single market data update.
type MarketTick struct {
	Ticker    string
	Timestamp time.Time
	Price     float64
	Volume    float64
	Labels    map[string]string
}

// MarketSeriesStore persists market ticks into RedisTimeSeries.
type MarketSeriesStore struct {
	ts        *TimeSeries
	retention time.Duration
}

// NewMarketSeriesStore constructs a Redis-backed market data store.
func NewMarketSeriesStore(ts *TimeSeries, retention time.Duration) *MarketSeriesStore {
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	}
	return &MarketSeriesStore{
		ts:        ts,
		retention: retention,
	}
}

// AddTick writes price (and optional volume) samples for the provided ticker.
func (m *MarketSeriesStore) AddTick(ctx context.Context, tick MarketTick) error {
	if m == nil || m.ts == nil {
		return fmt.Errorf("timeseries client is not configured")
	}
	if tick.Ticker == "" {
		return fmt.Errorf("ticker is required")
	}
	if tick.Timestamp.IsZero() {
		tick.Timestamp = time.Now().UTC()
	}

	baseLabels := map[string]string{
		"ticker": fmt.Sprintf("%s", tick.Ticker),
	}
	for k, v := range tick.Labels {
		if k == "metric" {
			continue
		}
		baseLabels[k] = v
	}

	priceKey := fmt.Sprintf("markets:%s:price", sanitizeID(tick.Ticker))
	if err := m.ts.Create(ctx, priceKey, SeriesOptions{
		Retention: m.retention,
		Labels:    mergeLabels(baseLabels, map[string]string{"metric": "price"}),
	}); err != nil {
		return err
	}
	if err := m.ts.Add(ctx, priceKey, tick.Timestamp, tick.Price, nil); err != nil {
		return err
	}

	if tick.Volume > 0 {
		volumeKey := fmt.Sprintf("markets:%s:volume", sanitizeID(tick.Ticker))
		if err := m.ts.Create(ctx, volumeKey, SeriesOptions{
			Retention: m.retention,
			Labels:    mergeLabels(baseLabels, map[string]string{"metric": "volume"}),
		}); err != nil {
			return err
		}
		if err := m.ts.Add(ctx, volumeKey, tick.Timestamp, tick.Volume, nil); err != nil {
			return err
		}
	}

	return nil
}

func mergeLabels(base map[string]string, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func sanitizeID(id string) string {
	if id == "" {
		return "unknown"
	}
	id = strings.ToLower(id)
	var b strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ':' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "unknown"
	}
	return out
}
