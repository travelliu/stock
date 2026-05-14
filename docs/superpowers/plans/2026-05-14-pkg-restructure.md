# pkg/ Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Flatten `pkg/shared/`, dissolve `pkg/stockd/models/` into `pkg/models/`, rename `stockctl` → `cli`, unify `DailyBar` + `Spreads`, and split `window.go` into `pkg/analysis` (domain logic) and `pkg/utils` (generic math) — all within one atomic commit.

**Architecture:** Mechanical file moves, import rewrites, and type unification. `DailyBar` gains an embedded `models.Spreads` with `gorm:"embedded;embeddedPrefix:spread_"` to keep the existing DB schema. `window.go` is split: analysis-specific constants/functions (`Names`, `Days`, `Make`, `Means`, `Composite`) move to `pkg/analysis`; generic math (`Distribution`, `RecommendedRange`) move to `pkg/utils`.

**Tech Stack:** Go 1.25, GORM v2, testify, SQLite (in-memory test DB)

**Reference spec:** `docs/superpowers/specs/2026-05-14-pkg-restructure-design.md`

---

### Task 1: Create `pkg/models/` and `pkg/utils/`

**Files:**
- Create: `pkg/models/spreads.go`
- Create: `pkg/models/daily_bar.go`
- Create: `pkg/utils/stockcode.go`
- Create: `pkg/utils/stockcode_test.go`
- Create: `pkg/utils/spreads.go`
- Create: `pkg/utils/spreads_test.go`
- Create: `pkg/utils/distribution.go`
- Create: `pkg/utils/distribution_test.go`
- Create: `pkg/utils/range.go`
- Create: `pkg/utils/range_test.go`
- Create: `pkg/utils/math.go`
- Move: `pkg/stockd/models/*.go` → `pkg/models/*.go`

- [ ] **Step 1: Move `pkg/stockd/models/` to `pkg/models/` (git mv, then adjust `daily_bar.go`)**

Run:
```bash
git mv pkg/stockd/models/daily_bar.go      pkg/models/daily_bar.go
git mv pkg/stockd/models/intraday_draft.go pkg/models/intraday_draft.go
git mv pkg/stockd/models/job_run.go        pkg/models/job_run.go
git mv pkg/stockd/models/portfolio.go      pkg/models/portfolio.go
git mv pkg/stockd/models/stock.go          pkg/models/stock.go
git mv pkg/stockd/models/user.go           pkg/models/user.go
git mv pkg/stockd/models/api_token.go      pkg/models/api_token.go
git mv pkg/stockd/models/models_test.go    pkg/models/models_test.go
```

Overwrite `pkg/models/daily_bar.go`:

```go
package models

type DailyBar struct {
	TsCode    string  `gorm:"primaryKey;size:16"`
	TradeDate string  `gorm:"primaryKey;size:8"`
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Vol       float64
	Amount    float64
	Spreads   Spreads `gorm:"embedded;embeddedPrefix:spread_"`
}
```

- [ ] **Step 2: Create `pkg/models/spreads.go`**

```go
package models

type Spreads struct {
	OH float64
	OL float64
	HL float64
	OC float64
	HC float64
	LC float64
}
```

- [ ] **Step 3: Move `pkg/shared/stockcode/` to `pkg/utils/`**

```bash
git mv pkg/shared/stockcode/stockcode.go      pkg/utils/stockcode.go
git mv pkg/shared/stockcode/stockcode_test.go pkg/utils/stockcode_test.go
```

Update `pkg/utils/stockcode.go` — change package declaration to `package utils`.

Overwrite `pkg/utils/stockcode_test.go`:

```go
package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/utils"
)

func TestToTushareCode(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"sh-600", "600537", "600537.SH", false},
		{"sh-601", "601398", "601398.SH", false},
		{"sh-603", "603778", "603778.SH", false},
		{"sh-605", "605588", "605588.SH", false},
		{"sh-688-star", "688001", "688001.SH", false},
		{"sh-900-b", "900901", "900901.SH", false},
		{"sh-510-etf", "510300", "510300.SH", false},
		{"sh-515-etf", "515170", "515170.SH", false},
		{"sz-000", "000001", "000001.SZ", false},
		{"sz-001", "001872", "001872.SZ", false},
		{"sz-002", "002594", "002594.SZ", false},
		{"sz-300-gem", "300750", "300750.SZ", false},
		{"sz-200-b", "200012", "200012.SZ", false},
		{"sz-159-etf", "159915", "159915.SZ", false},
		{"passthrough-suffixed", "600537.SH", "600537.SH", false},
		{"passthrough-sz", "000890.SZ", "000890.SZ", false},
		{"too-short", "12345", "", true},
		{"ipo-subscription-730", "730001", "", true},
		{"ipo-subscription-732", "732001", "", true},
		{"unknown-prefix-400", "400001", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := utils.ToTushareCode(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
```

- [ ] **Step 4: Create `pkg/utils/math.go`**

