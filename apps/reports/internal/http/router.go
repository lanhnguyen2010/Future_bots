package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/future-bots/platform/httpx"
)

// PnLReport represents a simplified profit and loss report response.
type PnLReport struct {
	AccountID  string    `json:"account_id"`
	BotID      string    `json:"bot_id"`
	Realized   float64   `json:"realized"`
	Unrealized float64   `json:"unrealized"`
	Window     string    `json:"window"`
	Generated  time.Time `json:"generated_at"`
}

// NewRouter assembles the reports service HTTP routes.
func NewRouter(logger *slog.Logger) http.Handler {
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
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			accountID = "acct-123"
		}
		report := PnLReport{
			AccountID:  accountID,
			BotID:      r.URL.Query().Get("bot_id"),
			Realized:   1525.23,
			Unrealized: 210.42,
			Window:     "1d",
			Generated:  time.Now().UTC(),
		}
		logger.Info("generated pnl report", "account_id", report.AccountID, "bot_id", report.BotID)
		httpx.JSON(w, http.StatusOK, report)
	})

	return mux
}
