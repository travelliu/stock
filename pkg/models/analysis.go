package models

type WindowInfo struct {
	Id   string
	Name string
	Day  int
}

type MeansAvgData struct {
	Avg          float64       `json:"avg"`          // 平均值
	Median       float64       `json:"median"`       // 中位数
	Mean         float64       `json:"mean"`         // 中位数和平均数的平均
	Distribution []*DistBucket `json:"distribution"` // 区间分布
}

type MeansData struct {
	SpreadOH *MeansAvgData `json:"spreadOH"` // 开盘与最高价
	SpreadOL *MeansAvgData `json:"spreadOL"` // 开盘与最低价
	SpreadHL *MeansAvgData `json:"spreadHL"` // 最高与最低价
	SpreadHC *MeansAvgData `json:"spreadHC"` // 最高与收盘价
	SpreadLC *MeansAvgData `json:"spreadLC"` // 最低与收盘价
	SpreadOC *MeansAvgData `json:"spreadOC"` // 开盘与收盘价
}

type WindowData struct {
	Info  *WindowInfo `json:"info"`
	Rows  []*DailyBar `json:"-"`
	Means *MeansData  `json:"means"`
}

type DistBucket struct {
	Lower float64 `json:"lower,omitempty"` // 区间下界（含）
	Upper float64 `json:"upper,omitempty"` // 区间上界（不含，最后一个含）
	Count int     `json:"count,omitempty"` // 落入数量
	Pct   float64 `json:"pct,omitempty"`   // 占比 0~100
}

// Input is everything Build needs.
type Input struct {
	TsCode      string      `json:"tsCode,omitempty"`
	StockName   string      `json:"stockName,omitempty"`
	Rows        []*DailyBar `json:"rows,omitempty"` // raw daily history
	OpenPrice   *float64    `json:"openPrice,omitempty"`
	ActualHigh  *float64    `json:"actualHigh,omitempty"`
	ActualLow   *float64    `json:"actualLow,omitempty"`
	ActualClose *float64    `json:"actualClose,omitempty"`
}

// AnalysisResult is the canonical output. Field naming matches the design spec §3.4.
type AnalysisResult struct {
	TsCode         string             `json:"tsCode"`
	StockName      string             `json:"stockName"`
	Windows        []*WindowInfo      `json:"windows"` // ["历史","近3月","近1月","近2周"]
	OpenPrice      *float64           `json:"openPrice,omitempty"`
	ActualHigh     *float64           `json:"actualHigh,omitempty"`
	ActualLow      *float64           `json:"actualLow,omitempty"`
	ActualClose    *float64           `json:"actualClose,omitempty"`
	CompositeMeans map[string]float64 `json:"compositeMeans"`
	ModelTable     ModelTable         `json:"modelTable"` // 价差
	ReferenceTable ReferenceTable     `json:"referenceTable"`
}

// ModelTable is the 4-window × 6-spread table (plus a composite row).
// 价差模型
type ModelTable struct {
	Headers []string   `json:"headers"` // ["时段","开盘与最高价",...]
	Rows    [][]string `json:"rows"`    // formatted cells, "%.2f" or "-"
}

// ReferenceTable is the 3-prediction-row table (high/low/close).
type ReferenceTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}