```go
package utils

import "math"

func mean(xs []float64) float64 {
	var sum float64
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

func roundTo(v float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(v*pow) / pow
}
```

- [ ] **Step 5: Create `pkg/utils/spreads.go`**

```go
package utils

import (
	"math"

	"stock/pkg/models"
)

func ComputeSpreads(open, high, low, close float64) models.Spreads {
	return models.Spreads{
		OH: math.Abs(high - open),
		OL: math.Abs(open - low),
		HL: math.Abs(high - low),
		OC: math.Abs(open - close),
		HC: math.Abs(high - close),
		LC: math.Abs(low - close),
	}
}
```

- [ ] **Step 6: Create `pkg/utils/spreads_test.go`**

```go
package utils_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/utils"
)

func TestComputeSpreads(t *testing.T) {
	got := utils.ComputeSpreads(100.0, 105.0, 98.0, 102.0)
	assert.InDelta(t, 5.0, got.OH, 1e-9, "spread_oh")
	assert.InDelta(t, 2.0, got.OL, 1e-9, "spread_ol")
	assert.InDelta(t, 7.0, got.HL, 1e-9, "spread_hl")
	assert.InDelta(t, 2.0, got.OC, 1e-9, "spread_oc")
	assert.InDelta(t, 3.0, got.HC, 1e-9, "spread_hc")
	assert.InDelta(t, 4.0, got.LC, 1e-9, "spread_lc")
}

func TestComputeSpreads_AllAbsolute(t *testing.T) {
	got := utils.ComputeSpreads(100.0, 100.5, 95.0, 96.0)
	assert.True(t, got.OC >= 0, "spread_oc must be absolute, got %v", got.OC)
	assert.InDelta(t, 4.0, got.OC, 1e-9)
}

func TestComputeSpreads_Zero(t *testing.T) {
	got := utils.ComputeSpreads(50.0, 50.0, 50.0, 50.0)
	assert.Equal(t, 0.0, math.Abs(got.OH+got.OL+got.HL+got.OC+got.HC+got.LC))
}
```

- [ ] **Step 7: Create `pkg/utils/distribution.go`**

```go
package utils

import "sort"

type Bin struct {
	Low   float64
	High  float64
	Count int
	Pct   float64
}

func Distribution(values []float64, numBins int) []Bin {
	if len(values) == 0 {
		return nil
	}
	vmin, vmax := values[0], values[0]
	for _, v := range values {
		if v < vmin {
			vmin = v
		}
		if v > vmax {
			vmax = v
		}
	}
	if vmin == vmax {
		return []Bin{{Low: vmin, High: vmax, Count: len(values), Pct: 100.0}}
	}
	width := (vmax - vmin) / float64(numBins)
	out := make([]Bin, 0, numBins)
	for i := 0; i < numBins; i++ {
		low := vmin + float64(i)*width
		high := vmin + float64(i+1)*width
		count := 0
		if i == numBins-1 {
			for _, v := range values {
				if low <= v && v <= high {
					count++
				}
			}
		} else {
			for _, v := range values {
				if low <= v && v < high {
					count++
				}
			}
		}
		pct := roundTo(float64(count)/float64(len(values))*100.0, 1)
		out = append(out, Bin{Low: low, High: high, Count: count, Pct: pct})
	}
	return out
}
```

- [ ] **Step 8: Create `pkg/utils/distribution_test.go`**

```go
package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/utils"
)

func TestDistribution_Basic(t *testing.T) {
	bins := utils.Distribution([]float64{1, 2, 3, 4, 5}, 5)
	require.Len(t, bins, 5)
	total := 0
	for _, b := range bins {
		total += b.Count
	}
	assert.Equal(t, 5, total)
}

func TestDistribution_Empty(t *testing.T) {
	assert.Empty(t, utils.Distribution(nil, 10))
}

func TestDistribution_Single(t *testing.T) {
	bins := utils.Distribution([]float64{3.0}, 10)
	require.Len(t, bins, 1)
	assert.Equal(t, 1, bins[0].Count)
	assert.InDelta(t, 100.0, bins[0].Pct, 1e-9)
}
```

- [ ] **Step 9: Create `pkg/utils/range.go`**

```go
package utils

import (
	"math"
	"sort"
)

type Range struct {
	Low    float64
	High   float64
	CumPct float64
}

func RecommendedRange(values []float64, threshold float64) *Range {
	if len(values) == 0 {
		return nil
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	n := len(sorted)
	if n == 1 {
		return &Range{Low: sorted[0], High: sorted[0], CumPct: 100.0}
	}
	needed := int(math.Round(float64(n) * threshold / 100.0))
	if needed < 1 {
		needed = 1
	}
	bestLow := sorted[0]
	bestHigh := sorted[n-1]
	bestSpan := bestHigh - bestLow
	for i := 0; i+needed-1 < n; i++ {
		span := sorted[i+needed-1] - sorted[i]
		if span < bestSpan {
			bestSpan = span
			bestLow = sorted[i]
			bestHigh = sorted[i+needed-1]
		}
	}
	cum := roundTo(float64(needed)/float64(n)*100.0, 1)
	return &Range{Low: bestLow, High: bestHigh, CumPct: cum}
}
```

