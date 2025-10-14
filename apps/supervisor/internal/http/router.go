package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/future-bots/platform/httpx"
)

// BotSummary represents a condensed view of a trading bot.
type BotSummary struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Enabled     bool      `json:"enabled"`
	ConfigRev   int       `json:"config_rev"`
	UpdatedAt   time.Time `json:"updated_at"`
	Phase       string    `json:"phase"`
	Description string    `json:"description"`
}

// UpsertBotRequest models the payload used to create or update a bot desired state.
type UpsertBotRequest struct {
	ID        string          `json:"id"`
	AccountID string          `json:"account_id"`
	Name      string          `json:"name"`
	Image     string          `json:"image"`
	Enabled   bool            `json:"enabled"`
	Config    json.RawMessage `json:"config"`
}

// CommandRequest represents a start/stop command sent from the dashboard.
type CommandRequest struct {
	Type    string `json:"type"`
	Timeout int    `json:"timeout_ms"`
}

// NewRouter wires supervisor specific HTTP handlers.
func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("GET /api/v1/bots", func(w http.ResponseWriter, r *http.Request) {
		bots := sampleBots()
		httpx.JSON(w, http.StatusOK, map[string]any{"items": bots})
	})

	mux.HandleFunc("POST /api/v1/bots", func(w http.ResponseWriter, r *http.Request) {
		var payload UpsertBotRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("failed to decode bot upsert payload", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(payload.ID) == "" {
			httpx.Error(w, http.StatusBadRequest, "id is required")
			return
		}
		logger.Info("received bot upsert request", "bot_id", payload.ID, "enabled", payload.Enabled)
		httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
	})

	mux.HandleFunc("POST /api/v1/bots/{bot_id}/commands", func(w http.ResponseWriter, r *http.Request) {
		botID := r.PathValue("bot_id")
		var payload CommandRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("failed to decode command payload", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if payload.Type == "" {
			httpx.Error(w, http.StatusBadRequest, "type is required")
			return
		}
		logger.Info("received bot command", "bot_id", botID, "type", payload.Type)
		httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
	})

	return mux
}

func sampleBots() []BotSummary {
	return []BotSummary{
		{
			ID:          "bot-1",
			AccountID:   "acct-123",
			Name:        "mean-reversion-alpha",
			Image:       "registry.example.com/bots/mean-reversion:1.2.3",
			Enabled:     true,
			ConfigRev:   3,
			UpdatedAt:   time.Now().UTC(),
			Phase:       "running",
			Description: "Example bot placeholder",
		},
	}
}
