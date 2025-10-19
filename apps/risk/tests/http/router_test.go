package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	riskhttp "github.com/future-bots/risk/internal/http"
	"github.com/future-bots/risk/internal/repository"
	"github.com/future-bots/risk/internal/service"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newTestRouter(t *testing.T) stdhttp.Handler {
	t.Helper()
	repo := repository.NewMemory(10)
	svc := service.New(repo, func() time.Time { return time.Unix(0, 0).UTC() })
	return riskhttp.NewRouter(newTestLogger(), svc)
}

func TestHealthEndpoints(t *testing.T) {
	router := newTestRouter(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodGet, path, nil))
		if rr.Code != stdhttp.StatusOK {
			t.Fatalf("%s expected status 200 got %d", path, rr.Code)
		}
	}
}

func TestDocsEndpoints(t *testing.T) {
	router := newTestRouter(t)
	for _, tt := range []struct {
		path        string
		contentType string
	}{
		{path: "/openapi.json", contentType: "application/json"},
		{path: "/docs", contentType: "text/html; charset=utf-8"},
	} {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodGet, tt.path, nil))
		if rr.Code != stdhttp.StatusOK {
			t.Fatalf("%s expected status 200 got %d", tt.path, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, tt.contentType) {
			t.Fatalf("%s expected content type %s got %s", tt.path, tt.contentType, ct)
		}
	}
}

func TestEvaluateRiskEndpoint(t *testing.T) {
	router := newTestRouter(t)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/risk/evaluate", bytes.NewBufferString("{")))
	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400 for malformed request got %d", rr.Code)
	}

	req := service.RiskCheckRequest{BotID: "bot-1", AccountID: "acct", Symbol: "VN30F1M", ProposedSide: "buy", ProposedQty: 12}
	body, _ := json.Marshal(req)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/risk/evaluate", bytes.NewReader(body)))
	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var resp service.RiskCheckDecision
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Allowed {
		t.Fatalf("expected rejection when qty exceeds limit")
	}
	if resp.Reason == "" {
		t.Fatalf("expected rejection reason")
	}
}

func TestEvaluateRiskInvalidQuantity(t *testing.T) {
	router := newTestRouter(t)
	body, _ := json.Marshal(service.RiskCheckRequest{ProposedQty: 0})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodPost, "/api/v1/risk/evaluate", bytes.NewReader(body)))
	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rr.Code)
	}
}
