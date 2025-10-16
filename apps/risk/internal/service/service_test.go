package service

import (
	"context"
	"testing"
	"time"
)

type stubRepo struct {
	limit RiskLimits
	err   error
}

func (s stubRepo) FetchLimits(context.Context, string, string, string) (RiskLimits, error) {
	return s.limit, s.err
}

func TestEvaluateAllowsWhenWithinLimits(t *testing.T) {
	repo := stubRepo{limit: RiskLimits{MaxQuantity: 10}}
	svc := New(repo, func() time.Time { return time.Unix(0, 0).UTC() })

	decision, err := svc.Evaluate(context.Background(), RiskCheckRequest{ProposedQty: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("expected decision to allow order")
	}
	if decision.Reason != "" {
		t.Fatalf("expected empty reason got %s", decision.Reason)
	}
}

func TestEvaluateRejectsWhenExceedingLimits(t *testing.T) {
	repo := stubRepo{limit: RiskLimits{MaxQuantity: 10}}
	svc := New(repo, func() time.Time { return time.Unix(0, 0).UTC() })

	decision, err := svc.Evaluate(context.Background(), RiskCheckRequest{ProposedQty: 12})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Allowed {
		t.Fatalf("expected decision to reject order")
	}
	if decision.Reason == "" {
		t.Fatalf("expected reason to be populated")
	}
}

func TestEvaluateRejectsInvalidQuantity(t *testing.T) {
	repo := stubRepo{limit: RiskLimits{MaxQuantity: 10}}
	svc := New(repo, func() time.Time { return time.Unix(0, 0).UTC() })

	if _, err := svc.Evaluate(context.Background(), RiskCheckRequest{ProposedQty: 0}); err == nil {
		t.Fatalf("expected error for zero quantity")
	}
}
