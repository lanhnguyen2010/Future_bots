package bots_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"log/slog"

	"github.com/future-bots/supervisor/internal/bots"
)

type noopLogger struct{}

func (noopLogger) Enabled(context.Context, slog.Level) bool  { return false }
func (noopLogger) Handle(context.Context, slog.Record) error { return nil }
func (noopLogger) WithAttrs([]slog.Attr) slog.Handler        { return noopLogger{} }
func (noopLogger) WithGroup(string) slog.Handler             { return noopLogger{} }
func newTestLogger() *slog.Logger                            { return slog.New(noopLogger{}) }

func TestServiceUpsertCreatesManifest(t *testing.T) {
	tmp := t.TempDir()
	repo := bots.NewMemoryRepository()
	writer := bots.NewFileManifestWriter(tmp)

	now := time.Unix(0, 0).UTC()
	svc := bots.NewService(repo, writer, newTestLogger()).WithNow(func() time.Time { return now })

	cfg := json.RawMessage(`{"threshold": 1}`)
	bot, err := svc.UpsertBot(context.Background(), bots.UpsertInput{
		ID:        "Bot-1",
		AccountID: "acct-1",
		Name:      "test-bot",
		Image:     "registry.example.com/bot:latest",
		Enabled:   true,
		Config:    cfg,
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	if bot.ConfigRev != 1 {
		t.Fatalf("expected config rev 1 got %d", bot.ConfigRev)
	}
	if bot.Phase != "desired" {
		t.Fatalf("expected phase desired got %s", bot.Phase)
	}

	path := filepath.Join(tmp, "bot-1.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected manifest file %s: %v", path, err)
	}
}

func TestUpsertValidation(t *testing.T) {
	repo := bots.NewMemoryRepository()
	svc := bots.NewService(repo, nil, newTestLogger())

	_, err := svc.UpsertBot(context.Background(), bots.UpsertInput{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
