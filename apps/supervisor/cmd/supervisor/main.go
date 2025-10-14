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
	"github.com/future-bots/supervisor/internal/http"
	"github.com/future-bots/supervisor/internal/migrations"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	addr := config.EnvOrDefault("SUPERVISOR_ADDR", ":8080")
	shutdownTimeout := config.DurationFromEnv("SUPERVISOR_SHUTDOWN_TIMEOUT", 10*time.Second)

	handler := http.NewRouter(logger)

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
