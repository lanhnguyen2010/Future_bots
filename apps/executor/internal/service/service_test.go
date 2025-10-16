package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubRepo struct {
	stored Order
	err    error
	getErr error
}

func (s *stubRepo) Create(_ context.Context, order Order) error {
	if s.err != nil {
		return s.err
	}
	s.stored = order
	return nil
}

func (s *stubRepo) Get(_ context.Context, id string) (Order, error) {
	if s.getErr != nil {
		return Order{}, s.getErr
	}
	if s.stored.ID != "" && s.stored.ID == id {
		return s.stored, nil
	}
	return Order{}, ErrOrderNotFound
}

func TestSubmitOrderValidatesIntent(t *testing.T) {
	repo := &stubRepo{}
	svc := New(repo, func() time.Time { return time.Unix(0, 0).UTC() })

	_, err := svc.SubmitOrder(context.Background(), OrderIntent{})
	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected validation error got %v", err)
	}
}

func TestSubmitOrderPersistsAndReturnsOrder(t *testing.T) {
	repo := &stubRepo{}
	svc := New(repo, func() time.Time { return time.Unix(0, 0).UTC() })

	order, err := svc.SubmitOrder(context.Background(), OrderIntent{
		BotID:    "bot-1",
		Symbol:   "SYM",
		Side:     "buy",
		Quantity: 1,
		Price:    100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID == "" {
		t.Fatalf("expected order id to be populated")
	}
	if repo.stored.ID != order.ID {
		t.Fatalf("expected order to be stored")
	}
	if order.Status != "accepted" {
		t.Fatalf("expected status accepted got %s", order.Status)
	}
	if !order.UpdatedAt.Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("expected deterministic timestamp got %v", order.UpdatedAt)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	repo := &stubRepo{}
	svc := New(repo, func() time.Time { return time.Unix(0, 0).UTC() })

	_, err := svc.GetOrder(context.Background(), "missing")
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound got %v", err)
	}
}
