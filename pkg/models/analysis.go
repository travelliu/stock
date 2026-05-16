// Package models defines the data structures for A-share intraday spread analysis.
//
// Data flow:
//
//	DailyBar rows → WindowData (sliced by time window) → MeansData (spread statistics)
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
	Info    *WindowInfo    `json:"info"`
	Rows    []*DailyBar    `json:"-"`
	Means   *MeansData     `json:"means"`
	Predict *WindowPredict `json:"predict,omitempty"`
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
	TsCode      string      `json:"tsCode,omitempty"`
	StockName   string      `json:"stockName,omitempty"`
	Rows        []*DailyBar `json:"rows,omitempty"`
	OpenPrice   *float64    `json:"openPrice,omitempty"`
	ActualHigh  *float64    `json:"actualHigh,omitempty"`
	ActualLow   *float64    `json:"actualLow,omitempty"`
	ActualClose *float64    `json:"actualClose,omitempty"`
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

// AnalysisResult is the canonical output of Build.
// Windows are ordered oldest-to-newest (All → last_90 → last_30 → last_15).
// CompositeMeans is the spread-level cross-window average used for the summary model table.
type AnalysisResult struct {
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
