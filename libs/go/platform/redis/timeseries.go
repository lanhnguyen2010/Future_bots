package redis

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// CmdExecutor captures the subset of go-redis functionality used by the timeseries helpers.
type CmdExecutor interface {
	Do(ctx context.Context, args ...any) *goredis.Cmd
}

// TimeSeries wraps go-redis commands for RedisTimeSeries.
type TimeSeries struct {
	client CmdExecutor
}

// SeriesOptions configure metadata for a Redis time series key.
type SeriesOptions struct {
	Retention       time.Duration
	DuplicatePolicy string
	ChunkSize       int
	Labels          map[string]string
}

// Sample models a single data point returned from RedisTimeSeries.
type Sample struct {
	Timestamp time.Time
	Value     float64
}

// RangeOptions allow narrowing TS.RANGE queries.
type RangeOptions struct {
	Count int64
}

// NewTimeSeries exposes RedisTimeSeries helpers using the provided client.
func NewTimeSeries(client CmdExecutor) *TimeSeries {
	return &TimeSeries{client: client}
}

// Create ensures the given key exists with the provided options. If the series already exists it is treated as success.
func (ts *TimeSeries) Create(ctx context.Context, key string, opts SeriesOptions) error {
	args := []any{"TS.CREATE", key}

	if opts.Retention > 0 {
		args = append(args, "RETENTION", int64(opts.Retention/time.Millisecond))
	}
	if opts.DuplicatePolicy != "" {
		args = append(args, "DUPLICATE_POLICY", strings.ToUpper(opts.DuplicatePolicy))
	}
	if opts.ChunkSize > 0 {
		args = append(args, "CHUNK_SIZE", opts.ChunkSize)
	}
	if len(opts.Labels) > 0 {
		args = append(args, "LABELS")
		for _, kv := range labelsSlice(opts.Labels) {
			args = append(args, kv.key, kv.value)
		}
	}

	if err := ts.client.Do(ctx, args...).Err(); err != nil {
		if isSeriesExistsError(err) {
			return nil
		}
		return fmt.Errorf("ts.create %q: %w", key, err)
	}
	return nil
}

// Add inserts a sample into the given series. When ts is the zero value the server assigns the timestamp.
func (ts *TimeSeries) Add(ctx context.Context, key string, tsStamp time.Time, value float64, labels map[string]string) error {
	timestamp := any("*")
	if !tsStamp.IsZero() {
		timestamp = tsStamp.UnixMilli()
	}

	args := []any{"TS.ADD", key, timestamp, value}
	if len(labels) > 0 {
		args = append(args, "LABELS")
		for _, kv := range labelsSlice(labels) {
			args = append(args, kv.key, kv.value)
		}
	}

	if err := ts.client.Do(ctx, args...).Err(); err != nil {
		return fmt.Errorf("ts.add %q: %w", key, err)
	}
	return nil
}

// IncrBy increments the current value by delta, optionally forcing the timestamp when provided.
func (ts *TimeSeries) IncrBy(ctx context.Context, key string, delta float64, tsStamp time.Time) error {
	args := []any{"TS.INCRBY", key, delta}
	if !tsStamp.IsZero() {
		args = append(args, "TIMESTAMP", tsStamp.UnixMilli())
	}
	if err := ts.client.Do(ctx, args...).Err(); err != nil {
		return fmt.Errorf("ts.incrby %q: %w", key, err)
	}
	return nil
}

// Range retrieves samples between the provided time bounds (inclusive).
// If from is the zero value, "-" is used. If to is zero, "+" is used.
func (ts *TimeSeries) Range(ctx context.Context, key string, from, to time.Time, opts RangeOptions) ([]Sample, error) {
	args := []any{"TS.RANGE", key, rangeBoundary(from, "-"), rangeBoundary(to, "+")}
	if opts.Count > 0 {
		args = append(args, "COUNT", opts.Count)
	}
	cmd := ts.client.Do(ctx, args...)
	result, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("ts.range %q: %w", key, err)
	}
	items, ok := result.([]any)
	if !ok {
		return nil, fmt.Errorf("ts.range %q: unexpected response type %T", key, result)
	}
	samples := make([]Sample, 0, len(items))
	for _, raw := range items {
		entry, ok := raw.([]any)
		if !ok || len(entry) != 2 {
			return nil, fmt.Errorf("ts.range %q: malformed entry %#v", key, raw)
		}
		tsMilli, err := anyToInt64(entry[0])
		if err != nil {
			return nil, fmt.Errorf("ts.range %q: parse timestamp: %w", key, err)
		}
		val, err := anyToFloat64(entry[1])
		if err != nil {
			return nil, fmt.Errorf("ts.range %q: parse value: %w", key, err)
		}
		samples = append(samples, Sample{
			Timestamp: time.UnixMilli(tsMilli).UTC(),
			Value:     val,
		})
	}
	return samples, nil
}

func isSeriesExistsError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, goredis.Nil) || strings.Contains(err.Error(), "TSDB: key already exists")
}

type labelKV struct {
	key   string
	value string
}

func labelsSlice(labels map[string]string) []labelKV {
	items := make([]labelKV, 0, len(labels))
	for k, v := range labels {
		items = append(items, labelKV{key: k, value: v})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].key < items[j].key
	})
	return items
}

func rangeBoundary(ts time.Time, fallback string) any {
	if ts.IsZero() {
		return fallback
	}
	return ts.UnixMilli()
}

func anyToInt64(val any) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unexpected type %T", val)
	}
}

func anyToFloat64(val any) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unexpected type %T", val)
	}
}
