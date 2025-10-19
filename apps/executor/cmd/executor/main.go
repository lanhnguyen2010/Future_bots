package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/future-bots/executor/internal/http"
	"github.com/future-bots/executor/internal/migrations"
	"github.com/future-bots/executor/internal/repository"
	"github.com/future-bots/executor/internal/service"
	"github.com/future-bots/platform/config"
	platformdb "github.com/future-bots/platform/db"
	"github.com/future-bots/platform/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	addr := config.EnvOrDefault("EXECUTOR_ADDR", ":8081")
	shutdownTimeout := config.DurationFromEnv("EXECUTOR_SHUTDOWN_TIMEOUT", 10*time.Second)

	repo := repository.NewMemory()
	svc := service.New(repo, nil)
	handler := http.NewRouter(logger, svc)

	if dsn := os.Getenv("EXECUTOR_DATABASE_URL"); dsn != "" {
		driverName := config.EnvOrDefault("EXECUTOR_DATABASE_DRIVER", "pgx")
		migrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := platformdb.RunFromDSN(migrateCtx, driverName, dsn, migrations.Files, migrations.Dir); err != nil {
			logger.Error("failed to run database migrations", "error", err)
			os.Exit(1)
		}
		logger.Info("database migrations applied")
	} else {
		logger.Warn("EXECUTOR_DATABASE_URL not set, skipping database migrations")
	}

	if err := server.Run(ctx, handler, server.Config{Addr: addr, ShutdownTimeout: shutdownTimeout}, logger); err != nil {
		logger.Error("executor service exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info("executor service stopped")
}
