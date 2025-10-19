package stocks_parser

import (
	"strings"
	"testing"
	"time"
)

func TestHoseStockParser_Parse(t *testing.T) {
	parser := &HoseStockParser{BaseParser{Columns: HoseColumns}}
	sample := "1754535567282|MAIN|S#41I1F8000|1717.9|1|1717.8|55|1717.7|10|1717.6|9|1717.5|23|1717.4|15|1717.3|27|1717.2|13|1717.1|37|1717|84|1718.1|2|1718.3|7|1718.4|5|1718.5|55|1718.6|11|1718.8|15|1718.9|31|1719|33|1719.2|61|1719.3|18|1718|1|1732.2|vnf|1710|1719.91|3264|561391850000|3490|600494840000|-1.5|-0.09|172480|29665063450000|16931|8494|56214|1839.8|1599.2|1719.5|21/08/2025||||f|N|20250821||VN30|0||||1722|||||||||||||||||||||1722|1687|2.5|0.15|1754531999939|1754547232540"

	result, err := parser.Parse(sample)
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if len(result) == 0 {
		t.Fatalf("expected parsed result, got empty map")
	}

	if board, ok := result["board"].(string); !ok || board != "MAIN" {
		t.Fatalf("expected board MAIN got %v", result["board"])
	}

	if raw, ok := result["raw_symbol"].(string); !ok || raw != "S#41I1F8000" {
		t.Fatalf("unexpected raw_symbol: %v", result["raw_symbol"])
	}

	if code, ok := result["code"].(string); !ok || code != "41I1F8000" {
		t.Fatalf("unexpected code: %v", result["code"])
	}

	ts, ok := result["timestamp"].(time.Time)
	if !ok {
		t.Fatalf("timestamp missing or wrong type: %T", result["timestamp"])
	}
	if ts.Location() != localTimezone() {
		t.Fatalf("timestamp should use ICT timezone, got %v", ts.Location())
	}

	if srv, ok := result["server_timestamp"].(time.Time); !ok || srv.IsZero() {
		t.Fatalf("expected server_timestamp to be parsed, got %v", result["server_timestamp"])
	}
}

func TestIndexStockParser_Parse(t *testing.T) {
	parser := &IndexStockParser{BaseParser{Columns: IndexColumns}}

	invalidFields := []string{"1735533056471", "I#VNFINLEAD", "2146.75"}
	invalidSample := strings.Join(invalidFields, "|")
	if _, err := parser.Parse(invalidSample); err == nil {
		t.Fatalf("expected error for malformed payload")
	}

	fields := []string{
		"1735533056471", "I#VNFINLEAD", "2146.75", "56055700", "1375219000000",
		"6", "13", "4", "0", "0", "74089477", "1658163000000", "", "",
		"1735558255447", "56055700", "2157.14", "2157.84", "2142.67", "-7.33",
		"-0.34", "", "", "", "", "",
	}
	validSample := strings.Join(fields, "|")

	res, err := parser.Parse(validSample)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id, ok := res["indexId"].(string); !ok || id != "VNFINLEAD" {
		t.Fatalf("expected indexId VNFINLEAD got %v", res["indexId"])
	}
}