- [ ] **Step 10: Create `pkg/utils/range_test.go`**

```go
package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/utils"
)

func TestRecommendedRange_Empty(t *testing.T) {
	r := utils.RecommendedRange(nil, 60.0)
	assert.Nil(t, r)
}

func TestRecommendedRange_Single(t *testing.T) {
	r := utils.RecommendedRange([]float64{3.0}, 60.0)
	require.NotNil(t, r)
	assert.InDelta(t, 3.0, r.Low, 1e-9)
	assert.InDelta(t, 3.0, r.High, 1e-9)
	assert.InDelta(t, 100.0, r.CumPct, 1e-9)
}

func TestRecommendedRange_Sliding(t *testing.T) {
	vals := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := utils.RecommendedRange(vals, 30.0)
	require.NotNil(t, r)
	assert.InDelta(t, 2.0, r.High-r.Low, 1e-9, "tightest contiguous span of 3 values is 2.0")
}

func TestRecommendedRange_SkewedTight(t *testing.T) {
	vals := []float64{0.1, 0.2, 0.15, 0.25, 0.3, 0.35, 0.18, 0.22, 1.0, 2.0}
	r := utils.RecommendedRange(vals, 60.0)
	require.NotNil(t, r)
	assert.True(t, r.CumPct >= 60.0)
	assert.True(t, r.High-r.Low < 1.0)
}
```

---

### Task 2: Absorb `window` logic into `pkg/analysis`

**Files:**
- Create: `pkg/analysis/window.go`
- Create: `pkg/analysis/window_test.go`
- Modify: `pkg/analysis/model.go`
- Modify: `pkg/analysis/builder.go`
- Modify: `pkg/analysis/builder_test.go`
- Modify: `pkg/analysis/parity_test.go`

- [ ] **Step 1: Create `pkg/analysis/window.go`**

```go
package analysis

import (
	"sort"

	"stock/pkg/models"
)

var Names = []string{"历史", "近3月", "近1月", "近2周"}

var Days = []*int{nil, ptr(90), ptr(30), ptr(15)}

func ptr(v int) *int { return &v }

var SpreadKeys = []string{
	"spread_oh", "spread_ol", "spread_hl",
	"spread_hc", "spread_lc", "spread_oc",
}

type Window struct {
	Name string
	Rows []models.DailyBar
}

func Make(rows []models.DailyBar) []Window {
	sorted := make([]models.DailyBar, len(rows))
	copy(sorted, rows)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].TradeDate > sorted[j].TradeDate })
	out := make([]Window, len(Names))
	for i, name := range Names {
		if Days[i] == nil {
			out[i] = Window{Name: name, Rows: sorted}
			continue
		}
		end := *Days[i]
		if end > len(sorted) {
			end = len(sorted)
		}
		out[i] = Window{Name: name, Rows: sorted[:end]}
	}
	return out
}

type MeansResult map[string]map[string]*float64

func Means(windows []Window) MeansResult {
	out := make(MeansResult, len(windows))
	for _, w := range windows {
		row := make(map[string]*float64, len(SpreadKeys))
		for _, key := range SpreadKeys {
			vals := extract(w.Rows, key)
			if len(vals) == 0 {
				row[key] = nil
				continue
			}
			m := mean(vals)
			row[key] = &m
		}
		out[w.Name] = row
	}
	return out
}

func Composite(m MeansResult) map[string]float64 {
	out := make(map[string]float64, len(SpreadKeys))
	for _, key := range SpreadKeys {
		var vals []float64
		for _, name := range Names {
			if v := m[name][key]; v != nil {
				vals = append(vals, *v)
			}
		}
		if len(vals) == 0 {
			out[key] = 0.0
			continue
		}
		out[key] = mean(vals)
	}
	return out
}

func extract(rows []models.DailyBar, key string) []float64 {
	out := make([]float64, 0, len(rows))
	for _, r := range rows {
		switch key {
		case "spread_oh":
			out = append(out, r.Spreads.OH)
		case "spread_ol":
			out = append(out, r.Spreads.OL)
		case "spread_hl":
			out = append(out, r.Spreads.HL)
		case "spread_oc":
			out = append(out, r.Spreads.OC)
		case "spread_hc":
			out = append(out, r.Spreads.HC)
		case "spread_lc":
			out = append(out, r.Spreads.LC)
		}
	}
	return out
}
```

- [ ] **Step 2: Create `pkg/analysis/window_test.go`**

