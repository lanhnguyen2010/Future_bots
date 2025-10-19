package producer

import (
	"testing"
	"time"
)

func TestMapToSsiPsSnapshot(t *testing.T) {
	now := time.Unix(1700, 0)
	payload := map[string]interface{}{
		"board":              "MAIN",
		"raw_symbol":         "S#ABC",
		"code":               "ABC",
		"timestamp":          now,
		"server_timestamp":   now.Add(time.Second),
		"bestBid1":           1.1,
		"bestBid1Volume":     int64(10),
		"bestOffer1":         2.2,
		"bestOffer1Volume":   int64(20),
		"last_":              1.5,
		"highestPrice":       2.0,
		"type":               "vnf",
		"lowestPrice":        1.0,
		"foreignerBuyVolume": int64(5),
		"change":             -0.5,
		"totalMatchVolume":   int64(100),
		"cellingPrice":       3.0,
		"floorPrice":         0.5,
		"refPrice":           1.2,
		"group_":             "VN30",
		"floating":           int64(999),
	}

	snapshot, err := MapToSsiPsSnapshot(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snapshot.GetCode() != "ABC" {
		t.Fatalf("unexpected code %s", snapshot.GetCode())
	}
	if got := snapshot.GetBestBid_1(); got != 1.1 {
		t.Fatalf("unexpected best bid %v", got)
	}
	if snapshot.GetBestOffer_1Volume() != 20 {
		t.Fatalf("unexpected offer vol %d", snapshot.GetBestOffer_1Volume())
	}
	if !snapshot.GetTimestamp().AsTime().Equal(now) {
		t.Fatalf("unexpected timestamp %v", snapshot.GetTimestamp().AsTime())
	}
	if !snapshot.GetServerTimestamp().AsTime().Equal(now.Add(time.Second)) {
		t.Fatalf("unexpected server timestamp %v", snapshot.GetServerTimestamp().AsTime())
	}
	if snapshot.GetFloatingShares() != 999 {
		t.Fatalf("unexpected floating shares %d", snapshot.GetFloatingShares())
	}
}

func TestMapToSsiPsSnapshotValidation(t *testing.T) {
	_, err := MapToSsiPsSnapshot(map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected error when fields missing")
	}
}
