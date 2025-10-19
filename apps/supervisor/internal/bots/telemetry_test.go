package bots

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	platformredis "github.com/future-bots/platform/redis"
)

type fakeExecutor struct {
	commands [][]any
	errors   []error
	index    int
}

func (f *fakeExecutor) Do(ctx context.Context, args ...any) *goredis.Cmd {
	f.commands = append(f.commands, args)
	cmd := goredis.NewCmd(ctx)
	if f.index < len(f.errors) && f.errors[f.index] != nil {
		cmd.SetErr(f.errors[f.index])
	}
	f.index++
	return cmd
}

func TestTimeSeriesTelemetryRecordsSamples(t *testing.T) {
	exec := &fakeExecutor{}
	ts := platformredis.NewTimeSeries(exec)
	telemetry := NewTimeSeriesTelemetry(ts, time.Hour)

	now := time.Unix(1700, 0).UTC()
	bot := Bot{
		ID:        "Bot-Alpha_01",
		AccountID: "acct-1",
		ConfigRev: 3,
		Enabled:   true,
		UpdatedAt: now,
	}

	if err := telemetry.RecordBotUpsert(context.Background(), bot); err != nil {
		t.Fatalf("RecordBotUpsert returned error: %v", err)
	}

	if len(exec.commands) != 4 {
		t.Fatalf("expected 4 redis commands, got %d", len(exec.commands))
	}

	expect := [][]any{
		{"TS.CREATE", "bots:bot_alpha_01:config_rev", "RETENTION", int64(time.Hour / time.Millisecond), "LABELS", "account_id", "acct-1", "bot_id", "Bot-Alpha_01", "metric", "config_revision"},
		{"TS.ADD", "bots:bot_alpha_01:config_rev", now.UnixMilli(), float64(3)},
		{"TS.CREATE", "bots:bot_alpha_01:enabled", "RETENTION", int64(time.Hour / time.Millisecond), "LABELS", "account_id", "acct-1", "bot_id", "Bot-Alpha_01", "metric", "enabled"},
		{"TS.ADD", "bots:bot_alpha_01:enabled", now.UnixMilli(), float64(1)},
	}

	for i, cmd := range exec.commands {
		if len(cmd) != len(expect[i]) {
			t.Fatalf("command %d expected %v got %v", i, expect[i], cmd)
		}
		for j := range cmd {
			if cmd[j] != expect[i][j] {
				t.Fatalf("command %d arg %d expected %v got %v", i, j, expect[i][j], cmd[j])
			}
		}
	}
}

func TestSanitizeKey(t *testing.T) {
	cases := map[string]string{
		"Bot-1":        "bot_1",
		"":             "unknown",
		"--":           "unknown",
		"Bot:Child/42": "bot:child_42",
	}
	for input, want := range cases {
		if got := sanitizeKey(input); got != want {
			t.Fatalf("sanitizeKey(%q) = %q want %q", input, got, want)
		}
	}
}
