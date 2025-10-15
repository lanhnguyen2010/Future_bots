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
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
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
	router := NewRouter(newTestLogger())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBufferString("{")))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed payload got %d", rr.Code)
	}

	invalid := OrderIntent{BotID: " ", Symbol: "", Side: "hold", Quantity: -1}
	body, _ := json.Marshal(invalid)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid payload got %d", rr.Code)
	}

	valid := OrderIntent{BotID: "bot-1", Symbol: "VN30F1M", Side: "buy", Quantity: 1, Price: 1400}
	body, _ = json.Marshal(valid)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body)))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for valid payload got %d", rr.Code)
	}
	var status OrderStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode status: %v", err)
	}
	if status.BotID != valid.BotID || status.Symbol != valid.Symbol {
		t.Fatalf("unexpected status response: %+v", status)
	}
}

func TestGetOrderStatus(t *testing.T) {
	router := NewRouter(newTestLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/ord-123", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var status OrderStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.ID != "ord-123" {
		t.Fatalf("expected order id ord-123 got %s", status.ID)
	}
}

func TestValidateIntent(t *testing.T) {
	cases := []struct {
		name   string
		intent OrderIntent
		ok     bool
	}{
		{name: "missing bot", intent: OrderIntent{BotID: ""}, ok: false},
		{name: "missing symbol", intent: OrderIntent{BotID: "bot", Symbol: ""}, ok: false},
		{name: "invalid side", intent: OrderIntent{BotID: "bot", Symbol: "SYM", Side: "hold", Quantity: 1}, ok: false},
		{name: "negative qty", intent: OrderIntent{BotID: "bot", Symbol: "SYM", Side: "buy", Quantity: 0}, ok: false},
		{name: "valid", intent: OrderIntent{BotID: "bot", Symbol: "SYM", Side: "sell", Quantity: 1}, ok: true},
	}
	for _, tc := range cases {
		err := validateIntent(tc.intent)
		if tc.ok && err != nil {
			t.Fatalf("%s expected no error got %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s expected error", tc.name)
		}
	}
}
