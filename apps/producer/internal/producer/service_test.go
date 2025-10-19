package producer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/future-bots/producer/internal/producer"
)

type fakeWriter struct {
	messages []kafka.Message
	closed   bool
	err      error
}

func (f *fakeWriter) WriteMessages(_ context.Context, msgs ...kafka.Message) error {
	if f.err != nil {
		return f.err
	}
	f.messages = append(f.messages, msgs...)
	return nil
}

func (f *fakeWriter) Close() error {
	f.closed = true
	return nil
}

func TestServiceProduceUsesDefaultTopic(t *testing.T) {
	writer := &fakeWriter{}
	svc := producer.NewService("orders", func(topic string) (producer.Writer, error) {
		if topic != "orders" {
			t.Fatalf("expected topic orders got %s", topic)
		}
		return writer, nil
	}).WithNow(func() time.Time { return time.Unix(1700, 0).UTC() })

	err := svc.Produce(context.Background(), producer.Message{Value: []byte("hello")})
	if err != nil {
		t.Fatalf("Produce returned error: %v", err)
	}

	if len(writer.messages) != 1 {
		t.Fatalf("expected message to be written")
	}
	if writer.messages[0].Topic != "orders" {
		t.Fatalf("unexpected topic %s", writer.messages[0].Topic)
	}
	if writer.messages[0].Time.Unix() != 1700 {
		t.Fatalf("expected timestamp override")
	}
}

func TestServiceProduceRequiresTopic(t *testing.T) {
	svc := producer.NewService("", nil)
	err := svc.Produce(context.Background(), producer.Message{Value: []byte("test")})
	if !errors.Is(err, producer.ErrNoTopic) {
		t.Fatalf("expected ErrNoTopic got %v", err)
	}
}

func TestServiceClosePropagatesErrors(t *testing.T) {
	closeErr := errors.New("boom")
	svc := producer.NewService("orders", func(topic string) (producer.Writer, error) {
		return &struct{ fakeWriter }{fakeWriter{err: nil}}, nil
	})

	// add writer to map by producing
	_ = svc.Produce(context.Background(), producer.Message{Value: []byte("hello")})

	svc2 := producer.NewService("orders", func(topic string) (producer.Writer, error) {
		return &errWriter{closeErr: closeErr}, nil
	})
	_ = svc2.Produce(context.Background(), producer.Message{Value: []byte("hello")})

	if err := svc2.Close(); !errors.Is(err, closeErr) {
		t.Fatalf("expected close error, got %v", err)
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("expected close success, got %v", err)
	}
}

type errWriter struct {
	fakeWriter
	closeErr error
}

func (e *errWriter) Close() error {
	return e.closeErr
}