```go
package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/analysis"
	"stock/pkg/models"
)

func bar(date string, oh, ol, hl, oc, hc, lc float64) models.DailyBar {
	return models.DailyBar{
		TradeDate: date,
		Spreads:   models.Spreads{OH: oh, OL: ol, HL: hl, OC: oc, HC: hc, LC: lc},
	}
}

func TestWindowNamesAndDaysSync(t *testing.T) {
	assert.Equal(t, []string{"历史", "近3月", "近1月", "近2周"}, analysis.Names)
	assert.Equal(t, 4, len(analysis.Days))
	assert.Nil(t, analysis.Days[0], "first window is unbounded (历史)")
	assert.Equal(t, 90, *analysis.Days[1])
	assert.Equal(t, 30, *analysis.Days[2])
	assert.Equal(t, 15, *analysis.Days[3])
}

func TestMakeWindows_SlicesByDate(t *testing.T) {
	rows := []models.DailyBar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	got := analysis.Make(rows)
	require.Len(t, got, 4)
	assert.Equal(t, "历史", got[0].Name)
	assert.Len(t, got[0].Rows, 3)
	assert.Equal(t, "近3月", got[1].Name)
	assert.Len(t, got[1].Rows, 3)
	assert.Equal(t, "近1月", got[2].Name)
	assert.Len(t, got[2].Rows, 3)
	assert.Equal(t, "近2周", got[3].Name)
	assert.Len(t, got[3].Rows, 3)
}

func TestWindowMeans_Basic(t *testing.T) {
	rows := []models.DailyBar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	means := analysis.Means(analysis.Make(rows))
	assert.InDelta(t, 2.0, *means["历史"]["spread_oh"], 1e-9)
	assert.InDelta(t, 1.0, *means["历史"]["spread_ol"], 1e-9)
}

func TestWindowMeans_Empty(t *testing.T) {
	means := analysis.Means(analysis.Make(nil))
	for _, name := range analysis.Names {
		for _, key := range analysis.SpreadKeys {
			assert.Nil(t, means[name][key], "%s/%s should be nil", name, key)
		}
	}
}

func TestCompositeMeans_NoneTreatedAsZero(t *testing.T) {
	m := func(v float64) *float64 { return &v }
	wm := analysis.MeansResult{
		"历史":  {"spread_oh": m(4.0), "spread_ol": nil},
		"近3月": {"spread_oh": m(2.0), "spread_ol": nil},
		"近1月": {"spread_oh": m(1.0), "spread_ol": nil},
		"近2周": {"spread_oh": m(0.5), "spread_ol": nil},
	}
	comp := analysis.Composite(wm)
	assert.InDelta(t, 1.875, comp["spread_oh"], 1e-9)
	assert.Equal(t, 0.0, comp["spread_ol"], "all-None composite must collapse to 0.0")
}
```

- [ ] **Step 3: Rewrite `pkg/analysis/model.go`**

Overwrite the entire file:

```go
// Package analysis ports the Python price-spread analysis pipeline:
// window means, composite means, the spread-model table and the reference
// (predicted price) table. The output of Build is the canonical analysis
// payload returned by the HTTP API and rendered by the CLI.
package analysis

import "stock/pkg/models"

// Input is everything Build needs.
type Input struct {
	TsCode      string
	StockName   string
	Rows        []models.DailyBar // raw daily history
	OpenPrice   *float64
	ActualHigh  *float64
	ActualLow   *float64
	ActualClose *float64
}

// AnalysisResult is the canonical output. Field naming matches the design spec §3.4.
type AnalysisResult struct {
	TsCode         string             `json:"ts_code"`
	StockName      string             `json:"stock_name"`
	YesterdayClose *float64           `json:"yesterday_close,omitempty"`
	Windows        []string           `json:"windows"` // ["历史","近3月","近1月","近2周"]
	OpenPrice      *float64           `json:"open_price,omitempty"`
	ActualHigh     *float64           `json:"actual_high,omitempty"`
	ActualLow      *float64           `json:"actual_low,omitempty"`
	ActualClose    *float64           `json:"actual_close,omitempty"`
	WindowMeans    MeansResult        `json:"window_means"` // window -> spread_key -> *float64
	CompositeMeans map[string]float64 `json:"composite_means"`
	ModelTable     ModelTable         `json:"model_table"`
	ReferenceTable ReferenceTable     `json:"reference_table"`
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
```

- [ ] **Step 4: Rewrite `pkg/analysis/builder.go`**

Overwrite the entire file (drop `spread` and `window` imports; replace
`window.X` with `X`; replace `[]spread.Bar` with `[]models.DailyBar`):

