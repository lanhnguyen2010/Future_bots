package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/future-bots/platform/httpx"
)

// OrderIntent represents the payload required to submit an order from a bot.
type OrderIntent struct {
	BotID    string  `json:"bot_id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
}

// OrderStatus describes the status of an order after processing.
type OrderStatus struct {
	ID        string    `json:"id"`
	BotID     string    `json:"bot_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRouter constructs an HTTP handler exposing the executor API surface.
func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST /api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		var intent OrderIntent
		if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
			logger.Error("invalid order intent", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := validateIntent(intent); err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		logger.Info("accepted order intent", "bot_id", intent.BotID, "symbol", intent.Symbol, "side", intent.Side)
		status := OrderStatus{
			ID:        "ord-" + time.Now().UTC().Format("20060102150405"),
			BotID:     intent.BotID,
			Symbol:    intent.Symbol,
			Side:      intent.Side,
			Quantity:  intent.Quantity,
			Price:     intent.Price,
			Status:    "accepted",
			UpdatedAt: time.Now().UTC(),
		}
		httpx.JSON(w, http.StatusAccepted, status)
	})

	mux.HandleFunc("GET /api/v1/orders/{order_id}", func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("order_id")
		status := OrderStatus{
			ID:        orderID,
			BotID:     "bot-1",
			Symbol:    "VN30F1M",
			Side:      "buy",
			Quantity:  1,
			Price:     1425.5,
			Status:    "filled",
			UpdatedAt: time.Now().UTC(),
		}
		httpx.JSON(w, http.StatusOK, status)
	})

	return mux
}

func validateIntent(intent OrderIntent) error {
	if strings.TrimSpace(intent.BotID) == "" {
		return errors.New("bot_id is required")
	}
	if strings.TrimSpace(intent.Symbol) == "" {
		return errors.New("symbol is required")
	}
	if intent.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	if intent.Side != "buy" && intent.Side != "sell" {
		return errors.New("side must be buy or sell")
	}
	return nil
}
