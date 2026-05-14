package models

// Input is everything Build needs.
type Input struct {
	TsCode      string     `json:"tsCode,omitempty"`
	StockName   string     `json:"stockName,omitempty"`
	Rows        []DailyBar `json:"rows,omitempty"` // raw daily history
	OpenPrice   *float64   `json:"openPrice,omitempty"`
	ActualHigh  *float64   `json:"actualHigh,omitempty"`
	ActualLow   *float64   `json:"actualLow,omitempty"`
	ActualClose *float64   `json:"actualClose,omitempty"`
}

// AnalysisResult is the canonical output. Field naming matches the design spec §3.4.
type AnalysisResult struct {
	TsCode         string             `json:"tsCode"`
	StockName      string             `json:"stockName"`
	YesterdayClose *float64           `json:"yesterdayClose,omitempty"`
	Windows        []string           `json:"windows"` // ["历史","近3月","近1月","近2周"]
	OpenPrice      *float64           `json:"openPrice,omitempty"`
	ActualHigh     *float64           `json:"actualHigh,omitempty"`
	ActualLow      *float64           `json:"actualLow,omitempty"`
	ActualClose    *float64           `json:"actualClose,omitempty"`
	WindowMeans    MeansResult        `json:"windowMeans"` // window -> spread_key -> *float64
	CompositeMeans map[string]float64 `json:"compositeMeans"`
	ModelTable     ModelTable         `json:"modelTable"`
	ReferenceTable ReferenceTable     `json:"referenceTable"`
}

// ModelTable is the 4-window × 6-spread table (plus a composite row).
type ModelTable struct {
	Headers []string   `json:"headers"` // ["时段","开盘与最高价",...]
	Rows    [][]string `json:"rows"`    // formatted cells, "%.2f" or "-"
}

// ReferenceTable is the 3-prediction-row table (high/low/close).
type ReferenceTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

type MeansResult map[string]map[string]*float64

type Window struct {
	Name string
	Rows []DailyBar
}
