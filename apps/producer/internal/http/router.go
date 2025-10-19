package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/future-bots/platform/httpx"
	"github.com/future-bots/producer/internal/producer"
)

// Producer defines the behaviour expected from the producer service.
type Producer interface {
	Produce(ctx context.Context, msg producer.Message) error
}

// ProduceRequest captures the POST payload for publishing a message.
type ProduceRequest struct {
	Topic   string            `json:"topic"`
	Key     string            `json:"key"`
	Value   string            `json:"value"`
	Headers map[string]string `json:"headers"`
}

// NewRouter wires the HTTP endpoints for the producer service.
func NewRouter(logger *slog.Logger, svc Producer) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		var payload ProduceRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Error("decode payload failed", "error", err)
			httpx.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if payload.Value == "" {
			httpx.Error(w, http.StatusBadRequest, "value is required")
			return
		}

		msg := producer.Message{
			Topic: payload.Topic,
			Key:   []byte(payload.Key),
			Value: []byte(payload.Value),
		}
		for k, v := range payload.Headers {
			msg.Headers = append(msg.Headers, producer.Header{Key: k, Value: []byte(v)})
		}

		if err := svc.Produce(r.Context(), msg); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, producer.ErrNoTopic) {
				status = http.StatusBadRequest
			}
			logger.Error("produce message failed", "error", err)
			httpx.Error(w, status, err.Error())
			return
		}

		httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
	})

	return mux
}
