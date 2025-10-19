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
	"github.com/future-bots/platform/server"
	"github.com/future-bots/risk/internal/http"
	"github.com/future-bots/risk/internal/migrations"
	"github.com/future-bots/risk/internal/repository"
	"github.com/future-bots/risk/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	addr := config.EnvOrDefault("RISK_ADDR", ":8082")
	shutdownTimeout := config.DurationFromEnv("RISK_SHUTDOWN_TIMEOUT", 10*time.Second)

	repo := repository.NewMemory(10)
	svc := service.New(repo, nil)
	handler := http.NewRouter(logger, svc)

	if dsn := os.Getenv("RISK_DATABASE_URL"); dsn != "" {
		driverName := config.EnvOrDefault("RISK_DATABASE_DRIVER", "pgx")
		migrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := platformdb.RunFromDSN(migrateCtx, driverName, dsn, migrations.Files, migrations.Dir); err != nil {
			logger.Error("failed to run database migrations", "error", err)
			os.Exit(1)
		}
		logger.Info("database migrations applied")
	} else {
		logger.Warn("RISK_DATABASE_URL not set, skipping database migrations")
	}

	if err := server.Run(ctx, handler, server.Config{Addr: addr, ShutdownTimeout: shutdownTimeout}, logger); err != nil {
		logger.Error("risk service exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info("risk service stopped")
}
