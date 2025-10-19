package stocks_parser

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type ColumnDefinition struct {
	Name    string
	Parse   func(string) interface{}
	SQLType string
}

func ParseToInt(s string) interface{} {
	if s == "" {
		return int64(0)
	}
	s = strings.TrimSpace(s)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	return int64(0)
}

func RemoveFirstTwoChars(s string) interface{} {
	if len(s) <= 2 {
		return ""
	}
	return s[2:]
}
func ParseToFloat(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return float64(0)
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return float64(0)
}

func GetNow(s string) interface{} {
	return time.Now().Unix()
}

func ParseString(s string) interface{} {
	return s
}

func ParseTimeStamp(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	millis, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}
	}
	seconds := millis / 1000
	nanos := (millis % 1000) * int64(time.Millisecond)
	return time.Unix(seconds, nanos).In(localTimezone())
}

var HoseColumns = map[int]ColumnDefinition{
	0:  {"code", RemoveFirstTwoChars, "varchar(10)"},
	1:  {"bestBid1", ParseToFloat, "decimal (20,2)"},
	2:  {"bestBid1Volume", ParseToInt, "bigint"},
	3:  {"bestBid2", ParseToFloat, "decimal (20,2)"},
	4:  {"bestBid2Volume", ParseToInt, "bigint"},
	5:  {"bestBid3", ParseToFloat, "decimal (20,2)"},
	6:  {"bestBid3Volume", ParseToInt, "bigint"},
	7:  {"bestBid4", ParseToFloat, "decimal (20,2)"},
	8:  {"bestBid4Volume", ParseToInt, "bigint"},
	9:  {"bestBid5", ParseToFloat, "decimal (20,2)"},
	10: {"bestBid5Volume", ParseToInt, "bigint"},
	11: {"bestBid6", ParseToFloat, "decimal (20,2)"},
	12: {"bestBid6Volume", ParseToInt, "bigint"},
	13: {"bestBid7", ParseToFloat, "decimal (20,2)"},
	14: {"bestBid7Volume", ParseToInt, "bigint"},
	15: {"bestBid8", ParseToFloat, "decimal (20,2)"},
	16: {"bestBid8Volume", ParseToInt, "bigint"},
	17: {"bestBid9", ParseToFloat, "decimal (20,2)"},
	18: {"bestBid9Volume", ParseToInt, "bigint"},
	19: {"bestBid10", ParseToFloat, "decimal (20,2)"},
	20: {"bestBid10Volume", ParseToInt, "bigint"},
	21: {"bestOffer1", ParseToFloat, "decimal (20,2)"},
	22: {"bestOffer1Volume", ParseToInt, "bigint"},
	23: {"bestOffer2", ParseToFloat, "decimal (20,2)"},
	24: {"bestOffer2Volume", ParseToInt, "bigint"},
	25: {"bestOffer3", ParseToFloat, "decimal (20,2)"},
	26: {"bestOffer3Volume", ParseToInt, "bigint"},
	27: {"bestOffer4", ParseToFloat, "decimal (20,2)"},
	28: {"bestOffer4Volume", ParseToInt, "bigint"},
	29: {"bestOffer5", ParseToFloat, "decimal (20,2)"},
	30: {"bestOffer5Volume", ParseToInt, "bigint"},
	31: {"bestOffer6", ParseToFloat, "decimal (20,2)"},
	32: {"bestOffer6Volume", ParseToInt, "bigint"},
	33: {"bestOffer7", ParseToFloat, "decimal (20,2)"},
	34: {"bestOffer7Volume", ParseToInt, "bigint"},
	35: {"bestOffer8", ParseToFloat, "decimal (20,2)"},
	36: {"bestOffer8Volume", ParseToInt, "bigint"},
	37: {"bestOffer9", ParseToFloat, "decimal (20,2)"},
	38: {"bestOffer9Volume", ParseToInt, "bigint"},
	39: {"bestOffer10", ParseToFloat, "decimal (20,2)"},
	40: {"bestOffer10Volume", ParseToInt, "bigint"},
	41: {"last_", ParseToFloat, "decimal (20,2)"},
	43: {"highestPrice", ParseToFloat, "decimal (20,2)"},
	44: {"type", ParseString, "varchar(10)"},
	45: {"lowestPrice", ParseToFloat, "decimal (20,2)"},
	47: {"foreignerBuyVolume", ParseToInt, "bigint"},
	48: {"foreignerBuyValue", ParseToInt, "bigint"},
	49: {"foreignerSellVolume", ParseToInt, "bigint"},
	50: {"foreignerSellValue", ParseToInt, "bigint"},
	51: {"change", ParseToFloat, "decimal (20,2)"},
	53: {"totalMatchVolume", ParseToInt, "bigint"},
	54: {"totalMatchValue", ParseToInt, "bigint"},
	55: {"totalBidVolume", ParseToInt, "bigint"},
	56: {"totalAskVolume", ParseToInt, "bigint"},
	57: {"openInterest", ParseToInt, "bigint"},
	58: {"cellingPrice", ParseToFloat, "decimal (20,2)"},
	59: {"floorPrice", ParseToFloat, "decimal (20,2)"},
	60: {"refPrice", ParseToFloat, "decimal (20,2)"},
	61: {"expireDate", ParseString, "varchar(10)"},
	62: {"session", ParseString, "varchar(10)"},
	64: {"foreignRoom", ParseString, "varchar(10)"}, // Not too sure
	69: {"group_", ParseString, "varchar(10)"},
	//70: {"70", ParseToInt, "bigint"},
	71: {"floating", ParseToInt, "bigint"},
	-2: {"server_timestamp", ParseTimeStamp, "timestamp"},
	-1: {"timestamp", GetNow, "timestamp"},
}

