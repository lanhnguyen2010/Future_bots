package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	platformhttpx "github.com/future-bots/platform/httpx"
)

func TestJSONWritesPayload(t *testing.T) {
	rr := httptest.NewRecorder()
	payload := map[string]string{"status": "ok"}
	platformhttpx.JSON(rr, http.StatusCreated, payload)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d got %d", http.StatusCreated, rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json header got %s", got)
	}
	var decoded map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if decoded["status"] != "ok" {
		t.Fatalf("expected payload to contain status ok got %+v", decoded)
	}
}

func TestJSONHandlesNilPayload(t *testing.T) {
	rr := httptest.NewRecorder()
	platformhttpx.JSON(rr, http.StatusNoContent, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d got %d", http.StatusNoContent, rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected no body when payload nil got %q", rr.Body.String())
	}
}

func TestErrorWrapsMessage(t *testing.T) {
	rr := httptest.NewRecorder()
	platformhttpx.Error(rr, http.StatusBadRequest, "boom")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d got %d", http.StatusBadRequest, rr.Code)
	}
	var decoded map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if decoded["error"] != "boom" {
		t.Fatalf("expected boom error got %+v", decoded)
	}
}
