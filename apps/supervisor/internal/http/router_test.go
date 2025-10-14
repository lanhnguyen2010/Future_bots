package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestHealthEndpoints(t *testing.T) {
	router := NewRouter(newTestLogger())
	tests := []struct {
		path string
	}{
		{path: "/healthz"},
		{path: "/readyz"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s expected status 200 got %d", tt.path, rr.Code)
		}
	}
}

func TestListBots(t *testing.T) {
	router := NewRouter(newTestLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bots", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}
	var payload map[string][]BotSummary
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload["items"]) == 0 {
		t.Fatalf("expected sample bots in response")
	}
}

func TestUpsertBotValidation(t *testing.T) {
	router := NewRouter(newTestLogger())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/bots", bytes.NewBufferString("{")))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed payload got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ := json.Marshal(UpsertBotRequest{AccountID: "acct", Name: "bot"})
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/bots", bytes.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when id missing got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ = json.Marshal(UpsertBotRequest{ID: "bot-1", AccountID: "acct", Name: "bot"})
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/bots", bytes.NewReader(body)))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for valid payload got %d", rr.Code)
	}
}

func TestCommandValidation(t *testing.T) {
	router := NewRouter(newTestLogger())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/bots/bot-1/commands", bytes.NewBufferString("{")))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed command got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ := json.Marshal(CommandRequest{})
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/bots/bot-1/commands", bytes.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when type missing got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ = json.Marshal(CommandRequest{Type: "bot.start"})
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/bots/bot-1/commands", bytes.NewReader(body)))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 when command valid got %d", rr.Code)
	}
}
