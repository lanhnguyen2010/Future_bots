package redis

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type fakeResponse struct {
	value any
	err   error
}

type fakeExecutor struct {
	responses []fakeResponse
	commands  [][]any
}

func (f *fakeExecutor) Do(ctx context.Context, args ...any) *goredis.Cmd {
	cmd := goredis.NewCmd(ctx)
	f.commands = append(f.commands, args)
	idx := len(f.commands) - 1
	if idx < len(f.responses) {
		resp := f.responses[idx]
		if resp.err != nil {
			cmd.SetErr(resp.err)
		} else {
			cmd.SetVal(resp.value)
		}
	}
	return cmd
}

func TestTimeSeriesRangeParsesSamples(t *testing.T) {
	now := time.UnixMilli(1700).UTC()
	exec := &fakeExecutor{
		responses: []fakeResponse{
			{value: []any{
				[]any{int64(now.UnixMilli()), "1.25"},
				[]any{[]byte("1701"), []byte("2.5")},
			}},
		},
	}
	ts := NewTimeSeries(exec)
	items, err := ts.Range(context.Background(), "series:key", now.Add(-time.Minute), now, RangeOptions{Count: 100})
	if err != nil {
		t.Fatalf("Range returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 samples got %d", len(items))
	}
	if items[0].Timestamp != now {
		t.Fatalf("expected timestamp %v got %v", now, items[0].Timestamp)
	}
	if items[1].Value != 2.5 {
		t.Fatalf("expected value 2.5 got %f", items[1].Value)
	}

	if len(exec.commands) != 1 {
		t.Fatalf("expected 1 command recorded got %d", len(exec.commands))
	}
	cmd := exec.commands[0]
	if cmd[0] != "TS.RANGE" || cmd[1] != "series:key" {
		t.Fatalf("unexpected command: %#v", cmd)
	}
}

func TestMarketSeriesStoreAddTick(t *testing.T) {
	exec := &fakeExecutor{
		responses: []fakeResponse{
			{value: "OK"},     // TS.CREATE price
			{value: int64(1)}, // TS.ADD price
			{value: "OK"},     // TS.CREATE volume
			{value: int64(1)}, // TS.ADD volume
		},
	}
	store := NewMarketSeriesStore(NewTimeSeries(exec), time.Hour)
	tick := MarketTick{
		Ticker:    "VN30F1M",
		Timestamp: time.UnixMilli(2000).UTC(),
		Price:     1234.5,
		Volume:    10,
		Labels: map[string]string{
			"exchange": "hose",
		},
	}
	if err := store.AddTick(context.Background(), tick); err != nil {
		t.Fatalf("AddTick returned error: %v", err)
	}
	if len(exec.commands) != 4 {
		t.Fatalf("expected 4 commands got %d", len(exec.commands))
	}
	if exec.commands[0][0] != "TS.CREATE" {
		t.Fatalf("expected TS.CREATE got %v", exec.commands[0][0])
	}
	if exec.commands[1][0] != "TS.ADD" {
		t.Fatalf("expected TS.ADD got %v", exec.commands[1][0])
	}
	if exec.commands[1][1] != "markets:vn30f1m:price" {
		t.Fatalf("unexpected price key %v", exec.commands[1][1])
	}
	if exec.commands[2][1] != "markets:vn30f1m:volume" {
		t.Fatalf("unexpected volume key %v", exec.commands[2][1])
	}
}
