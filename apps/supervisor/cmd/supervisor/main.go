package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/future-bots/platform/config"
	platformdb "github.com/future-bots/platform/db"
	platformredis "github.com/future-bots/platform/redis"
	"github.com/future-bots/platform/server"
	"github.com/future-bots/supervisor/internal/bots"
	"github.com/future-bots/supervisor/internal/http"
	"github.com/future-bots/supervisor/internal/migrations"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	addr := config.EnvOrDefault("SUPERVISOR_ADDR", ":8080")
	shutdownTimeout := config.DurationFromEnv("SUPERVISOR_SHUTDOWN_TIMEOUT", 10*time.Second)

	manifestDir := config.EnvOrDefault("SUPERVISOR_BOT_MANIFEST_DIR", "infra/k8s/bots")

	repo := bots.NewMemoryRepository()
	writer := bots.NewFileManifestWriter(manifestDir)
	service := bots.NewService(repo, writer, logger)

	if addr := os.Getenv("SUPERVISOR_REDIS_ADDR"); addr != "" {
		redisCfg := platformredis.Config{
			Addr:         addr,
			Username:     os.Getenv("SUPERVISOR_REDIS_USERNAME"),
			Password:     os.Getenv("SUPERVISOR_REDIS_PASSWORD"),
			DB:           config.IntFromEnv("SUPERVISOR_REDIS_DB", 0),
			DialTimeout:  config.DurationFromEnv("SUPERVISOR_REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  config.DurationFromEnv("SUPERVISOR_REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: config.DurationFromEnv("SUPERVISOR_REDIS_WRITE_TIMEOUT", 3*time.Second),
		}
		redisClient := platformredis.NewClient(redisCfg)

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if err := redisClient.Ping(pingCtx).Err(); err != nil {
			logger.Error("failed to connect to redis", "addr", redisCfg.Addr, "error", err)
			redisClient.Close()
		} else {
			retention := config.DurationFromEnv("SUPERVISOR_REDIS_METRIC_RETENTION", 30*24*time.Hour)
			service = service.WithTelemetry(bots.NewTimeSeriesTelemetry(platformredis.NewTimeSeries(redisClient), retention))
			logger.Info("redis telemetry enabled", "addr", redisCfg.Addr, "retention", retention)
			defer func() {
				if err := redisClient.Close(); err != nil {
					logger.Warn("failed to close redis client", "error", err)
				}
			}()
		}
		cancel()
	}

	handler := http.NewRouter(logger, service)

	if dsn := os.Getenv("SUPERVISOR_DATABASE_URL"); dsn != "" {
		driverName := config.EnvOrDefault("SUPERVISOR_DATABASE_DRIVER", "pgx")
		migrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := platformdb.RunFromDSN(migrateCtx, driverName, dsn, migrations.Files, migrations.Dir); err != nil {
			logger.Error("failed to run database migrations", "error", err)
			os.Exit(1)
		}
		logger.Info("database migrations applied")
	} else {
		logger.Warn("SUPERVISOR_DATABASE_URL not set, skipping database migrations")
	}

	if err := server.Run(ctx, handler, server.Config{Addr: addr, ShutdownTimeout: shutdownTimeout}, logger); err != nil {
		logger.Error("supervisor service exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info("supervisor service stopped")
}
