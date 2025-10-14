package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Config controls how the HTTP server behaves.
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}

// Run starts an HTTP server with the provided handler and blocks until the
// context is cancelled. The server is shut down gracefully using the configured
// timeout.
func Run(ctx context.Context, handler http.Handler, cfg Config, logger *slog.Logger) error {
	if handler == nil {
		return errors.New("http handler must not be nil")
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 10 * time.Second
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	srv := &http.Server{ //nolint:gosec // basic server with explicit address
		Addr:    cfg.Addr,
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		logger.Info("http server shutting down", "timeout", cfg.ShutdownTimeout.String())
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
