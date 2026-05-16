package models

import "time"

// StockRealtime 股票最新价格信息
type StockRealtime struct {
	TsCode         string    `json:"tsCode"`
	Name           string    `json:"name"`
	Price          float64   `json:"price"`          // [3]  当前价格
	PrevClose      float64   `json:"prevClose"`      // [4]  昨收
	Open           float64   `json:"open"`           // [5]  今开
	Vol            float64   `json:"vol"`            // [6]  成交量（手）
	OuterVol       float64   `json:"outerVol"`       // [7]  外盘
	InnerVol       float64   `json:"innerVol"`       // [8]  内盘
	High           float64   `json:"high"`           // [33] 最高
	Low            float64   `json:"low"`            // [34] 最低
	TotalVol       float64   `json:"totalVol"`       // [36] 成交量（手）
	Amount         float64   `json:"amount"`         // [37] 成交额（万元）
	TurnoverRate   float64   `json:"turnoverRate"`   // [38] 换手率
	PE             float64   `json:"pe"`             // [39] 市盈率
	High52w        float64   `json:"high52w"`        // [41] 52周最高
	Low52w         float64   `json:"low52w"`         // [42] 52周最低
	Amplitude      float64   `json:"amplitude"`      // [43] 振幅
	CircMarketCap  float64   `json:"circMarketCap"`  // [44] 流通市值
	TotalMarketCap float64   `json:"totalMarketCap"` // [45] 总市值
	PB             float64   `json:"pb"`             // [46] 市净率
	Change         float64   `json:"change"`         // [31] 涨跌
	ChangePct      float64   `json:"changePct"`      // [32] 涨跌%
	LimitUp        float64   `json:"limitUp"`        // [47] 涨停价
	LimitDown      float64   `json:"limitDown"`      // [48] 跌停价
	QuoteTime      string    `json:"quoteTime"`      // [30] 行情时间
	UpdatedAt      time.Time `json:"updatedAt"`
}

type StockRealtimeAndAnalysis struct {
	StockRealtime       *StockRealtime
	StockAnalysisResult *StockAnalysisResult
}
