package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/future-bots/supervisor/internal/bots"
	supervisorhttp "github.com/future-bots/supervisor/internal/http"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newRouter(t *testing.T) stdhttp.Handler {
	t.Helper()
	repo := bots.NewMemoryRepository()
	writer := bots.NewFileManifestWriter(t.TempDir())
	svc := bots.NewService(repo, writer, newTestLogger())
	_, err := svc.UpsertBot(context.Background(), bots.UpsertInput{
		ID:        "bot-1",
		AccountID: "acct-1",
		Name:      "sample",
		Image:     "registry.example.com/sample:latest",
		Enabled:   true,
		Config:    json.RawMessage(`{"threshold":1}`),
	})
	if err != nil {
		t.Fatalf("seed bot: %v", err)
	}
	return supervisorhttp.NewRouter(newTestLogger(), svc)
}

func TestHealthEndpoints(t *testing.T) {
	router := newRouter(t)
	tests := []struct {
		path string
	}{
		{path: "/healthz"},
		{path: "/readyz"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(stdhttp.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != stdhttp.StatusOK {
			t.Fatalf("%s expected status 200 got %d", tt.path, rr.Code)
		}
	}
}

func TestDocsEndpoints(t *testing.T) {
	router := newRouter(t)
	for _, tt := range []struct {
		path        string
		contentType string
	}{
		{path: "/openapi.json", contentType: "application/json"},
		{path: "/docs", contentType: "text/html; charset=utf-8"},
	} {
		req := httptest.NewRequest(stdhttp.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != stdhttp.StatusOK {
			t.Fatalf("%s expected status 200 got %d", tt.path, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, tt.contentType) {
			t.Fatalf("%s expected content type %s got %s", tt.path, tt.contentType, ct)
		}
	}
}

func TestListBots(t *testing.T) {
	router := newRouter(t)
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/v1/bots", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}
	var payload map[string][]bots.Bot
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload["items"]) != 1 {
		t.Fatalf("expected seeded bot in response")
	}
}

func TestUpsertBotValidation(t *testing.T) {
	router := newRouter(t)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/bots", bytes.NewBufferString("{")))
	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400 for malformed payload got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ := json.Marshal(supervisorhttp.UpsertBotRequest{AccountID: "acct", Name: "bot", Config: json.RawMessage(`{}`), Image: "img"})
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/bots", bytes.NewReader(body)))
	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400 when id missing got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ = json.Marshal(supervisorhttp.UpsertBotRequest{ID: "bot-1", AccountID: "acct", Name: "bot", Image: "img", Config: json.RawMessage(`{}`)})
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/bots", bytes.NewReader(body)))
	if rr.Code != stdhttp.StatusAccepted {
		t.Fatalf("expected 202 for valid payload got %d", rr.Code)
	}
}

func TestCommandValidation(t *testing.T) {
	router := newRouter(t)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/bots/bot-1/commands", bytes.NewBufferString("{")))
	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400 for malformed command got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ := json.Marshal(supervisorhttp.CommandRequest{})
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/bots/bot-1/commands", bytes.NewReader(body)))
	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400 when type missing got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	body, _ = json.Marshal(supervisorhttp.CommandRequest{Type: "bot.start"})
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/bots/bot-1/commands", bytes.NewReader(body)))
	if rr.Code != stdhttp.StatusAccepted {
		t.Fatalf("expected 202 when command valid got %d", rr.Code)
	}
}
