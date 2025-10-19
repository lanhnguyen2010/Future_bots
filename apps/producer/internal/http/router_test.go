package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	producerhttp "github.com/future-bots/producer/internal/http"
	"github.com/future-bots/producer/internal/producer"
)

type stubProducer struct {
	last producer.Message
	err  error
}

func (s *stubProducer) Produce(_ context.Context, msg producer.Message) error {
	s.last = msg
	return s.err
}

func TestProduceMessage(t *testing.T) {
	stub := &stubProducer{}
	router := producerhttp.NewRouter(nil, stub)

	body, _ := json.Marshal(producerhttp.ProduceRequest{Value: "hello", Key: "k", Topic: "orders", Headers: map[string]string{"source": "test"}})
	req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/messages", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusAccepted {
		t.Fatalf("expected 202 got %d", rec.Code)
	}
	if stub.last.Topic != "orders" {
		t.Fatalf("expected topic orders got %s", stub.last.Topic)
	}
	if string(stub.last.Key) != "k" {
		t.Fatalf("expected key k got %s", string(stub.last.Key))
	}
	if len(stub.last.Headers) != 1 {
		t.Fatalf("expected header captured")
	}
}

func TestProduceValidation(t *testing.T) {
	stub := &stubProducer{}
	router := producerhttp.NewRouter(nil, stub)

	req := httptest.NewRequest(nethttp.MethodPost, "/api/v1/messages", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected 400 for malformed body got %d", rec.Code)
	}

	req = httptest.NewRequest(nethttp.MethodPost, "/api/v1/messages", bytes.NewReader([]byte(`{"topic":"t"}`)))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected 400 when value missing got %d", rec.Code)
	}

	stub.err = producer.ErrNoTopic
	body, _ := json.Marshal(producerhttp.ProduceRequest{Value: "hello"})
	req = httptest.NewRequest(nethttp.MethodPost, "/api/v1/messages", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected 400 when service returns ErrNoTopic got %d", rec.Code)
	}
}
