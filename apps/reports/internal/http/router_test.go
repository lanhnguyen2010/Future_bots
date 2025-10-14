package http

import (
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
	for _, path := range []string{"/healthz", "/readyz"} {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s expected status 200 got %d", path, rr.Code)
		}
	}
}

func TestPnLReportEndpoint(t *testing.T) {
	router := NewRouter(newTestLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/pnl?account_id=acct-1&bot_id=bot-1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var report PnLReport
	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to decode report: %v", err)
	}
	if report.AccountID != "acct-1" || report.BotID != "bot-1" {
		t.Fatalf("unexpected report identifiers: %+v", report)
	}
}

func TestPnLReportDefaultAccount(t *testing.T) {
	router := NewRouter(newTestLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/pnl", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var report PnLReport
	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to decode report: %v", err)
	}
	if report.AccountID == "" {
		t.Fatalf("expected default account id to be set")
	}
}
