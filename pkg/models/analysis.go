// Package models defines the data structures for A-share intraday spread analysis.
//
// Data flow:
//
//	StockDailyBar rows → WindowData (sliced by time window) → MeansData (spread statistics)
//	→ WindowPredict (price predictions per window) → RefTable (cross-window summary)
//
// Six price spreads are tracked for each bar:
//
//	OH = High − Open   (上涨空间，高抛参考)
//	OL = Open − Low    (下跌空间，低吸参考)
//	HL = High − Low    (全日振幅，做T空间)
//	OC = Open − Close  (日向：正=收阴，负=收阳)
//	HC = High − Close  (高收差)
//	LC = Close − Low   (低收差)
package models

import (
	"encoding/json"
	"time"
)

// WindowInfo describes a time window slice (e.g. last 30 trading days).
type WindowInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Day  int    `json:"day"`
}

// RecommendRangeResult is the output of RecommendRange: the narrowest price band
// that covers at least the requested percentage of historical observations.
type RecommendRangeResult struct {
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	CumPct float64 `json:"cumPct"` // actual coverage percentage of the returned band
}

// MeansAvgData holds statistical summary for one spread type within one time window.
// Multiple estimators are provided so callers can choose robustness vs. recency:
//   - Avg/Median/Mean: classic descriptive stats; Mean = (Avg+Median)/2 dampens outliers
//   - EWMA: exponentially weighted (λ=0.9, newest-first), emphasises recent days
//   - AvgRatio/EWMARatio: spread expressed as a fraction of the opening price,
//     making predictions scale-invariant when the stock price level shifts
type MeansAvgData struct {
	Count        int                   `json:"count"`
	Avg          float64               `json:"avg"`       // arithmetic mean
	Median       float64               `json:"median"`    // 50th percentile, robust to outlier days
	Mean         float64               `json:"mean"`      // (Avg + Median) / 2
	EWMA         float64               `json:"ewma"`      // λ=0.9 exponential weighted mean
	StdDev       float64               `json:"stdDev"`    // population standard deviation
	AvgRatio     float64               `json:"avgRatio"`  // Avg / open price
	EWMARatio    float64               `json:"ewmaRatio"` // EWMA / open price
	Distribution []*DistBucket         `json:"distribution"`
	Recommend    *RecommendRangeResult `json:"recommend,omitempty"`
}

// MeansData groups the six spread statistics for one time window.
type MeansData struct {
	SpreadOH *MeansAvgData `json:"spreadOH"` // high − open
	SpreadOL *MeansAvgData `json:"spreadOL"` // open − low
	SpreadHL *MeansAvgData `json:"spreadHL"` // high − low  (full-day amplitude)
	SpreadHC *MeansAvgData `json:"spreadHC"` // high − close
	SpreadLC *MeansAvgData `json:"spreadLC"` // close − low
	SpreadOC *MeansAvgData `json:"spreadOC"` // open − close (positive = bearish day)
}

// PredictBreakdown holds every prediction method for one price target (High/Low/Close)
// within one time window. All fields are absolute prices (not spreads).
//
// Forward methods (require today's open price):
//
//	ByMean   = open ± spread.Mean
//	ByMedian = open ± spread.Median
//	ByEWMA   = open ± spread.EWMA
//	ByRatio  = open ± open × spread.EWMARatio   (scale-adaptive)
//
// Reverse methods (require the actual intraday extreme, available after the move):
//
//	ReverseLow  = actualLow  + spread_hl.Mean  (for High);  actualLow  + spread_lc.Mean  (for Close)
//	ReverseHigh = actualHigh − spread_hl.Mean  (for Low);   actualHigh − spread_hc.Mean  (for Close)
//
// Mean is the arithmetic average of all non-zero method values above.
type PredictBreakdown struct {
	ByMean      float64 `json:"byMean"`
	ByMedian    float64 `json:"byMedian"`
	ByEWMA      float64 `json:"byEwma"`
	ByRatio     float64 `json:"byRatio"`
	ReverseLow  float64 `json:"reverseLow"`
	ReverseHigh float64 `json:"reverseHigh"`
	Mean        float64 `json:"mean"` // avg of all non-zero methods above
}

