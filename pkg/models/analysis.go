package models

type WindowInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Day  int    `json:"day"` // 数据查询前多少天
}

type RecommendRangeResult struct {
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	CumPct float64 `json:"cumPct"`
}

type MeansAvgData struct {
	Count        int                   `json:"count"`
	Avg          float64               `json:"avg"`
	Median       float64               `json:"median"`
	Mean         float64               `json:"mean"`
	Distribution []*DistBucket         `json:"distribution"`
	Recommend    *RecommendRangeResult `json:"recommend,omitempty"`
}

type MeansData struct {
	SpreadOH *MeansAvgData `json:"spreadOH"`
	SpreadOL *MeansAvgData `json:"spreadOL"`
	SpreadHL *MeansAvgData `json:"spreadHL"`
	SpreadHC *MeansAvgData `json:"spreadHC"`
	SpreadLC *MeansAvgData `json:"spreadLC"`
	SpreadOC *MeansAvgData `json:"spreadOC"`
}

type WindowData struct {
	Info  *WindowInfo `json:"info"`
	Rows  []*DailyBar `json:"-"`
	Means *MeansData  `json:"means"`
}

type DistBucket struct {
	Lower float64 `json:"lower,omitempty"`
	Upper float64 `json:"upper,omitempty"`
	Count int     `json:"count,omitempty"`
	Pct   float64 `json:"pct,omitempty"`
}

// Input is everything Build needs.
type Input struct {
	TsCode      string      `json:"tsCode,omitempty"`
	StockName   string      `json:"stockName,omitempty"`
	Rows        []*DailyBar `json:"rows,omitempty"`
	OpenPrice   *float64    `json:"openPrice,omitempty"`
	ActualHigh  *float64    `json:"actualHigh,omitempty"`
	ActualLow   *float64    `json:"actualLow,omitempty"`
	ActualClose *float64    `json:"actualClose,omitempty"`
}

// PredictRow holds one row of the prediction reference table.
// Float64 fields use 0 to indicate "no data".
type PredictRow struct {
	Windows     map[string]float64 `json:"windows"`
	ReverseLow  float64            `json:"reverseLow"`
	ReverseHigh float64            `json:"reverseHigh"`
	Mean        float64            `json:"mean"`
	Direction   string             `json:"direction"`
}

// RefTable is the pre-computed prediction reference table (最高/最低/收盘价预测).
type RefTable struct {
	High  PredictRow `json:"high"`
	Low   PredictRow `json:"low"`
	Close PredictRow `json:"close"`
}

// AnalysisResult is the canonical output with raw computed data.
type AnalysisResult struct {
	TsCode         string             `json:"tsCode"`
	StockName      string             `json:"stockName"`
	Windows        []*WindowData      `json:"windows"`
	CompositeMeans map[string]float64 `json:"compositeMeans"`
	RefTable       *RefTable          `json:"refTable,omitempty"`
	OpenPrice      *float64           `json:"openPrice,omitempty"`
	ActualHigh     *float64           `json:"actualHigh,omitempty"`
	ActualLow      *float64           `json:"actualLow,omitempty"`
	ActualClose    *float64           `json:"actualClose,omitempty"`
}
