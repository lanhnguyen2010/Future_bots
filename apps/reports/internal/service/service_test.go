package service

import (
	"context"
	"testing"
	"time"
)

type stubRepo struct {
	data PnLData
	err  error
}

func (s stubRepo) FetchPnL(context.Context, string, string, string) (PnLData, error) {
	return s.data, s.err
}

func TestGeneratePnLReportUsesDefaults(t *testing.T) {
	repo := stubRepo{data: PnLData{Realized: 10, Unrealized: 5}}
	svc := New(repo, "", func() time.Time { return time.Unix(0, 0).UTC() })

	report, err := svc.GeneratePnLReport(context.Background(), PnLQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.AccountID != "acct-123" {
		t.Fatalf("expected default account id, got %s", report.AccountID)
	}
	if report.Window != "1d" {
		t.Fatalf("expected default window, got %s", report.Window)
	}
	if report.Realized != 10 || report.Unrealized != 5 {
		t.Fatalf("unexpected pnl values: %+v", report)
	}
	if !report.Generated.Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("unexpected generated time: %v", report.Generated)
	}
}

func TestGeneratePnLReportRespectsQuery(t *testing.T) {
	repo := stubRepo{data: PnLData{Realized: 20, Unrealized: 8}}
	svc := New(repo, "", func() time.Time { return time.Unix(0, 0).UTC() })

	report, err := svc.GeneratePnLReport(context.Background(), PnLQuery{AccountID: "acct-x", BotID: "bot-1", Window: "1w"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.AccountID != "acct-x" || report.BotID != "bot-1" || report.Window != "1w" {
		t.Fatalf("unexpected identifiers: %+v", report)
	}
	if report.Realized != 20 || report.Unrealized != 8 {
		t.Fatalf("unexpected pnl values: %+v", report)
	}
}
