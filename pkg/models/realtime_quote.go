package models

import "time"

type RealtimeQuote struct {
	TsCode    string    `json:"tsCode"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`      // [3]  当前价
	PrevClose float64   `json:"prevClose"`  // [4]  昨收
	Open      float64   `json:"open"`       // [5]  今开
	Vol       float64   `json:"vol"`        // [6]  成交量（手）
	High      float64   `json:"high"`       // [33] 最高
	Low       float64   `json:"low"`        // [34] 最低
	Amount    float64   `json:"amount"`     // [37] 成交额（万元）
	Change    float64   `json:"change"`     // [31] 涨跌
	ChangePct float64   `json:"changePct"`  // [32] 涨跌%
	LimitUp   float64   `json:"limitUp"`    // [47] 涨停价
	LimitDown float64   `json:"limitDown"`  // [48] 跌停价
	QuoteTime string    `json:"quoteTime"`  // [30] 行情时间（原始字符串）
	UpdatedAt time.Time `json:"updatedAt"`
}
