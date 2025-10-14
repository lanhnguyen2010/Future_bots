package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// EnvOrDefault returns the environment variable value for key if set, otherwise the fallback.
func EnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// DurationFromEnv parses a time.Duration from the given environment variable.
// If the value is empty or parsing fails, the fallback duration is returned.
func DurationFromEnv(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return dur
}

// MustGetEnv returns the environment variable for the provided key.
// It panics if the variable is not defined.
func MustGetEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("environment variable %s must be set", key))
}

// IntFromEnv parses an integer from the given environment variable key.
// If parsing fails the fallback value is returned.
func IntFromEnv(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}
