package producer

import (
	"fmt"
	"time"

	marketsv1 "github.com/future-bots/proto/markets/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MapToSsiPsSnapshot converts the Hose parser output into a strongly-typed protobuf message.
func MapToSsiPsSnapshot(values map[string]interface{}) (*marketsv1.SsiPsSnapshot, error) {
	if values == nil {
		return nil, fmt.Errorf("nil payload")
	}

	snapshot := &marketsv1.SsiPsSnapshot{}

	snapshot.Board = stringValue(values, "board")
	snapshot.RawSymbol = stringValue(values, "raw_symbol")
	snapshot.Code = stringValue(values, "code")
	if snapshot.Code == "" {
		return nil, fmt.Errorf("code missing")
	}

	if ts := timestampValue(values, "timestamp"); ts != nil {
		snapshot.Timestamp = ts
	} else {
		return nil, fmt.Errorf("timestamp missing")
	}
	if ts := timestampValue(values, "server_timestamp"); ts != nil {
		snapshot.ServerTimestamp = ts
	}

	snapshot.BestBid_1 = floatValue(values, "bestBid1")
	snapshot.BestBid_1Volume = intValue(values, "bestBid1Volume")
	snapshot.BestBid_2 = floatValue(values, "bestBid2")
	snapshot.BestBid_2Volume = intValue(values, "bestBid2Volume")
	snapshot.BestBid_3 = floatValue(values, "bestBid3")
	snapshot.BestBid_3Volume = intValue(values, "bestBid3Volume")
	snapshot.BestBid_4 = floatValue(values, "bestBid4")
	snapshot.BestBid_4Volume = intValue(values, "bestBid4Volume")
	snapshot.BestBid_5 = floatValue(values, "bestBid5")
	snapshot.BestBid_5Volume = intValue(values, "bestBid5Volume")
	snapshot.BestBid_6 = floatValue(values, "bestBid6")
	snapshot.BestBid_6Volume = intValue(values, "bestBid6Volume")
	snapshot.BestBid_7 = floatValue(values, "bestBid7")
	snapshot.BestBid_7Volume = intValue(values, "bestBid7Volume")
	snapshot.BestBid_8 = floatValue(values, "bestBid8")
	snapshot.BestBid_8Volume = intValue(values, "bestBid8Volume")
	snapshot.BestBid_9 = floatValue(values, "bestBid9")
	snapshot.BestBid_9Volume = intValue(values, "bestBid9Volume")
	snapshot.BestBid_10 = floatValue(values, "bestBid10")
	snapshot.BestBid_10Volume = intValue(values, "bestBid10Volume")

	snapshot.BestOffer_1 = floatValue(values, "bestOffer1")
	snapshot.BestOffer_1Volume = intValue(values, "bestOffer1Volume")
	snapshot.BestOffer_2 = floatValue(values, "bestOffer2")
	snapshot.BestOffer_2Volume = intValue(values, "bestOffer2Volume")
	snapshot.BestOffer_3 = floatValue(values, "bestOffer3")
	snapshot.BestOffer_3Volume = intValue(values, "bestOffer3Volume")
	snapshot.BestOffer_4 = floatValue(values, "bestOffer4")
	snapshot.BestOffer_4Volume = intValue(values, "bestOffer4Volume")
	snapshot.BestOffer_5 = floatValue(values, "bestOffer5")
	snapshot.BestOffer_5Volume = intValue(values, "bestOffer5Volume")
	snapshot.BestOffer_6 = floatValue(values, "bestOffer6")
	snapshot.BestOffer_6Volume = intValue(values, "bestOffer6Volume")
	snapshot.BestOffer_7 = floatValue(values, "bestOffer7")
	snapshot.BestOffer_7Volume = intValue(values, "bestOffer7Volume")
	snapshot.BestOffer_8 = floatValue(values, "bestOffer8")
	snapshot.BestOffer_8Volume = intValue(values, "bestOffer8Volume")
	snapshot.BestOffer_9 = floatValue(values, "bestOffer9")
	snapshot.BestOffer_9Volume = intValue(values, "bestOffer9Volume")
	snapshot.BestOffer_10 = floatValue(values, "bestOffer10")
	snapshot.BestOffer_10Volume = intValue(values, "bestOffer10Volume")

	snapshot.LastPrice = floatValue(values, "last_")
	snapshot.HighestPrice = floatValue(values, "highestPrice")
	snapshot.InstrumentType = stringValue(values, "type")
	snapshot.LowestPrice = floatValue(values, "lowestPrice")

	snapshot.ForeignerBuyVolume = intValue(values, "foreignerBuyVolume")
	snapshot.ForeignerBuyValue = intValue(values, "foreignerBuyValue")
	snapshot.ForeignerSellVolume = intValue(values, "foreignerSellVolume")
	snapshot.ForeignerSellValue = intValue(values, "foreignerSellValue")

	snapshot.Change = floatValue(values, "change")
	snapshot.TotalMatchVolume = intValue(values, "totalMatchVolume")
	snapshot.TotalMatchValue = intValue(values, "totalMatchValue")
	snapshot.TotalBidVolume = intValue(values, "totalBidVolume")
	snapshot.TotalAskVolume = intValue(values, "totalAskVolume")
	snapshot.OpenInterest = intValue(values, "openInterest")

	snapshot.CeilingPrice = floatValue(values, "cellingPrice")
	snapshot.FloorPrice = floatValue(values, "floorPrice")
	snapshot.ReferencePrice = floatValue(values, "refPrice")

	snapshot.ExpireDate = stringValue(values, "expireDate")
	snapshot.Session = stringValue(values, "session")
	snapshot.ForeignRoom = stringValue(values, "foreignRoom")
	snapshot.Group = stringValue(values, "group_")
	snapshot.FloatingShares = intValue(values, "floating")

	return snapshot, nil
}

func stringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch value := v.(type) {
		case string:
			return value
		}
	}
	return ""
}

func intValue(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch value := v.(type) {
		case int64:
			return value
		case int:
			return int64(value)
		case float64:
			return int64(value)
		}
	}
	return 0
}

func floatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch value := v.(type) {
		case float64:
			return value
		case float32:
			return float64(value)
		case int64:
			return float64(value)
		case int:
			return float64(value)
		}
	}
	return 0
}

func timestampValue(m map[string]interface{}, key string) *timestamppb.Timestamp {
	if v, ok := m[key]; ok {
		switch value := v.(type) {
		case time.Time:
			if value.IsZero() {
				return nil
			}
			ts := timestamppb.New(value)
			if err := ts.CheckValid(); err != nil {
				return nil
			}
			return ts
		}
	}
	return nil
}
