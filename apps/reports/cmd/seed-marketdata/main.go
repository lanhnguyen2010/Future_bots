package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/future-bots/platform/config"
	platformredis "github.com/future-bots/platform/redis"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	redisAddr := config.EnvOrDefault("MARKETDATA_REDIS_ADDR", "localhost:6379")
	retention := config.DurationFromEnv("MARKETDATA_RETENTION", 7*24*time.Hour)

	redisClient := platformredis.NewClient(platformredis.Config{Addr: redisAddr})
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Warn("failed to close redis client", "error", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping failed", "addr", redisAddr, "error", err)
		os.Exit(1)
	}

	store := platformredis.NewMarketSeriesStore(platformredis.NewTimeSeries(redisClient), retention)

	ticker := config.EnvOrDefault("MARKETDATA_TICKER", "VN30F1M")
	start := time.Now().UTC().Add(-time.Hour)

	samples := []platformredis.MarketTick{}
	for i := 0; i < 12; i++ {
		ts := start.Add(time.Duration(i*5) * time.Minute)
		samples = append(samples, platformredis.MarketTick{
			Ticker:    ticker,
			Timestamp: ts,
			Price:     1000.0 + float64(i)*2.5,
			Volume:    100 + float64(i*5),
			Labels: map[string]string{
				"source": "seed-script",
			},
		})
	}

	for _, tick := range samples {
		if err := store.AddTick(ctx, tick); err != nil {
			logger.Error("failed to add tick", "ticker", ticker, "error", err)
			os.Exit(1)
		}
	}

	logger.Info("seeded market data", "ticker", ticker, "count", len(samples), "retention", retention, "redis_addr", redisAddr)
	fmt.Println("Sample data available under keys:")
	keySuffix := strings.ToLower(ticker)
	fmt.Printf(" - markets:%s:price\n", keySuffix)
	fmt.Printf(" - markets:%s:volume\n", keySuffix)
}