var ThoaThuanCloumns = []ColumnDefinition{
	{"_id", ParseString, "varchar(50) PRIMARY KEY"},
	{"stockSymbol", ParseString, "varchar(20)"},
	{"exchange", ParseString, "varchar(10)"},
	{"refPrice", ParseToInt, "bigint"},
	{"ceiling", ParseToInt, "bigint"},
	{"floor", ParseToInt, "bigint"},
	{"vol", ParseToInt, "bigint"},
	{"val", ParseToInt, "bigint"},
	{"price", ParseToInt, "bigint"},
	{"board", ParseString, "varchar(20)"},
	{"ptTotalTradedQty", ParseToInt, "bigint"},
	{"ptTotalTradedValue", ParseToInt, "bigint"},
	{"createdAt", ParseString, "TIME"},
	{"__v", ParseToInt, "int"},
	{"timestamp", GetNow, "timestamp"},
}

// I#VNINDEX|1284.41|993342917|25420890000000|146|301|63|9|3|993342917|25420890000000|||1718118124789|993342917|1295.03|1296.41|1279.47|-6.26|-0.49|||||

var IndexColumns = map[int]ColumnDefinition{
	0:  {"timestamp", ParseTimeStamp, "timestamp"},
	1:  {"indexId", RemoveFirstTwoChars, "varchar(20)"},
	2:  {"last", ParseToFloat, "decimal (20,2)"},
	3:  {"vol_no_pt", ParseToInt, "bigint"},
	4:  {"val_no_pt", ParseToInt, "bigint"},
	5:  {"increase", ParseToInt, "bigint"},
	6:  {"decrease", ParseToInt, "bigint"},
	7:  {"unchanged", ParseToInt, "bigint"},
	10: {"vol_pt", ParseToInt, "bigint"},
	11: {"val_pt", ParseToInt, "bigint"},
	12: {"matched_timestamp", ParseToInt, "bigint"},
	13: {"matched_vol", ParseToInt, "bigint"},
	14: {"server_time", ParseTimeStamp, "timestamp"},
	15: {"prev_vol_no_pt", ParseToInt, "bigint"},
	16: {"open", ParseToFloat, "decimal(20,2)"},
	17: {"high", ParseToFloat, "decimal(20,2)"},
	18: {"low", ParseToFloat, "decimal(20,2)"},
	19: {"change", ParseToFloat, "decimal(20,2)"},
	20: {"pct_change", ParseToFloat, "decimal(20,2)"},
	25: {"projected_settlement", ParseToFloat, "decimal(20,2)"},
}

type IParser interface {
	LoadParserCongfig()
	Parse(data string) (interface{}, error)
}

type BaseParser struct {
	Columns map[int]ColumnDefinition
}

func (p *BaseParser) LoadParserCongfig() {
}

type HoseStockParser struct {
	BaseParser
}

type IndexStockParser struct {
	BaseParser
}

func (p *HoseStockParser) Parse(data string) (map[string]interface{}, error) {
	payload := strings.TrimSpace(data)
	if payload == "" {
		return nil, errors.New("empty payload")
	}

	parts := strings.SplitN(payload, "|", 2)
	if len(parts) < 2 {
		return nil, errors.New("invalid payload, missing board/data section")
	}

	timestamp := ParseTimeStamp(parts[0]).(time.Time)

	remaining := strings.SplitN(parts[1], "|", 2)
	if len(remaining) < 2 {
		return nil, errors.New("invalid payload, missing symbol section")
	}

	splits := strings.Split(remaining[1], "|")
	parsedData := make(map[string]interface{})

	board := strings.TrimSpace(remaining[0])
	if board != "" {
		parsedData["board"] = board
	}

	if len(splits) > 0 {
		rawSymbol := strings.TrimSpace(splits[0])
		if rawSymbol != "" {
			parsedData["raw_symbol"] = rawSymbol
		}
	}

	for i, v := range splits {
		if column, ok := p.Columns[i]; ok {
			parsedData[column.Name] = column.Parse(v)
		}
	}

	parsedData["timestamp"] = timestamp

	if len(splits) >= 2 {
		candidate := strings.TrimSpace(splits[len(splits)-2])
		if candidate != "" {
			parsed := ParseTimeStamp(candidate)
			if ts, ok := parsed.(time.Time); ok && !ts.IsZero() {
				parsedData["server_timestamp"] = ts
			}
		}
	}
	return parsedData, nil
}

func (p *IndexStockParser) Parse(data string) (map[string]interface{}, error) {
	splits := strings.Split(strings.TrimSpace(data), "|")
	if len(splits) != 26 {
		return nil, errors.New("unexpected field count for index payload")
	}

	parsedData := make(map[string]interface{})
	for i, v := range splits {
		if column, ok := p.Columns[i]; ok {
			parsedData[column.Name] = column.Parse(v)
		}
	}
	return parsedData, nil
}

var ict = time.FixedZone("ICT", 7*60*60)

func localTimezone() *time.Location {
	return ict
}
