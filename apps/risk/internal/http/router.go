package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/future-bots/platform/httpx"
)

// RiskCheckRequest represents a request to evaluate a risk exposure.
type RiskCheckRequest struct {
	BotID        string  `json:"bot_id"`
	AccountID    string  `json:"account_id"`
	Symbol       string  `json:"symbol"`
	ProposedSide string  `json:"proposed_side"`
	ProposedQty  float64 `json:"proposed_qty"`
}

// RiskCheckResponse holds the risk decision for a given request.
type RiskCheckResponse struct {
	Allowed   bool      `json:"allowed"`
	Reason    string    `json:"reason"`
	CheckedAt time.Time `json:"checked_at"`
}

// NewRouter creates the risk HTTP API routes.
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

	mux.HandleFunc("POST /api/v1/risk/evaluate", func(w http.ResponseWriter, r *http.Request) {
		var req RiskCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("invalid risk evaluation request", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}

		decision := evaluate(req)
		logger.Info("risk decision computed", "bot_id", req.BotID, "allowed", decision.Allowed)
		httpx.JSON(w, http.StatusOK, decision)
	})

	return mux
}

func evaluate(req RiskCheckRequest) RiskCheckResponse {
	// Placeholder evaluation logic. Real implementation will inspect exposure
	// limits, PnL, and other guardrails. For now, reject orders above 10 lots.
	allowed := req.ProposedQty <= 10
	response := RiskCheckResponse{
		Allowed:   allowed,
		Reason:    "",
		CheckedAt: time.Now().UTC(),
	}
	if !allowed {
		response.Reason = "quantity exceeds maximum lot size"
	}
	return response
}
