package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/future-bots/executor/internal/repository"
	"github.com/future-bots/executor/internal/service"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newTestRouter(t *testing.T) (http.Handler, *repository.Memory) {
	t.Helper()
	repo := repository.NewMemory()
	svc := service.New(repo, func() time.Time { return time.Unix(0, 0).UTC() })
	return NewRouter(newTestLogger(), svc), repo
}

func TestHealthEndpoints(t *testing.T) {
	router, _ := newTestRouter(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s expected status 200 got %d", path, rr.Code)
		}
	}
}

func TestDocsEndpoints(t *testing.T) {
	router, _ := newTestRouter(t)
	for _, tt := range []struct {
		path        string
		contentType string
	}{
		{path: "/openapi.json", contentType: "application/json"},
		{path: "/docs", contentType: "text/html; charset=utf-8"},
	} {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s expected status 200 got %d", tt.path, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, tt.contentType) {
			t.Fatalf("%s expected content type %s got %s", tt.path, tt.contentType, ct)
		}
	}
}

func TestCreateOrderValidations(t *testing.T) {
	router, _ := newTestRouter(t)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString("{")))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed payload got %d", rr.Code)
	}

	invalid := service.OrderIntent{BotID: " ", Symbol: "", Side: "hold", Quantity: -1}
	body, _ := json.Marshal(invalid)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid payload got %d", rr.Code)
	}

	valid := service.OrderIntent{BotID: "bot-1", Symbol: "VN30F1M", Side: "buy", Quantity: 1, Price: 1400}
	body, _ = json.Marshal(valid)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body)))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for valid payload got %d", rr.Code)
	}
	var status service.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode status: %v", err)
	}
	if status.BotID != valid.BotID || status.Symbol != valid.Symbol {
		t.Fatalf("unexpected status response: %+v", status)
	}
}

func TestGetOrderStatus(t *testing.T) {
	router, repo := newTestRouter(t)

	order := service.Order{
		ID:        "ord-123",
		BotID:     "bot-1",
		Symbol:    "VN30F1M",
		Side:      "buy",
		Quantity:  1,
		Price:     1425.5,
		Status:    "filled",
		UpdatedAt: time.Unix(0, 0).UTC(),
	}
	if err := repo.Create(context.Background(), order); err != nil {
		t.Fatalf("failed to seed repository: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/ord-123", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var status service.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.ID != "ord-123" {
		t.Fatalf("expected order id ord-123 got %s", status.ID)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	router, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/missing", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rr.Code)
	}
}