```go
package analysis

import (
	"fmt"
	"strconv"
)

// ModelSpreadKeys matches analysis.py MODEL_SPREAD_KEYS exactly.
var ModelSpreadKeys = []string{
	"spread_oh", "spread_ol", "spread_hl",
	"spread_hc", "spread_lc", "spread_oc",
}

// ModelSpreadLabels matches analysis.py MODEL_SPREAD_LABELS exactly.
var ModelSpreadLabels = []string{
	"开盘与最高价", "开盘与最低价", "最高与最低价",
	"最高与收盘价", "最低与收盘价", "开盘与收盘价",
}

// Build runs the full pipeline.
func Build(in Input) AnalysisResult {
	windows := Make(in.Rows)
	wmeans := Means(windows)
	comp := Composite(wmeans)

	result := AnalysisResult{
		TsCode:         in.TsCode,
		StockName:      in.StockName,
		Windows:        Names,
		OpenPrice:      in.OpenPrice,
		ActualHigh:     in.ActualHigh,
		ActualLow:      in.ActualLow,
		ActualClose:    in.ActualClose,
		WindowMeans:    wmeans,
		CompositeMeans: comp,
		ModelTable:     buildModelTable(wmeans, comp),
	}
	if in.OpenPrice != nil {
		result.ReferenceTable = buildReferenceTable(*in.OpenPrice, in.ActualHigh, in.ActualLow, in.ActualClose, wmeans, comp)
	}
	if last := lastClose(in.Rows); last != nil {
		result.YesterdayClose = last
	}
	return result
}

func buildModelTable(wmeans MeansResult, comp map[string]float64) ModelTable {
	headers := append([]string{"时段"}, ModelSpreadLabels...)
	rows := make([][]string, 0, len(Names)+1)
	for _, wname := range Names {
		row := []string{wname}
		for _, key := range ModelSpreadKeys {
			row = append(row, formatPtr(wmeans[wname][key]))
		}
		rows = append(rows, row)
	}
	comprow := []string{"综合均值"}
	for _, key := range ModelSpreadKeys {
		comprow = append(comprow, fmt.Sprintf("%.2f", comp[key]))
	}
	rows = append(rows, comprow)
	return ModelTable{Headers: headers, Rows: rows}
}

func buildReferenceTable(openPrice float64, actualHigh, actualLow, actualClose *float64, wm MeansResult, _ map[string]float64) ReferenceTable {
	headers := []string{
		"", "历史参考价", "近3月参考价", "近1月参考价", "近2周参考价",
		"最低价反推(当日最低价)", "最高价反推(当日最高价)", "均值", "正负算一",
	}
	rows := [][]string{
		highRow(openPrice, actualLow, wm),
		lowRow(openPrice, actualHigh, wm),
		closeRow(actualHigh, actualLow, wm),
	}
	return ReferenceTable{Headers: headers, Rows: rows}
}

func highRow(openPrice float64, actualLow *float64, wm MeansResult) []string {
	row := []string{"最高价预测"}
	for _, wname := range Names {
		v := wm[wname]["spread_oh"]
		if v == nil {
			row = append(row, "/")
			continue
		}
		row = append(row, fmt.Sprintf("%.2f", openPrice+*v))
	}
	hl2w := wm["近2周"]["spread_hl"]
	if actualLow != nil && hl2w != nil {
		row = append(row, fmt.Sprintf("%.2f", *actualLow+*hl2w))
	} else {
		row = append(row, "/")
	}
	row = append(row, "/")
	row = append(row, meanOfNumericCells(row[1:5]))
	row = append(row, "+")
	return row
}

func lowRow(openPrice float64, actualHigh *float64, wm MeansResult) []string {
	row := []string{"最低价预测"}
	for _, wname := range Names {
		v := wm[wname]["spread_ol"]
		if v == nil {
			row = append(row, "/")
			continue
		}
		row = append(row, fmt.Sprintf("%.2f", openPrice-*v))
	}
	row = append(row, "/")
	hl2w := wm["近2周"]["spread_hl"]
	if actualHigh != nil && hl2w != nil {
		row = append(row, fmt.Sprintf("%.2f", *actualHigh-*hl2w))
	} else {
		row = append(row, "/")
	}
	row = append(row, meanOfNumericCells(row[1:5]))
	row = append(row, "-")
	return row
}

func closeRow(actualHigh, actualLow *float64, wm MeansResult) []string {
	row := []string{"收盘价预测", "/", "/", "/", "/"}
	lc2w := wm["近2周"]["spread_lc"]
	if actualLow != nil && lc2w != nil {
		row = append(row, fmt.Sprintf("%.2f", *actualLow+*lc2w))
	} else {
		row = append(row, "/")
	}
	hc2w := wm["近2周"]["spread_hc"]
	if actualHigh != nil && hc2w != nil {
		row = append(row, fmt.Sprintf("%.2f", *actualHigh-*hc2w))
	} else {
		row = append(row, "/")
	}
	row = append(row, meanOfNumericCells(row[5:7]))
	row = append(row, "-")
	return row
}

func meanOfNumericCells(cells []string) string {
	var nums []float64
	for _, c := range cells {
		if v, err := strconv.ParseFloat(c, 64); err == nil {
			nums = append(nums, v)
		}
	}
	if len(nums) == 0 {
		return "/"
	}
	var sum float64
	for _, v := range nums {
		sum += v
	}
	return fmt.Sprintf("%.2f", sum/float64(len(nums)))
}

func formatPtr(p *float64) string {
	if p == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *p)
}

func lastClose(rows []models.DailyBar) *float64 {
	if len(rows) == 0 {
		return nil
	}
	latest := rows[0]
	for _, r := range rows[1:] {
		if r.TradeDate > latest.TradeDate {
			latest = r
		}
	}
	v := latest.Close
	return &v
}
```