// WindowPredict holds all price predictions for one time window,
// populated by BuildRefTable after MeansData is computed.
type WindowPredict struct {
	High  PredictBreakdown `json:"high"`
	Low   PredictBreakdown `json:"low"`
	Close PredictBreakdown `json:"close"`
}

// WindowData is one time-slice of history together with its computed statistics.
// Rows is excluded from JSON because it can be very large; consumers use Means/Predict instead.
type WindowData struct {
	Info    *WindowInfo      `json:"info"`
	Rows    []*StockDailyBar `json:"-"`
	Means   *MeansData       `json:"means"`
	Predict *WindowPredict   `json:"predict,omitempty"`
}

// DistBucket is one histogram bin in a spread value distribution.
type DistBucket struct {
	Lower float64 `json:"lower,omitempty"`
	Upper float64 `json:"upper,omitempty"`
	Count int     `json:"count,omitempty"`
	Pct   float64 `json:"pct,omitempty"` // percentage of total samples in this bin
}

// Input is everything Build needs to run the full analysis pipeline.
// OpenPrice is known at session open; ActualHigh/Low/Close are only available
// post-session and are optional — reverse-lookup predictions are skipped when nil.
type Input struct {
	TsCode      string           `json:"tsCode,omitempty"`
	StockName   string           `json:"stockName,omitempty"`
	Rows        []*StockDailyBar `json:"rows,omitempty"`
	OpenPrice   *float64         `json:"openPrice,omitempty"`
	ActualHigh  *float64         `json:"actualHigh,omitempty"`
	ActualLow   *float64         `json:"actualLow,omitempty"`
	ActualClose *float64         `json:"actualClose,omitempty"`
}

// PredictRow holds the cross-window average of per-window Mean values for one price target.
// Detailed per-window breakdowns live in WindowData.Predict.
type PredictRow struct {
	Mean float64 `json:"mean"` // arithmetic average of WindowPredict.{High|Low|Close}.Mean across all windows
}

// RefTable aggregates predictions across all time windows into a single summary row per target.
// It is the top-level reference for a quick read; drill into WindowData.Predict for per-window detail.
type RefTable struct {
	High  PredictRow `json:"high"`
	Low   PredictRow `json:"low"`
	Close PredictRow `json:"close"`
}

// StockAnalysisResult is the canonical output of Build.
// Windows are ordered oldest-to-newest (All → last_90 → last_30 → last_15).
// CompositeMeans is the spread-level cross-window average used for the summary model table.
type StockAnalysisResult struct {
	TsCode         string             `json:"tsCode"`
	StockName      string             `json:"stockName"`
	Windows        []*WindowData      `json:"windows"`
	CompositeMeans map[string]float64 `json:"compositeMeans"` // spread key → avg of MeansAvgData.Mean across windows
	RefTable       *RefTable          `json:"refTable,omitempty"`
	OpenPrice      *float64           `json:"openPrice,omitempty"`
	ActualHigh     *float64           `json:"actualHigh,omitempty"`
	ActualLow      *float64           `json:"actualLow,omitempty"`
	ActualClose    *float64           `json:"actualClose,omitempty"`
}

type StockAnalysisPrediction struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	TsCode         string          `gorm:"uniqueIndex:idx_pred_code_date;size:16;not null" json:"tsCode"`
	TradeDate      string          `gorm:"uniqueIndex:idx_pred_code_date;size:8;not null" json:"tradeDate"`
	SampleCounts   json.RawMessage `gorm:"type:json" json:"sampleCounts"`
	WindowMeans    json.RawMessage `gorm:"type:json" json:"windowMeans"`
	CompositeMeans json.RawMessage `gorm:"type:json" json:"compositeMeans"`
	OpenPrice      float64         `json:"openPrice"`
	PredictHigh    float64         `json:"predictHigh"`
	PredictLow     float64         `json:"predictLow"`
	PredictClose   float64         `json:"predictClose"`
	ActualHigh     float64         `json:"actualHigh"`
	ActualLow      float64         `json:"actualLow"`
	ActualClose    float64         `json:"actualClose"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

type AnalysisInput struct {
	UserID      uint
	TsCode      string
	OpenPrice   *float64
	ActualHigh  *float64
	ActualLow   *float64
	ActualClose *float64
}

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
	StockRealtime       *StockRealtime       `json:"stockRealtime"`
	StockAnalysisResult *StockAnalysisResult `json:"stockAnalysisResult"`
}
