package config_test

import (
	"os"
	"testing"
	"time"

	platformconfig "github.com/future-bots/platform/config"
)

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("FOO", "bar")
	if got := platformconfig.EnvOrDefault("FOO", "baz"); got != "bar" {
		t.Fatalf("expected bar got %s", got)
	}
	if got := platformconfig.EnvOrDefault("MISSING", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback got %s", got)
	}
}

func TestDurationFromEnv(t *testing.T) {
	t.Setenv("DURATION", "5s")
	if got := platformconfig.DurationFromEnv("DURATION", time.Second); got != 5*time.Second {
		t.Fatalf("expected 5s got %s", got)
	}
	t.Setenv("DURATION", "bad")
	if got := platformconfig.DurationFromEnv("DURATION", 2*time.Second); got != 2*time.Second {
		t.Fatalf("expected fallback when parse fails got %s", got)
	}
	os.Unsetenv("DURATION")
	if got := platformconfig.DurationFromEnv("DURATION", 3*time.Second); got != 3*time.Second {
		t.Fatalf("expected fallback when missing got %s", got)
	}
}

func TestMustGetEnv(t *testing.T) {
	t.Setenv("REQUIRED", "value")
	if got := platformconfig.MustGetEnv("REQUIRED"); got != "value" {
		t.Fatalf("expected value got %s", got)
	}
	os.Unsetenv("REQUIRED")
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when env missing")
		}
	}()
	_ = platformconfig.MustGetEnv("REQUIRED")
}

func TestIntFromEnv(t *testing.T) {
	t.Setenv("INT_VALUE", "42")
	if got := platformconfig.IntFromEnv("INT_VALUE", 0); got != 42 {
		t.Fatalf("expected 42 got %d", got)
	}
	t.Setenv("INT_VALUE", "bad")
	if got := platformconfig.IntFromEnv("INT_VALUE", 7); got != 7 {
		t.Fatalf("expected fallback when parse fails got %d", got)
	}
	os.Unsetenv("INT_VALUE")
	if got := platformconfig.IntFromEnv("INT_VALUE", 3); got != 3 {
		t.Fatalf("expected fallback when missing got %d", got)
	}
}