- [ ] **Step 5: Rewrite `pkg/analysis/builder_test.go`**

Current test uses `spread.Bar` and `spread.Spreads`. Replace with `models.DailyBar` and `models.Spreads`:

```go
package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/models"
)

func bar(date string, oh, ol, hl, hc, lc, oc float64) models.DailyBar {
	return models.DailyBar{
		TradeDate: date,
		Spreads:   models.Spreads{OH: oh, OL: ol, HL: hl, HC: hc, LC: lc, OC: oc},
	}
}

func TestBuild_ModelTable(t *testing.T) {
	in := Input{
		TsCode:    "600519.SH",
		StockName: "贵州茅台",
		Rows:      []models.DailyBar{bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5)},
	}
	res := Build(in)
	assert.Equal(t, "600519.SH", res.TsCode)
	assert.Equal(t, "贵州茅台", res.StockName)
	assert.NotEmpty(t, res.ModelTable.Rows)
}

func TestBuild_ReferenceTable_WithOpenPrice(t *testing.T) {
	open := 100.0
	rows := []models.DailyBar{
		bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5),
		bar("20240103", 2, 1.0, 3.0, 1.0, 2.0, 1.0),
	}
	in := Input{TsCode: "X", Rows: rows, OpenPrice: &open}
	res := Build(in)
	assert.NotEmpty(t, res.ReferenceTable.Rows)
}
```

- [ ] **Step 6: Rewrite `pkg/analysis/parity_test.go`**

Remove the `stock/pkg/shared/spread` import. Add `stock/pkg/models`. Replace
`spread.Bar` → `models.DailyBar`, `spread.Spreads` → `models.Spreads`.

The top of the file becomes:

```go
package analysis_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/analysis"
	"stock/pkg/models"
)
```

Inside `TestParity_AgainstPythonFixtures`, replace the `bars` construction loop:

```go
bars := make([]models.DailyBar, 0, len(fx.Rows))
for _, r := range fx.Rows {
	bars = append(bars, models.DailyBar{
		TsCode: r.TsCode, TradeDate: r.TradeDate,
		Open: r.Open, High: r.High, Low: r.Low, Close: r.Close,
		Vol: r.Vol, Amount: r.Amount,
		Spreads: models.Spreads{
			OH: r.SpreadOH, OL: r.SpreadOL, HL: r.SpreadHL,
			OC: r.SpreadOC, HC: r.SpreadHC, LC: r.SpreadLC,
		},
	})
}
```

The rest of the parity test body is unchanged.

---

### Task 3: Update `pkg/tushare` and `pkg/stockd/services/analysis`

**Files:**
- Modify: `pkg/tushare/daily.go`
- Modify: `pkg/tushare/daily_test.go`
- Modify: `pkg/stockd/services/analysis/analysis.go`
- Modify: `pkg/stockd/services/analysis/analysis_test.go`

- [ ] **Step 1: Rewrite `pkg/tushare/daily.go`**

Replace the import block and return type:

```go
package tushare

import (
	"context"
	"fmt"

	"stock/pkg/models"
	"stock/pkg/utils"
)
```

Replace the function body (only the return type and bar construction change):

```go
// Daily fetches OHLCV rows and returns models.DailyBar with spreads pre-computed.
func Daily(ctx context.Context, c *Client, token string, req DailyRequest) ([]models.DailyBar, error) {
	params := map[string]any{
		"ts_code":    req.TsCode,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	}
	resp, err := c.Call(ctx, token, "daily", params,
		"ts_code,trade_date,open,high,low,close,vol,amount")
	if err != nil {
		return nil, err
	}
	idx, err := indexFields(resp.Fields,
		"ts_code", "trade_date", "open", "high", "low", "close", "vol", "amount")
	if err != nil {
		return nil, err
	}
	bars := make([]models.DailyBar, 0, len(resp.Items))
	for _, row := range resp.Items {
		tsCode, _ := row[idx["ts_code"]].(string)
		tradeDate, _ := row[idx["trade_date"]].(string)
		open, _ := toFloat(row[idx["open"]])
		high, _ := toFloat(row[idx["high"]])
		low, _ := toFloat(row[idx["low"]])
		close, _ := toFloat(row[idx["close"]])
		vol, _ := toFloat(row[idx["vol"]])
		amount, _ := toFloat(row[idx["amount"]])
		bar := models.DailyBar{
			TsCode:    tsCode,
			TradeDate: tradeDate,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Vol:       vol,
			Amount:    amount,
		}
		bar.Spreads = utils.ComputeSpreads(open, high, low, close)
		bars = append(bars, bar)
	}
	return bars, nil
}
```

- [ ] **Step 2: Update `pkg/tushare/daily_test.go`**

