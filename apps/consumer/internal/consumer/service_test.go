package consumer

import (
	"context"
	"testing"
	"time"

	marketsv1 "github.com/future-bots/proto/markets/v1"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeRedis struct {
	key        string
	args       redis.ZAddArgs
	stream     string
	streamArgs *redis.XAddArgs
	err        error
}

func (f *fakeRedis) ZAddArgs(ctx context.Context, key string, args redis.ZAddArgs) *redis.IntCmd {
	f.key = key
	f.args = args
	cmd := redis.NewIntCmd(ctx)
	if f.err != nil {
		cmd.SetErr(f.err)
	}
	return cmd
}

func (f *fakeRedis) XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	f.stream = a.Stream
	f.streamArgs = a
	cmd := redis.NewStringCmd(ctx)
	if f.err != nil {
		cmd.SetErr(f.err)
	}
	return cmd
}

func TestHandleMessage(t *testing.T) {
	snapshot := &marketsv1.SsiPsSnapshot{Code: "ABC", LastPrice: 1.23}
	snapshot.Timestamp = timestamppb.New(time.Unix(1700, 0))
	payload, _ := proto.Marshal(snapshot)

	fr := &fakeRedis{}
	svc := &Consumer{redis: fr, keyFmt: "ssi_ps:%s", streamFmt: "ssi_ps_stream:%s"}

	msg := kafka.Message{Value: payload}
	if err := svc.handleMessage(context.Background(), msg); err != nil {
		t.Fatalf("handleMessage returned error: %v", err)
	}

	if fr.key != "ssi_ps:ABC" {
		t.Fatalf("unexpected key %s", fr.key)
	}
	if len(fr.args.Members) != 1 {
		t.Fatalf("expected single member, got %v", fr.args.Members)
	}
	if v, ok := fr.args.Members[0].Member.(string); !ok || v == "" {
		t.Fatalf("expected string payload, got %T", fr.args.Members[0].Member)
	}
	if fr.stream != "ssi_ps_stream:ABC" {
		t.Fatalf("unexpected stream %s", fr.stream)
	}
	if fr.streamArgs == nil {
		t.Fatalf("expected stream args")
	}
	vals, ok := fr.streamArgs.Values.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map values, got %T", fr.streamArgs.Values)
	}
	if vals["payload"] == nil {
		t.Fatalf("expected stream payload")
	}
}
