package http_test

import (
	"encoding/json"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	reportshttp "github.com/future-bots/reports/internal/http"
	"github.com/future-bots/reports/internal/repository"
	reportsvc "github.com/future-bots/reports/internal/service"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newTestRouter(t *testing.T) stdhttp.Handler {
	t.Helper()
	repo := repository.NewMemory(0, 0)
	svc := reportsvc.New(repo, "", func() time.Time { return time.Unix(0, 0).UTC() })
	return reportshttp.NewRouter(newTestLogger(), svc)
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

func TestPnLReportEndpoint(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/v1/reports/pnl?account_id=acct-1&bot_id=bot-1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var report reportsvc.PnLReport
	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to decode report: %v", err)
	}
	if report.AccountID != "acct-1" || report.BotID != "bot-1" {
		t.Fatalf("unexpected report identifiers: %+v", report)
	}
}

func TestPnLReportDefaultAccount(t *testing.T) {
	router := newTestRouter(t)
	req := httptest.NewRequest(stdhttp.MethodGet, "/api/v1/reports/pnl", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var report reportsvc.PnLReport
	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to decode report: %v", err)
	}
	if report.AccountID == "" {
		t.Fatalf("expected default account id to be set")
	}
}
