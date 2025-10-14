package server

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestRunReturnsErrorWhenHandlerNil(t *testing.T) {
	err := Run(context.Background(), nil, Config{}, nil)
	if err == nil {
		t.Fatalf("expected error when handler nil")
	}
	if err.Error() != "http handler must not be nil" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := Config{Addr: "localhost:0", ShutdownTimeout: 500 * time.Millisecond}
	logger := slog.New(slog.NewTextHandler(new(noopWriter), nil))

	if err := Run(ctx, handler, cfg, logger); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

type noopWriter struct{}

func (n *noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
