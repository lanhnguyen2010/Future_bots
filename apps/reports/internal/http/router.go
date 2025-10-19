package http

import (
	"log/slog"
	"net/http"

	"github.com/future-bots/platform/httpx"
	"github.com/future-bots/reports/internal/service"
)

// NewRouter assembles the reports service HTTP routes.
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

	mux.HandleFunc("GET /api/v1/reports/pnl", func(w http.ResponseWriter, r *http.Request) {
		query := service.PnLQuery{
			AccountID: r.URL.Query().Get("account_id"),
			BotID:     r.URL.Query().Get("bot_id"),
			Window:    r.URL.Query().Get("window"),
		}
		report, err := svc.GeneratePnLReport(r.Context(), query)
		if err != nil {
			logger.Error("failed to generate pnl report", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to generate report")
			return
		}
		logger.Info("generated pnl report", "account_id", report.AccountID, "bot_id", report.BotID)
		httpx.JSON(w, http.StatusOK, report)
	})

	return mux
}