No import changes needed (the test already only imports `stock/pkg/tushare`;
accessing `bars[0].Spreads.OH` is legal through the imported type). The only
update is the assertion on line 48 which already reads `bars[0].Spreads.OH` —
that stays valid because `models.DailyBar` embeds `Spreads` with exported fields.
No file edit is required.

- [ ] **Step 3: Rewrite `pkg/stockd/services/analysis/analysis.go`**

Replace imports:

```go
import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	pkganalysis "stock/pkg/analysis"
	"stock/pkg/models"
)
```

Replace the `Run` method body. Delete the `rows` construction loop (lines
65–73 in the original). Replace with direct pass-through:

```go
func (s *Service) Run(ctx context.Context, in Input) (pkganalysis.AnalysisResult, error) {
	var bars []models.DailyBar
	err := s.db.WithContext(ctx).Where("ts_code = ?", in.TsCode).Order("trade_date ASC").Find(&bars).Error
	if err != nil {
		return pkganalysis.AnalysisResult{}, err
	}

	if in.WithDraft {
		today := time.Now().Format("20060102")
		var d models.IntradayDraft
		err := s.db.WithContext(ctx).
			Where("user_id = ? AND ts_code = ? AND trade_date = ?", in.UserID, in.TsCode, today).
			First(&d).Error
		if err == nil {
			if in.OpenPrice == nil {
				in.OpenPrice = d.Open
			}
			if in.ActualHigh == nil {
				in.ActualHigh = d.High
			}
			if in.ActualLow == nil {
				in.ActualLow = d.Low
			}
			if in.ActualClose == nil {
				in.ActualClose = d.Close
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return pkganalysis.AnalysisResult{}, err
		}
	}

	var name string
	var st models.Stock
	if s.db.WithContext(ctx).First(&st, "ts_code = ?", in.TsCode).Error == nil {
		name = st.Name
	}

	res := pkganalysis.Build(pkganalysis.Input{
		TsCode: in.TsCode, StockName: name,
		Rows:        bars,
		OpenPrice:   in.OpenPrice,
		ActualHigh:  in.ActualHigh,
		ActualLow:   in.ActualLow,
		ActualClose: in.ActualClose,
	})
	return res, nil
}
```

- [ ] **Step 4: Update `pkg/stockd/services/analysis/analysis_test.go`**

Change import `stock/pkg/stockd/models` → `stock/pkg/models`.

Line 30 (the inline `DailyBar` creation) changes from:

```go
require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250513", Open: 100, High: 102, Low: 98, Close: 101, SpreadOH: 2, SpreadOL: 2, SpreadHL: 4}).Error)
```

To:

```go
require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250513", Open: 100, High: 102, Low: 98, Close: 101, Spreads: models.Spreads{OH: 2, OL: 2, HL: 4}}).Error)
```

Line 43 (second `DailyBar` creation without spreads) stays valid — the zero
value of `Spreads` is acceptable.

---

### Task 4: Bulk import rewrites

**Files:** 19 non-test + 1 test files under `pkg/stockd/`, plus 11 files under
`pkg/stockctl/` and `cmd/stockctl/`.

- [ ] **Step 1: Rewrite `stock/pkg/stockd/models` → `stock/pkg/models`**

Run:

```bash
grep -rl '"stock/pkg/stockd/models"' pkg/stockd/ | xargs sed -i 's|"stock/pkg/stockd/models"|"stock/pkg/models"|g'
```

Verify the list matches the 19 files from the spec (excluding the moved
`models_test.go`):

```bash
grep -rl '"stock/pkg/stockd/models"' pkg/stockd/
```

Expected:
```
pkg/stockd/auth/auth_test.go
pkg/stockd/auth/middleware.go
pkg/stockd/bootstrap/bootstrap.go
pkg/stockd/bootstrap/bootstrap_test.go
pkg/stockd/db/db.go
pkg/stockd/http/handler/auth_test.go
pkg/stockd/services/analysis/analysis.go
pkg/stockd/services/analysis/analysis_test.go
pkg/stockd/services/bars/bars.go
pkg/stockd/services/bars/bars_test.go
pkg/stockd/services/draft/draft.go
pkg/stockd/services/portfolio/portfolio.go
pkg/stockd/services/scheduler/scheduler.go
pkg/stockd/services/scheduler/scheduler_test.go
pkg/stockd/services/stock/csv.go
pkg/stockd/services/stock/stock.go
pkg/stockd/services/stock/stock_test.go
pkg/stockd/services/token/token.go
pkg/stockd/services/user/user.go
```

- [ ] **Step 2: Rename `pkg/stockctl` → `pkg/cli` and rewrite imports**

```bash
git mv pkg/stockctl pkg/cli
grep -rl '"stock/pkg/stockctl/' pkg/cli/ cmd/stockctl/ | xargs sed -i 's|"stock/pkg/stockctl/|"stock/pkg/cli/|g'
```

Verify:

```bash
grep -rn '"stock/pkg/stockctl' pkg/cli/ cmd/stockctl/ || echo "Clean"
```

Expected: no matches (or only commented-out lines, if any).

