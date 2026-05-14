// Package analysis ports the Python price-spread analysis pipeline:
// window means, composite means, the spread-model table and the reference
// (predicted price) table. The output of Build is the canonical analysis
// payload returned by the HTTP API and rendered by the CLI.
package analysis

import (
	"stock/pkg/shared/spread"
	"stock/pkg/shared/window"
)

// Input is everything Build needs.
type Input struct {
	TsCode       string
	StockName    string
	Rows         []spread.Bar // raw daily history
	OpenPrice    *float64
	ActualHigh   *float64
	ActualLow    *float64
	ActualClose  *float64
}

// AnalysisResult is the canonical output. Field naming matches the design spec §3.4.
type AnalysisResult struct {
	TsCode          string             `json:"ts_code"`
	StockName       string             `json:"stock_name"`
	YesterdayClose  *float64           `json:"yesterday_close,omitempty"`
	Windows         []string           `json:"windows"` // ["历史","近3月","近1月","近2周"]
	OpenPrice       *float64           `json:"open_price,omitempty"`
	ActualHigh      *float64           `json:"actual_high,omitempty"`
	ActualLow       *float64           `json:"actual_low,omitempty"`
	ActualClose     *float64           `json:"actual_close,omitempty"`
	WindowMeans     window.MeansResult `json:"window_means"`    // window -> spread_key -> *float64
	CompositeMeans  map[string]float64 `json:"composite_means"`
	ModelTable      ModelTable         `json:"model_table"`
	ReferenceTable  ReferenceTable     `json:"reference_table"`
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
