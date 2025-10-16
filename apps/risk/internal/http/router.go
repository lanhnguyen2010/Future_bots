package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/future-bots/platform/httpx"
	"github.com/future-bots/risk/internal/service"
)

// NewRouter creates the risk HTTP API routes.
func NewRouter(logger *slog.Logger, svc service.Service) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /openapi.json", serveOpenAPI)
	mux.HandleFunc("GET /docs", serveDocs)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST /api/v1/risk/evaluate", func(w http.ResponseWriter, r *http.Request) {
		var req service.RiskCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("invalid risk evaluation request", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}

		decision, err := svc.Evaluate(r.Context(), req)
		if err != nil {
			if errors.Is(err, service.ErrInvalidQuantity) {
				httpx.Error(w, http.StatusBadRequest, err.Error())
				return
			}
			logger.Error("risk evaluation failed", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to evaluate risk")
			return
		}

		logger.Info("risk decision computed", "bot_id", req.BotID, "allowed", decision.Allowed)
		httpx.JSON(w, http.StatusOK, decision)
	})

	return mux
}
