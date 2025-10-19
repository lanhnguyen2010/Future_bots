package consumer

import (
	"context"
	"fmt"
	"time"

	marketsv1 "github.com/future-bots/proto/markets/v1"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// RedisWriter defines the subset of redis client used for sorted-set writes.
type RedisWriter interface {
	ZAddArgs(ctx context.Context, key string, args redis.ZAddArgs) *redis.IntCmd
	XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd
}

// Consumer streams Kafka messages and persists them to RedisTimeSeries.
type Consumer struct {
	reader    *kafka.Reader
	redis     RedisWriter
	keyFmt    string
	streamFmt string
	jsonOpts  protojson.MarshalOptions
}

// Config captures runtime settings for the consumer.
type Config struct {
	Brokers        []string
	Topic          string
	GroupID        string
	RedisKeyFmt    string // e.g. "ssi_ps:%s"
	RedisStreamFmt string // e.g. "ssi_ps_stream:%s"
	MetricPrefix   string // optional metric namespace
}

// New creates a consumer instance.
func New(cfg Config, redis RedisWriter) (*Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("brokers required")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("topic required")
	}
	if redis == nil {
		return nil, fmt.Errorf("redis client required")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 1,
		MaxBytes: 10 << 20,
	})

	prefix := cfg.MetricPrefix
	if prefix == "" {
		prefix = "ssi_ps"
	}

	keyFmt := cfg.RedisKeyFmt
	if keyFmt == "" {
		keyFmt = fmt.Sprintf("%s:%s", prefix, "%s")
	}

	streamFmt := cfg.RedisStreamFmt
	if streamFmt == "" {
		streamFmt = fmt.Sprintf("%s_stream:%s", prefix, "%s")
	}

	return &Consumer{
		reader:    reader,
		redis:     redis,
		keyFmt:    keyFmt,
		streamFmt: streamFmt,
		jsonOpts: protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseEnumNumbers:  false,
			Multiline:       false,
		},
	}, nil
}

// Close releases reader resources.
func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// Run consumes messages until context cancellation.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		if err := c.handleMessage(ctx, m); err != nil {
			return err
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			return fmt.Errorf("commit message: %w", err)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	var snapshot marketsv1.SsiPsSnapshot
	if err := proto.Unmarshal(msg.Value, &snapshot); err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	key := fmt.Sprintf(c.keyFmt, snapshot.GetCode())
	timestamp := time.Now()
	if ts := snapshot.GetTimestamp(); ts != nil {
		timestamp = ts.AsTime()
	}

	payload, err := c.jsonOpts.Marshal(&snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	zArgs := redis.ZAddArgs{
		NX: true,
		Members: []redis.Z{
			{
				Score:  float64(timestamp.UnixMilli()),
				Member: string(payload),
			},
		},
	}

	if err := c.redis.ZAddArgs(ctx, key, zArgs).Err(); err != nil {
		return fmt.Errorf("redis zadd (%s): %w", key, err)
	}

	streamKey := fmt.Sprintf(c.streamFmt, snapshot.GetCode())
	xArgs := &redis.XAddArgs{
		Stream: streamKey,
		ID:     "*",
		Values: map[string]interface{}{
			"code":      snapshot.GetCode(),
			"board":     snapshot.GetBoard(),
			"timestamp": timestamp.UnixMilli(),
			"payload":   string(payload),
		},
	}

	if err := c.redis.XAdd(ctx, xArgs).Err(); err != nil {
		return fmt.Errorf("redis xadd (%s): %w", streamKey, err)
	}
	return nil
}
