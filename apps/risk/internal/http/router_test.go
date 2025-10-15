package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestHealthEndpoints(t *testing.T) {
	router := NewRouter(newTestLogger())
	for _, path := range []string{"/healthz", "/readyz"} {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s expected status 200 got %d", path, rr.Code)
		}
	}
}

func TestDocsEndpoints(t *testing.T) {
	router := NewRouter(newTestLogger())
	for _, tt := range []struct {
		path        string
		contentType string
	}{
		{path: "/openapi.json", contentType: "application/json"},
		{path: "/docs", contentType: "text/html; charset=utf-8"},
	} {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s expected status 200 got %d", tt.path, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, tt.contentType) {
			t.Fatalf("%s expected content type %s got %s", tt.path, tt.contentType, ct)
		}
	}
}

func TestEvaluateRiskEndpoint(t *testing.T) {
	router := NewRouter(newTestLogger())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/risk/evaluate", bytes.NewBufferString("{")))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed request got %d", rr.Code)
	}

	req := RiskCheckRequest{BotID: "bot-1", AccountID: "acct", Symbol: "VN30F1M", ProposedSide: "buy", ProposedQty: 12}
	body, _ := json.Marshal(req)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/risk/evaluate", bytes.NewReader(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var resp RiskCheckResponse
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

func TestEvaluateHelper(t *testing.T) {
	resp := evaluate(RiskCheckRequest{ProposedQty: 5})
	if !resp.Allowed {
		t.Fatalf("expected allow for qty <= 10")
	}
	resp = evaluate(RiskCheckRequest{ProposedQty: 11})
	if resp.Allowed {
		t.Fatalf("expected rejection when qty > 10")
	}
	if resp.Reason == "" {
		t.Fatalf("expected rejection reason to be populated")
	}
}