---

### Task 5: Add GORM guard test and delete `pkg/shared/`

**Files:**
- Modify: `pkg/models/models_test.go`
- Delete: `pkg/shared/` (entire directory)

- [ ] **Step 1: Add `TestDailyBarHasSpreadColumns` to `pkg/models/models_test.go`**

Append this test to the existing file:

```go
func TestDailyBarHasSpreadColumns(t *testing.T) {
	gdb := openTestDB(t)
	cols := []string{"spread_oh", "spread_ol", "spread_hl", "spread_oc", "spread_hc", "spread_lc"}
	for _, col := range cols {
		assert.True(t, gdb.Migrator().HasColumn(&DailyBar{}, col), "column %s should exist", col)
	}

	bar := DailyBar{
		TsCode: "000001.SZ", TradeDate: "20250513",
		Open: 10, High: 11, Low: 9, Close: 10.5, Vol: 1000, Amount: 1e4,
		Spreads: Spreads{OH: 1, OL: 1, HL: 2, OC: 0.5, HC: 0.5, LC: 1.5},
	}
	require.NoError(t, gdb.Create(&bar).Error)

	var got DailyBar
	require.NoError(t, gdb.First(&got, "ts_code = ? AND trade_date = ?", bar.TsCode, bar.TradeDate).Error)
	assert.Equal(t, bar.Spreads, got.Spreads)
}
```

Also update the import block in `pkg/models/models_test.go`. It currently imports:

```go
"stock/pkg/stockd/db"
"stock/pkg/stockd/models"
```

After the move, replace with:

```go
"github.com/stretchr/testify/assert"
"stock/pkg/stockd/db"
```

Remove the `"stock/pkg/stockd/models"` import (the test is now in the same
package if using `package models`, or imports `stock/pkg/models` if using
`package models_test`). The current test file uses `package models_test`, so it
needs to import `stock/pkg/models` to reference `models.DailyBar` etc.

Wait — after moving the file to `pkg/models/models_test.go`, if it keeps
`package models_test`, it must import `stock/pkg/models`. Update the imports:

```go
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/stockd/db"
)
```

And all references to `models.DailyBar` stay as `models.DailyBar` (same as
before, just the import path changed).

- [ ] **Step 2: Delete `pkg/shared/`**

```bash
git rm -r pkg/shared/
```

---

### Task 6: Final verification and atomic commit

- [ ] **Step 1: Build**

```bash
go build ./...
```

Expected: clean (no output, exit 0).

- [ ] **Step 2: Vet**

```bash
go vet ./...
```

Expected: clean.

- [ ] **Step 3: Run all tests**

```bash
go test -race ./...
```

Expected: all PASS. Pay special attention to:
- `pkg/analysis/parity_test.go` (must match Python fixtures)
- `pkg/models/models_test.go` (new GORM column-name test)
- `pkg/utils/*_test.go` (new package tests)
- `pkg/analysis/window_test.go` (migrated from shared/window)

- [ ] **Step 4: Optional security scan**

```bash
# if gosec is installed
gosec ./...
```

Expected: no new findings introduced by the refactor.

- [ ] **Step 5: Stage and commit atomically**

```bash
git add -A
git status
```

Verify the staged changes include:
- `pkg/models/` (new files + moved files)
- `pkg/utils/` (new files + moved stockcode)
- `pkg/analysis/window.go` + `window_test.go`
- Modified `pkg/analysis/builder.go`, `model.go`, `builder_test.go`, `parity_test.go`
- Modified `pkg/tushare/daily.go`
- Modified `pkg/stockd/services/analysis/analysis.go`, `analysis_test.go`
- Modified `pkg/stockd/db/db.go`, `auth/middleware.go`, `bootstrap/bootstrap.go`, etc. (import rewrites)
- Renamed `pkg/cli/` (from `pkg/stockctl/`)
- Modified `cmd/stockctl/main.go`
- Deleted `pkg/shared/`

Commit:

```bash
git commit -m "refactor: flatten pkg/shared, unify DailyBar, rename stockctl->cli"
```

---

## Self-Review Checklist

- [ ] **Spec coverage:** Every §3–§4 item in the design spec has a corresponding task/step above.
- [ ] **Placeholder scan:** No TBD, TODO, "implement later", "fill in details", "similar to Task N", or "add appropriate error handling" anywhere in the plan.
- [ ] **Type consistency:** `models.DailyBar`, `models.Spreads`, `utils.ComputeSpreads`, `analysis.MeansResult`, `analysis.Make`, `utils.Distribution`, `utils.RecommendedRange` are used with the same signatures across all tasks.
- [ ] **Import sanity:** After all steps, no file should import `stock/pkg/shared/*` or `stock/pkg/stockd/models` or `stock/pkg/stockctl/*`.
- [ ] **Test completeness:** Every new `.go` file has a corresponding `_test.go` file, and every migrated test preserves its original assertions.
- [ ] **File paths:** All paths are exact and relative to repo root.
