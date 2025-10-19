package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/future-bots/consumer/internal/consumer"
	"github.com/future-bots/platform/config"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	brokersEnv := config.EnvOrDefault("CONSUMER_KAFKA_BROKERS", "localhost:9092")
	brokers := splitAndClean(brokersEnv)
	if len(brokers) == 0 {
		logger.Error("no kafka brokers configured", "env", brokersEnv)
		os.Exit(1)
	}

	topic := config.EnvOrDefault("CONSUMER_KAFKA_TOPIC", "ssi_ps")
	groupID := config.EnvOrDefault("CONSUMER_KAFKA_GROUP", "ssi_ps_consumer")

	redisAddr := config.EnvOrDefault("CONSUMER_REDIS_ADDR", "localhost:6379")
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping failed", "addr", redisAddr, "error", err)
		os.Exit(1)
	}

	svc, err := consumer.New(consumer.Config{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		RedisKeyFmt:    config.EnvOrDefault("CONSUMER_REDIS_KEY_FMT", "ssi_ps:%s"),
		RedisStreamFmt: config.EnvOrDefault("CONSUMER_REDIS_STREAM_FMT", "ssi_ps_stream:%s"),
		MetricPrefix:   os.Getenv("CONSUMER_METRIC_PREFIX"),
	}, redisClient)
	if err != nil {
		logger.Error("failed to init consumer", "error", err)
		os.Exit(1)
	}
	defer svc.Close()

	logger.Info("consumer started", "topic", topic, "group", groupID, "redis", redisAddr)
	if err := svc.Run(ctx); err != nil {
		logger.Error("consumer exited with error", "error", err)
		os.Exit(1)
	}
	logger.Info("consumer stopped")
}

func splitAndClean(csv string) []string {
	parts := strings.Split(csv, ",")
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			cleaned = append(cleaned, v)
		}
	}
	return cleaned
}
