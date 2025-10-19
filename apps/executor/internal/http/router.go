package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/future-bots/executor/internal/service"
	"github.com/future-bots/platform/httpx"
)

// NewRouter constructs an HTTP handler exposing the executor API surface.
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

	mux.HandleFunc("POST /api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		var intent service.OrderIntent
		if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
			logger.Error("invalid order intent", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}

		order, err := svc.SubmitOrder(r.Context(), intent)
		if err != nil {
			var ve service.ValidationError
			if errors.As(err, &ve) {
				httpx.Error(w, http.StatusBadRequest, ve.Error())
				return
			}
			logger.Error("order submission failed", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to submit order")
			return
		}

		logger.Info("accepted order intent", "bot_id", intent.BotID, "symbol", intent.Symbol, "side", intent.Side)
		httpx.JSON(w, http.StatusAccepted, order)
	})

	mux.HandleFunc("GET /api/v1/orders/{order_id}", func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("order_id")
		order, err := svc.GetOrder(r.Context(), orderID)
		if err != nil {
			if errors.Is(err, service.ErrOrderNotFound) {
				httpx.Error(w, http.StatusNotFound, "order not found")
				return
			}
			logger.Error("failed to fetch order", "order_id", orderID, "error", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to fetch order")
			return
		}
		httpx.JSON(w, http.StatusOK, order)
	})

	return mux
}
