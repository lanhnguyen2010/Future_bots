package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/future-bots/platform/httpx"
	"github.com/future-bots/supervisor/internal/bots"
)

// UpsertBotRequest models the payload used to create or update a bot desired state.
type UpsertBotRequest struct {
	ID          string          `json:"id"`
	AccountID   string          `json:"account_id"`
	Name        string          `json:"name"`
	Image       string          `json:"image"`
	Enabled     bool            `json:"enabled"`
	Config      json.RawMessage `json:"config"`
	Description string          `json:"description"`
}

// CommandRequest represents a start/stop command sent from the dashboard.
type CommandRequest struct {
	Type    string `json:"type"`
	Timeout int    `json:"timeout_ms"`
}

// BotService abstracts bot operations required by the HTTP layer.
type BotService interface {
	ListBots(ctx context.Context) ([]bots.Bot, error)
	UpsertBot(ctx context.Context, input bots.UpsertInput) (bots.Bot, error)
}

// NewRouter wires supervisor specific HTTP handlers.
func NewRouter(logger *slog.Logger, svc BotService) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /openapi.json", serveOpenAPI)
	mux.HandleFunc("GET /docs", serveDocs)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("GET /api/v1/bots", func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.ListBots(r.Context())
		if err != nil {
			logger.Error("failed to list bots", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to list bots")
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
	})

	mux.HandleFunc("POST /api/v1/bots", func(w http.ResponseWriter, r *http.Request) {
		var payload UpsertBotRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("failed to decode bot upsert payload", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
		bot, err := svc.UpsertBot(r.Context(), bots.UpsertInput{
			ID:          payload.ID,
			AccountID:   payload.AccountID,
			Name:        payload.Name,
			Image:       payload.Image,
			Enabled:     payload.Enabled,
			Config:      payload.Config,
			Description: payload.Description,
		})
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, bots.ErrValidation) {
				status = http.StatusBadRequest
			}
			logger.Error("failed to upsert bot", "bot_id", payload.ID, "error", err)
			httpx.Error(w, status, err.Error())
			return
		}
		httpx.JSON(w, http.StatusAccepted, bot)
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
