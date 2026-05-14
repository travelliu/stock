# P1 — Shared Libraries Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the Python analysis stack — `to_tushare_code`, `compute_spreads`, window slicing, the Tushare HTTP client, and the spread-model + reference table builders — into reusable Go packages with **bit-for-bit parity** asserted against fixtures captured from the existing Python implementation.

**Architecture:** Five independent packages under `pkg/` that the server (P2–P4) and CLI (P5) will compose. The packages know nothing about HTTP, GORM, or config; they take primitives in and return primitives out. Token is passed per-call to `pkg/tushare` so each user's override works.

**Tech Stack:** Go stdlib, `github.com/stretchr/testify`. No third-party math libraries — the Python implementation uses `statistics` stdlib, so Go's `math` and a tiny helper are sufficient.

**Reference spec:** `docs/superpowers/specs/2026-05-14-go-vue-rewrite-design.md` §1.1, §7.1, §7.2 (P1).
**Reference Python:** `config.py:36-68` (to_tushare_code), `fetcher.py:9-28` (compute_spreads), `analysis.py:9-273, 316-422` (window/means/tables), `analysis.py:113-153` (distribution), `analysis.py:155-194` (recommended range), `analysis.py:289-314` (header), `analysis.py:424-443` (trading plan).

---

## File overview

| File | Responsibility |
|------|----------------|
| `pkg/shared/stockcode/stockcode.go` | `ToTushareCode(code string) (string, error)`, market prefix maps |
| `pkg/shared/stockcode/stockcode_test.go` | Table-driven tests for all valid prefixes + error cases |
| `pkg/shared/spread/spread.go` | `Spreads` struct + `Compute(OHLC) Spreads`; `Bar` struct exposes both OHLCV and computed spreads |
| `pkg/shared/spread/spread_test.go` | Numerical tests with golden values |
| `pkg/shared/window/window.go` | `MakeWindows(rowsDesc []Bar) []Window`; `WindowMeans(...)`; `CompositeMeans(...)`; sliding-window `RecommendedRange`; histogram `Distribution` |
| `pkg/shared/window/window_test.go` | Mirror of `tests/test_analysis.py` cases |
| `pkg/tushare/client.go` | HTTP client (token per-call), retry+backoff, JSON body parser |
| `pkg/tushare/daily.go` | `Daily(ctx, token, req)` API binding |
| `pkg/tushare/stock_basic.go` | `StockBasic(ctx, token, req)` API binding |
| `pkg/tushare/*_test.go` | `httptest.Server`-driven tests for each endpoint, retry behaviour |
| `pkg/analysis/model.go` | Types: `WindowMeans`, `CompositeMeans`, `ModelTable`, `ReferenceTable`, `TradingPlan`, `AnalysisResult` |
| `pkg/analysis/builder.go` | `Build(input Input) AnalysisResult` — the entrypoint the server's analysis service will call |
| `pkg/analysis/format.go` | CJK-aware string helpers for the CLI to render text tables (port of `_format_table`, `_display_width`, `_rpad`, `_lpad`, `_join_tables_side_by_side`) |
| `pkg/analysis/testdata/*.json` | Parity fixtures captured from Python |
| `pkg/analysis/*_test.go` | Unit tests + parity tests (compare Go output to fixture, byte-equal for numeric fields) |
| `tools/dump_python_fixture.py` | One-shot fixture generator using the legacy Python code; run it once during this phase |

Crucial design decisions:

- **No floats in JSON parity fixtures use raw `float64`** — we compare to **4 decimal places** for `mean`/`median` numerical fields and to **2 decimal places** for rendered cell strings (which is what the Python implementation prints). Exact byte equality on rendered strings is the target.
- **Window names are kept as the Chinese strings** (`"历史"`, `"近3月"`, `"近1月"`, `"近2周"`) to match the Python output verbatim. Keep them as exported constants.
- **`spread` package** stores spreads as absolute values, matching `fetcher.py:21-27`.
- **`pkg/analysis/format.go`** is needed by the CLI renderer (P5). The server (P4) only ships JSON; it does NOT render tables.

---

### Task 5: `pkg/shared/stockcode` — port `to_tushare_code()`

**Files:**
- Create: `pkg/shared/stockcode/stockcode.go`
- Test: `pkg/shared/stockcode/stockcode_test.go`

**Python reference:** `config.py:30-68`.

- [ ] **Step 1: Write the failing test**

Create `pkg/shared/stockcode/stockcode_test.go`:
```go
package stockcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/shared/stockcode"
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
			got, err := stockcode.ToTushareCode(tc.in)
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

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/shared/stockcode/... -run TestToTushareCode -v`
Expected: build error (`undefined: stockcode.ToTushareCode`).

- [ ] **Step 3: Write minimal implementation**

Create `pkg/shared/stockcode/stockcode.go`:
```go
// Package stockcode converts plain A-share codes (e.g. "600537") into
// Tushare-suffixed codes ("600537.SH" / "000890.SZ").
package stockcode

import (
	"fmt"
	"strings"
)

var shPrefixes = map[string]struct{}{
	"600": {}, "601": {}, "603": {}, "605": {}, "688": {},
	"900": {},
	"510": {}, "511": {}, "512": {}, "513": {}, "515": {},
}

var szPrefixes = map[string]struct{}{
	"000": {}, "001": {}, "002": {}, "300": {},
	"200": {},
	"159": {},
}

// ToTushareCode converts a plain 6-digit A-share code to the Tushare suffix
// form. Already-suffixed inputs (containing ".") pass through unchanged.
func ToTushareCode(code string) (string, error) {
	if strings.Contains(code, ".") {
		return code, nil
	}
	if len(code) < 6 {
		return "", fmt.Errorf("invalid stock code %q: must be 6 digits", code)
	}
	prefix := code[:3]
	if _, ok := shPrefixes[prefix]; ok {
		return code + ".SH", nil
	}
	if _, ok := szPrefixes[prefix]; ok {
		return code + ".SZ", nil
	}
	return "", fmt.Errorf("cannot determine market for stock code %q (prefix %q is not a known SH/SZ prefix)", code, prefix)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/shared/stockcode/... -v`
Expected: `PASS` for all 19 sub-tests.

- [ ] **Step 5: Commit**

```bash
git add pkg/shared/stockcode/
git commit -m "feat(stockcode): port to_tushare_code from Python config.py"
```

---

### Task 6: `pkg/shared/spread` — port `compute_spreads()`

**Files:**
- Create: `pkg/shared/spread/spread.go`
- Test: `pkg/shared/spread/spread_test.go`

**Python reference:** `fetcher.py:9-28`.

- [ ] **Step 1: Write the failing test**

Create `pkg/shared/spread/spread_test.go`:
```go
package spread_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/shared/spread"
)

func TestCompute(t *testing.T) {
	in := spread.OHLC{Open: 100.0, High: 105.0, Low: 98.0, Close: 102.0}
	got := spread.Compute(in)

	assert.InDelta(t, 5.0, got.OH, 1e-9, "spread_oh")
	assert.InDelta(t, 2.0, got.OL, 1e-9, "spread_ol")
	assert.InDelta(t, 7.0, got.HL, 1e-9, "spread_hl")
	assert.InDelta(t, 2.0, got.OC, 1e-9, "spread_oc") // |100-102|
	assert.InDelta(t, 3.0, got.HC, 1e-9, "spread_hc") // |105-102|
	assert.InDelta(t, 4.0, got.LC, 1e-9, "spread_lc") // |98-102|
}

func TestCompute_AllAbsolute(t *testing.T) {
	// Down day: close < open, but spread_oc must be positive (absolute).
	in := spread.OHLC{Open: 100.0, High: 100.5, Low: 95.0, Close: 96.0}
	got := spread.Compute(in)
	assert.True(t, got.OC >= 0, "spread_oc must be absolute, got %v", got.OC)
	assert.InDelta(t, 4.0, got.OC, 1e-9)
}

func TestCompute_Zero(t *testing.T) {
	in := spread.OHLC{Open: 50.0, High: 50.0, Low: 50.0, Close: 50.0}
	got := spread.Compute(in)
	assert.Equal(t, 0.0, math.Abs(got.OH+got.OL+got.HL+got.OC+got.HC+got.LC))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/shared/spread/... -v`
Expected: build error (`undefined: spread.OHLC`).

- [ ] **Step 3: Write minimal implementation**

Create `pkg/shared/spread/spread.go`:
```go
// Package spread computes the six absolute price spreads used by the analysis pipeline.
package spread

import "math"

// OHLC is the input for spread computation.
type OHLC struct {
	Open, High, Low, Close float64
}

// Spreads holds the six absolute spreads.
type Spreads struct {
	OH float64 // |high - open|
	OL float64 // |open - low|
	HL float64 // |high - low|
	OC float64 // |open - close|
	HC float64 // |high - close|
	LC float64 // |low  - close|
}

// Compute returns the six absolute spreads for one bar.
func Compute(b OHLC) Spreads {
	return Spreads{
		OH: math.Abs(b.High - b.Open),
		OL: math.Abs(b.Open - b.Low),
		HL: math.Abs(b.High - b.Low),
		OC: math.Abs(b.Open - b.Close),
		HC: math.Abs(b.High - b.Close),
		LC: math.Abs(b.Low - b.Close),
	}
}

// Bar combines OHLCV with trade date and computed spreads. Used by downstream
// packages (analysis, db) as the canonical row representation.
type Bar struct {
	TsCode    string
	TradeDate string // YYYYMMDD (Tushare style); DB layer may also accept YYYY-MM-DD
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Vol       float64
	Amount    float64
	Spreads   Spreads
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/shared/spread/... -v`
Expected: all three tests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/shared/spread/
git commit -m "feat(spread): port compute_spreads and define Bar struct"
```

---

### Task 7: `pkg/shared/window` — windowing, means, distribution, recommended-range

**Files:**
- Create: `pkg/shared/window/window.go`
- Test: `pkg/shared/window/window_test.go`

**Python reference:** `analysis.py:73-194, 263-287`.

- [ ] **Step 1: Write the failing test**

Create `pkg/shared/window/window_test.go` (mirrors `tests/test_analysis.py`):
```go
package window_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/shared/spread"
	"stock/pkg/shared/window"
)

func bar(date string, oh, ol, hl, oc, hc, lc float64) spread.Bar {
	return spread.Bar{
		TradeDate: date,
		Spreads:   spread.Spreads{OH: oh, OL: ol, HL: hl, OC: oc, HC: hc, LC: lc},
	}
}

func TestWindowNamesAndDaysSync(t *testing.T) {
	assert.Equal(t, []string{"历史", "近3月", "近1月", "近2周"}, window.Names)
	assert.Equal(t, 4, len(window.Days))
	assert.Nil(t, window.Days[0], "first window is unbounded (历史)")
	assert.Equal(t, 90, *window.Days[1])
	assert.Equal(t, 30, *window.Days[2])
	assert.Equal(t, 15, *window.Days[3])
}

func TestMakeWindows_SlicesByDate(t *testing.T) {
	rows := []spread.Bar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	got := window.Make(rows)
	require.Len(t, got, 4)
	assert.Equal(t, "历史", got[0].Name)
	assert.Len(t, got[0].Rows, 3)
	assert.Equal(t, "近3月", got[1].Name)
	assert.Len(t, got[1].Rows, 3) // fewer than 90 rows, so all included
	assert.Equal(t, "近1月", got[2].Name)
	assert.Len(t, got[2].Rows, 3)
	assert.Equal(t, "近2周", got[3].Name)
	assert.Len(t, got[3].Rows, 3)
}

func TestWindowMeans_Basic(t *testing.T) {
	rows := []spread.Bar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	means := window.Means(window.Make(rows))
	assert.InDelta(t, 2.0, *means["历史"]["spread_oh"], 1e-9)
	assert.InDelta(t, 1.0, *means["历史"]["spread_ol"], 1e-9)
}

func TestWindowMeans_Empty(t *testing.T) {
	means := window.Means(window.Make(nil))
	for _, name := range window.Names {
		for _, key := range window.SpreadKeys {
			assert.Nil(t, means[name][key], "%s/%s should be nil", name, key)
		}
	}
}

func TestCompositeMeans_NoneTreatedAsZero(t *testing.T) {
	// Python: composite = mean of non-None window means; all-None -> 0.0
	m := func(v float64) *float64 { return &v }
	wm := window.MeansResult{
		"历史":  {"spread_oh": m(4.0), "spread_ol": nil},
		"近3月": {"spread_oh": m(2.0), "spread_ol": nil},
		"近1月": {"spread_oh": m(1.0), "spread_ol": nil},
		"近2周": {"spread_oh": m(0.5), "spread_ol": nil},
	}
	comp := window.Composite(wm)
	assert.InDelta(t, 1.875, comp["spread_oh"], 1e-9)
	assert.Equal(t, 0.0, comp["spread_ol"], "all-None composite must collapse to 0.0")
}

func TestRecommendedRange_Empty(t *testing.T) {
	r := window.RecommendedRange(nil, 60.0)
	assert.Nil(t, r)
}

func TestRecommendedRange_Single(t *testing.T) {
	r := window.RecommendedRange([]float64{3.0}, 60.0)
	require.NotNil(t, r)
	assert.InDelta(t, 3.0, r.Low, 1e-9)
	assert.InDelta(t, 3.0, r.High, 1e-9)
	assert.InDelta(t, 100.0, r.CumPct, 1e-9)
}

func TestRecommendedRange_Sliding(t *testing.T) {
	vals := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := window.RecommendedRange(vals, 30.0)
	require.NotNil(t, r)
	assert.InDelta(t, 2.0, r.High-r.Low, 1e-9, "tightest contiguous span of 3 values is 2.0")
}

func TestRecommendedRange_SkewedTight(t *testing.T) {
	vals := []float64{0.1, 0.2, 0.15, 0.25, 0.3, 0.35, 0.18, 0.22, 1.0, 2.0}
	r := window.RecommendedRange(vals, 60.0)
	require.NotNil(t, r)
	assert.True(t, r.CumPct >= 60.0)
	assert.True(t, r.High-r.Low < 1.0)
}

func TestDistribution_Basic(t *testing.T) {
	bins := window.Distribution([]float64{1, 2, 3, 4, 5}, 5)
	require.Len(t, bins, 5)
	total := 0
	for _, b := range bins {
		total += b.Count
	}
	assert.Equal(t, 5, total)
}

func TestDistribution_Empty(t *testing.T) {
	assert.Empty(t, window.Distribution(nil, 10))
}

func TestDistribution_Single(t *testing.T) {
	bins := window.Distribution([]float64{3.0}, 10)
	require.Len(t, bins, 1)
	assert.Equal(t, 1, bins[0].Count)
	assert.InDelta(t, 100.0, bins[0].Pct, 1e-9)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/shared/window/... -v`
Expected: build error (`undefined: window.Names`, etc.).

- [ ] **Step 3: Write the implementation**

Create `pkg/shared/window/window.go`:
```go
// Package window slices rows into the four time windows and computes window
// means, composite means, distribution bins, and tight recommended ranges.
package window

import (
	"math"
	"sort"

	"stock/pkg/shared/spread"
)

// Names are the user-facing window labels, in display order.
var Names = []string{"历史", "近3月", "近1月", "近2周"}

// Days is the row-count slice for each window. A nil entry means "all rows".
var Days = []*int{nil, ptr(90), ptr(30), ptr(15)}

func ptr(v int) *int { return &v }

// SpreadKeys is the canonical ordering used by means / composite / model table.
// Matches analysis.py MODEL_SPREAD_KEYS.
var SpreadKeys = []string{
	"spread_oh", "spread_ol", "spread_hl",
	"spread_hc", "spread_lc", "spread_oc",
}

// Window is a named slice of rows (sorted descending by trade_date by Make).
type Window struct {
	Name string
	Rows []spread.Bar
}

// Make sorts the rows descending by TradeDate and returns the four windows.
func Make(rows []spread.Bar) []Window {
	sorted := make([]spread.Bar, len(rows))
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

// MeansResult maps windowName -> spreadKey -> *float64 (nil when no data).
type MeansResult map[string]map[string]*float64

// Means computes the per-window means for each spread key. Nil entries are
// returned when a window has no observations for a key.
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

// Composite averages the non-nil window means per spread key. An all-nil
// spread key collapses to 0.0 to match the Python behaviour.
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

// Range is the result of the sliding-window tight-range search.
type Range struct {
	Low    float64
	High   float64
	CumPct float64
}

// RecommendedRange returns the narrowest contiguous interval covering at least
// `threshold` percent of observations. Returns nil when input is empty.
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

// Bin is one histogram bucket.
type Bin struct {
	Low   float64
	High  float64
	Count int
	Pct   float64
}

// Distribution returns `numBins` equal-width buckets of `values`. Pct is
// rounded to 1 decimal place (matches Python `round(pct, 1)`).
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

func extract(rows []spread.Bar, key string) []float64 {
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

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/shared/window/... -v`
Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/shared/window/
git commit -m "feat(window): port window slicing, means, distribution, recommended range"
```

---

### Task 8: `pkg/tushare` — HTTP client + `Daily` + `StockBasic`

**Files:**
- Create: `pkg/tushare/client.go`
- Create: `pkg/tushare/daily.go`
- Create: `pkg/tushare/stock_basic.go`
- Test: `pkg/tushare/client_test.go`
- Test: `pkg/tushare/daily_test.go`
- Test: `pkg/tushare/stock_basic_test.go`

**Python reference:** `fetcher.py:31-75` (daily), `company.py:47-56` (stock_basic).

The Tushare protocol: HTTP POST to `http://api.tushare.pro` with JSON body `{"api_name": "...", "token": "...", "params": {...}, "fields": "..."}`. Response is `{"code": 0, "msg": "", "data": {"fields": [...], "items": [[...],[...]] }}`. Code != 0 is an API error.

- [ ] **Step 1: Write the failing client test**

Create `pkg/tushare/client_test.go`:
```go
package tushare_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/tushare"
)

func TestClient_PostJSON_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "daily", req["api_name"])
		assert.Equal(t, "tok-123", req["token"])
		_, _ = io.WriteString(w, `{"code":0,"msg":"","data":{"fields":["a","b"],"items":[[1,2]]}}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL), tushare.WithTimeout(2*time.Second))
	resp, err := c.Call(context.Background(), "tok-123", "daily", map[string]any{"ts_code": "600000.SH"}, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, resp.Fields)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, []any{float64(1), float64(2)}, resp.Items[0])
}

func TestClient_RetriesOn5xxThenFails(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := tushare.NewClient(
		tushare.WithBaseURL(srv.URL),
		tushare.WithTimeout(500*time.Millisecond),
		tushare.WithMaxRetries(2),
		tushare.WithRetryDelay(1*time.Millisecond),
	)
	_, err := c.Call(context.Background(), "tok", "daily", nil, "")
	require.Error(t, err)
	assert.Equal(t, 3, calls, "1 try + 2 retries")
	assert.Contains(t, err.Error(), "500")
}

func TestClient_APIErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":40203,"msg":"token invalid","data":null}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL))
	_, err := c.Call(context.Background(), "bad", "daily", nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token invalid")
}

func TestClient_DefaultBaseURL(t *testing.T) {
	c := tushare.NewClient()
	assert.True(t, strings.HasPrefix(c.BaseURL(), "http"))
}
```

- [ ] **Step 2: Run the test (expect compile failure)**

Run: `go test ./pkg/tushare/... -v`
Expected: `undefined: tushare.NewClient`, etc.

- [ ] **Step 3: Write the client implementation**

Create `pkg/tushare/client.go`:
```go
// Package tushare is a minimal SDK for the Tushare Pro JSON API.
package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const DefaultBaseURL = "http://api.tushare.pro"

// Response is the decoded Tushare envelope. Items is row-major.
type Response struct {
	Fields []string `json:"fields"`
	Items  [][]any  `json:"items"`
}

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// Client is a thin Tushare Pro HTTP client. Construct with NewClient.
type Client struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

// Option configures the client.
type Option func(*Client)

func WithBaseURL(u string) Option        { return func(c *Client) { c.baseURL = u } }
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.httpClient.Timeout = d } }
func WithMaxRetries(n int) Option        { return func(c *Client) { c.maxRetries = n } }
func WithRetryDelay(d time.Duration) Option {
	return func(c *Client) { c.retryDelay = d }
}

// NewClient returns a Tushare client with sensible defaults.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		maxRetries: 2,
		retryDelay: 500 * time.Millisecond,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// BaseURL returns the configured base URL (used in tests).
func (c *Client) BaseURL() string { return c.baseURL }

// Call performs a Tushare API request. `token` is passed per call so each
// caller can use a different token.
func (c *Client) Call(ctx context.Context, token, apiName string, params map[string]any, fields string) (*Response, error) {
	body, err := json.Marshal(map[string]any{
		"api_name": apiName,
		"token":    token,
		"params":   params,
		"fields":   fields,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			jitter := time.Duration(rand.Int63n(int64(c.retryDelay)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay*time.Duration(attempt) + jitter):
			}
		}
		resp, err := c.doOnce(ctx, body)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryable(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *Client) doOnce(ctx context.Context, body []byte) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, retryable{err: err}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, retryable{err: err}
	}
	if resp.StatusCode >= 500 {
		return nil, retryable{err: fmt.Errorf("tushare http %d: %s", resp.StatusCode, string(raw))}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("tushare http %d: %s", resp.StatusCode, string(raw))
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode envelope: %w; body: %s", err, string(raw))
	}
	if env.Code != 0 {
		return nil, fmt.Errorf("tushare api error %d: %s", env.Code, env.Msg)
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return &Response{}, nil
	}
	var out Response
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return nil, fmt.Errorf("decode data: %w", err)
	}
	return &out, nil
}

type retryable struct{ err error }

func (r retryable) Error() string { return r.err.Error() }
func (r retryable) Unwrap() error { return r.err }

func isRetryable(err error) bool {
	var r retryable
	return errors.As(err, &r)
}
```

- [ ] **Step 4: Verify client tests pass**

Run: `go test ./pkg/tushare/ -run TestClient -v`
Expected: PASS.

- [ ] **Step 5: Write the daily test**

Create `pkg/tushare/daily_test.go`:
```go
package tushare_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/tushare"
)

func TestDaily_ParsesItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "daily", req["api_name"])
		params := req["params"].(map[string]any)
		assert.Equal(t, "600519.SH", params["ts_code"])

		_, _ = io.WriteString(w, `{
		  "code":0,"msg":"","data":{
		    "fields":["ts_code","trade_date","open","high","low","close","vol","amount"],
		    "items":[
		      ["600519.SH","20250513",1620.0,1655.0,1601.0,1632.0,3500.0,500000.0],
		      ["600519.SH","20250512",1610.0,1640.0,1590.0,1620.0,3300.0,480000.0]
		    ]
		  }
		}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL))
	bars, err := tushare.Daily(context.Background(), c, "tok", tushare.DailyRequest{
		TsCode:    "600519.SH",
		StartDate: "20250101",
		EndDate:   "20250513",
	})
	require.NoError(t, err)
	require.Len(t, bars, 2)
	assert.Equal(t, "20250513", bars[0].TradeDate)
	assert.InDelta(t, 1620.0, bars[0].Open, 1e-9)
	assert.InDelta(t, 35.0, bars[0].Spreads.OH, 1e-9) // |1655 - 1620|
}
```

- [ ] **Step 6: Write `pkg/tushare/daily.go`**

```go
package tushare

import (
	"context"
	"fmt"

	"stock/pkg/shared/spread"
)

// DailyRequest mirrors Tushare `daily` parameters used by the project.
type DailyRequest struct {
	TsCode    string
	StartDate string // YYYYMMDD
	EndDate   string // YYYYMMDD
}

// Daily fetches OHLCV rows and returns spread.Bars with spreads pre-computed.
// Rows are returned in the order Tushare provides them (newest first); callers
// re-sort as needed.
func Daily(ctx context.Context, c *Client, token string, req DailyRequest) ([]spread.Bar, error) {
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
	bars := make([]spread.Bar, 0, len(resp.Items))
	for _, row := range resp.Items {
		tsCode, _ := row[idx["ts_code"]].(string)
		tradeDate, _ := row[idx["trade_date"]].(string)
		open, _ := toFloat(row[idx["open"]])
		high, _ := toFloat(row[idx["high"]])
		low, _ := toFloat(row[idx["low"]])
		close, _ := toFloat(row[idx["close"]])
		vol, _ := toFloat(row[idx["vol"]])
		amount, _ := toFloat(row[idx["amount"]])
		bar := spread.Bar{
			TsCode:    tsCode,
			TradeDate: tradeDate,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Vol:       vol,
			Amount:    amount,
		}
		bar.Spreads = spread.Compute(spread.OHLC{Open: open, High: high, Low: low, Close: close})
		bars = append(bars, bar)
	}
	return bars, nil
}

func indexFields(fields []string, want ...string) (map[string]int, error) {
	out := make(map[string]int, len(want))
	for i, f := range fields {
		out[f] = i
	}
	for _, w := range want {
		if _, ok := out[w]; !ok {
			return nil, fmt.Errorf("tushare response missing field %q (got %v)", w, fields)
		}
	}
	return out, nil
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case nil:
		return 0, false
	}
	return 0, false
}
```

- [ ] **Step 7: Verify daily test**

Run: `go test ./pkg/tushare/ -run TestDaily -v`
Expected: PASS.

- [ ] **Step 8: Write stock_basic test + impl**

Create `pkg/tushare/stock_basic_test.go`:
```go
package tushare_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/tushare"
)

func TestStockBasic_ParsesItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
		  "code":0,"msg":"","data":{
		    "fields":["ts_code","symbol","name","area","industry","market","exchange","list_date"],
		    "items":[
		      ["600519.SH","600519","贵州茅台","贵州","白酒","主板","SSE","20010827"],
		      ["000001.SZ","000001","平安银行","深圳","银行","主板","SZSE","19910403"]
		    ]
		  }
		}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL))
	out, err := tushare.StockBasic(context.Background(), c, "tok", tushare.StockBasicRequest{})
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "600519.SH", out[0].TsCode)
	assert.Equal(t, "贵州茅台", out[0].Name)
	assert.Equal(t, "SSE", out[0].Exchange)
}
```

Create `pkg/tushare/stock_basic.go`:
```go
package tushare

import "context"

// StockBasicRow mirrors the columns used by the catalog.
type StockBasicRow struct {
	TsCode   string
	Symbol   string
	Name     string
	Area     string
	Industry string
	Market   string
	Exchange string
	ListDate string
}

// StockBasicRequest is the (optional) filter set used in this project.
type StockBasicRequest struct {
	TsCode     string
	ListStatus string // "L" listed, "D" delisted, "P" paused. Empty = all listed.
}

// StockBasic returns the catalog rows. Pagination should be added if/when
// Tushare starts capping a single response (currently 5000 rows).
func StockBasic(ctx context.Context, c *Client, token string, req StockBasicRequest) ([]StockBasicRow, error) {
	params := map[string]any{}
	if req.TsCode != "" {
		params["ts_code"] = req.TsCode
	}
	if req.ListStatus != "" {
		params["list_status"] = req.ListStatus
	}
	resp, err := c.Call(ctx, token, "stock_basic", params,
		"ts_code,symbol,name,area,industry,market,exchange,list_date")
	if err != nil {
		return nil, err
	}
	idx, err := indexFields(resp.Fields,
		"ts_code", "symbol", "name", "area", "industry", "market", "exchange", "list_date")
	if err != nil {
		return nil, err
	}
	out := make([]StockBasicRow, 0, len(resp.Items))
	for _, row := range resp.Items {
		out = append(out, StockBasicRow{
			TsCode:   strOrEmpty(row[idx["ts_code"]]),
			Symbol:   strOrEmpty(row[idx["symbol"]]),
			Name:     strOrEmpty(row[idx["name"]]),
			Area:     strOrEmpty(row[idx["area"]]),
			Industry: strOrEmpty(row[idx["industry"]]),
			Market:   strOrEmpty(row[idx["market"]]),
			Exchange: strOrEmpty(row[idx["exchange"]]),
			ListDate: strOrEmpty(row[idx["list_date"]]),
		})
	}
	return out, nil
}

func strOrEmpty(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
```

- [ ] **Step 9: Verify all tushare tests pass**

Run: `go test ./pkg/tushare/ -v`
Expected: PASS.

- [ ] **Step 10: Commit**

```bash
git add pkg/tushare/
git commit -m "feat(tushare): add Pro JSON client with daily and stock_basic"
```

---

### Task 9: `pkg/analysis` — port table builders + parity tests

**Files:**
- Create: `tools/dump_python_fixture.py`
- Create: `pkg/analysis/testdata/<TSCODE>.json` (generated)
- Create: `pkg/analysis/model.go`
- Create: `pkg/analysis/builder.go`
- Create: `pkg/analysis/format.go`
- Test: `pkg/analysis/builder_test.go`
- Test: `pkg/analysis/parity_test.go`
- Test: `pkg/analysis/format_test.go`

**Python reference:** `analysis.py:263-443`.

This is the linchpin task. Strategy:

1. Write a small Python helper that runs the legacy code and dumps a JSON fixture (window means, composite means, model table rows-as-strings, reference table rows-as-strings, header text).
2. Capture fixtures for at least 3 representative stocks: a high-priced (600519), mid-priced (000001), low-priced/ETF (159915).
3. Port the Go implementation to consume the same `[]spread.Bar` input and assert byte equality of rendered cells with the fixtures.

- [ ] **Step 1: Write the Python fixture dumper**

Create `tools/dump_python_fixture.py`:
```python
"""One-shot helper that dumps analysis fixtures for Go parity tests.

Run from project root:

    python tools/dump_python_fixture.py 600519 1620 1655 1601 1632 > pkg/analysis/testdata/600519.json

Args:
    code               6-digit stock code, present in the local SQLite db
    open               today's open (optional, pass 0 to skip header / plan)
    high low close     optional actual overrides
"""
import json
import sys
from datetime import datetime

from analysis import StockAnalyzer
from config import DB_PATH
from db import DailyDB


def main() -> None:
    if len(sys.argv) < 3:
        print("usage: dump_python_fixture.py CODE OPEN [HIGH LOW CLOSE]", file=sys.stderr)
        sys.exit(2)
    code = sys.argv[1]
    open_p = float(sys.argv[2])
    high = float(sys.argv[3]) if len(sys.argv) > 3 else None
    low = float(sys.argv[4]) if len(sys.argv) > 4 else None
    close = float(sys.argv[5]) if len(sys.argv) > 5 else None

    db = DailyDB(DB_PATH)
    db.init()
    rows = db.query_daily(code, "2000-01-01", datetime.now().strftime("%Y-%m-%d"))
    a = StockAnalyzer(
        code,
        all_rows=rows,
        open_price=open_p if open_p else None,
        actual_high=high, actual_low=low, actual_close=close,
    )
    window_means = a._compute_window_means()
    composite = a._compute_composite_means(window_means)

    fixture = {
        "code": code,
        "rows": rows,                                 # what Go must ingest
        "open_price": open_p,
        "actual_high": high,
        "actual_low": low,
        "actual_close": close,
        "window_means": {
            w: {k: window_means[w][k] for k in a.MODEL_SPREAD_KEYS}
            for w in a._WINDOW_NAMES
        },
        "composite_means": composite,
        "header_text": a._format_header(open_p, composite) if open_p else "",
        "model_table_text": a._build_spread_model_table(window_means, composite),
        "reference_table_text": a._build_reference_table(open_p, window_means, composite)
            if open_p else "",
    }
    json.dump(fixture, sys.stdout, ensure_ascii=False, indent=2, default=str)


if __name__ == "__main__":
    main()
```

- [ ] **Step 2: Generate fixtures**

For each test stock (assuming Python data exists locally), run:
```bash
mkdir -p pkg/analysis/testdata
python tools/dump_python_fixture.py 600519 1620 1655 1601 1632 > pkg/analysis/testdata/600519.json
python tools/dump_python_fixture.py 000001 11.50 11.80 11.40 11.55 > pkg/analysis/testdata/000001.json
python tools/dump_python_fixture.py 159915 2.20 2.25 2.18 2.22 > pkg/analysis/testdata/159915.json
```

If a stock has no rows locally, run `python stock.py fetch --stocks <code>` first.

Commit the fixtures (they are small, deterministic, and the foundation of parity testing).

- [ ] **Step 3: Write `pkg/analysis/model.go`**

```go
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
```

- [ ] **Step 4: Write `pkg/analysis/builder.go`**

```go
package analysis

import (
	"fmt"
	"strconv"

	"stock/pkg/shared/window"
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
	windows := window.Make(in.Rows)
	wmeans := window.Means(windows)
	comp := window.Composite(wmeans)

	result := AnalysisResult{
		TsCode:         in.TsCode,
		StockName:      in.StockName,
		Windows:        window.Names,
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

func buildModelTable(wmeans window.MeansResult, comp map[string]float64) ModelTable {
	headers := append([]string{"时段"}, ModelSpreadLabels...)
	rows := make([][]string, 0, len(window.Names)+1)
	for _, wname := range window.Names {
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

// buildReferenceTable mirrors analysis.py:336-422 exactly.
func buildReferenceTable(openPrice float64, actualHigh, actualLow, actualClose *float64, wm window.MeansResult, _ map[string]float64) ReferenceTable {
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

func highRow(openPrice float64, actualLow *float64, wm window.MeansResult) []string {
	row := []string{"最高价预测"}
	for _, wname := range window.Names {
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

func lowRow(openPrice float64, actualHigh *float64, wm window.MeansResult) []string {
	row := []string{"最低价预测"}
	for _, wname := range window.Names {
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

func closeRow(actualHigh, actualLow *float64, wm window.MeansResult) []string {
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

func lastClose(rows []interface{}) *float64 { return nil } // overridden below
```

NOTE: replace the placeholder `lastClose` with the real implementation:

```go
// Replace the placeholder lastClose in builder.go with:

func lastClose(rows []spread.Bar) *float64 {
	if len(rows) == 0 {
		return nil
	}
	// rows are unsorted; find the max TradeDate's close.
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

And add `import "stock/pkg/shared/spread"`.

- [ ] **Step 5: Write `pkg/analysis/format.go` (table rendering for the CLI)**

```go
package analysis

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// DisplayWidth returns the terminal cell width, treating CJK wide chars as 2.
func DisplayWidth(s string) int {
	w := 0
	for _, r := range s {
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isWide(r rune) bool {
	// Treat East Asian "W" or "F" as width 2. We approximate by Unicode block.
	if r >= 0x1100 && r <= 0x115F { // Hangul Jamo
		return true
	}
	if r >= 0x2E80 && r <= 0x303E { // CJK Radicals etc.
		return true
	}
	if r >= 0x3041 && r <= 0x33FF { // Hiragana/Katakana/CJK Sym & Punct
		return true
	}
	if r >= 0x3400 && r <= 0x4DBF { // CJK Ext A
		return true
	}
	if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified
		return true
	}
	if r >= 0xA000 && r <= 0xA4CF { // Yi
		return true
	}
	if r >= 0xAC00 && r <= 0xD7A3 { // Hangul Syllables
		return true
	}
	if r >= 0xF900 && r <= 0xFAFF { // CJK Compatibility
		return true
	}
	if r >= 0xFE30 && r <= 0xFE4F { // CJK Compatibility Forms
		return true
	}
	if r >= 0xFF00 && r <= 0xFF60 { // Fullwidth ASCII
		return true
	}
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return true
	}
	if unicode.Is(unicode.Han, r) {
		return true
	}
	if utf8.RuneLen(r) >= 3 { // catch-all conservative
		return true
	}
	return false
}

// Rpad right-pads `s` with spaces to the requested display width.
func Rpad(s string, width int) string {
	if d := width - DisplayWidth(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

// FormatTable renders a CJK-aware ASCII table.
func FormatTable(headers []string, rows [][]string) string {
	colW := make([]int, len(headers))
	for i, h := range headers {
		colW[i] = DisplayWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colW) && DisplayWidth(cell) > colW[i] {
				colW[i] = DisplayWidth(cell)
			}
		}
	}
	sep := "+"
	for _, w := range colW {
		sep += strings.Repeat("-", w+2) + "+"
	}
	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString("|")
	for i, h := range headers {
		b.WriteString(" " + Rpad(h, colW[i]) + " |")
	}
	b.WriteString("\n" + sep + "\n")
	for _, row := range rows {
		b.WriteString("|")
		for i, cell := range row {
			if i < len(colW) {
				b.WriteString(" " + Rpad(cell, colW[i]) + " |")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString(sep)
	return b.String()
}
```

- [ ] **Step 6: Write `pkg/analysis/format_test.go`**

```go
package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/analysis"
)

func TestDisplayWidth(t *testing.T) {
	assert.Equal(t, 5, analysis.DisplayWidth("hello"))
	assert.Equal(t, 4, analysis.DisplayWidth("开盘"))
	assert.Equal(t, 3, analysis.DisplayWidth("开A"))
}

func TestFormatTable_Shape(t *testing.T) {
	out := analysis.FormatTable([]string{"时段", "数值"}, [][]string{{"历史", "1.23"}, {"近1月", "0.45"}})
	assert.Contains(t, out, "时段")
	assert.Contains(t, out, "历史")
	assert.Contains(t, out, "+")
	lines := 0
	for _, c := range out {
		if c == '\n' {
			lines++
		}
	}
	assert.Equal(t, 5, lines, "6 lines means 5 newlines")
}
```

- [ ] **Step 7: Write unit tests in `pkg/analysis/builder_test.go`**

```go
package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/analysis"
	"stock/pkg/shared/spread"
)

func bar(date string, oh, ol, hl, hc, lc, oc float64) spread.Bar {
	return spread.Bar{
		TradeDate: date,
		Close:     100,
		Spreads:   spread.Spreads{OH: oh, OL: ol, HL: hl, HC: hc, LC: lc, OC: oc},
	}
}

func TestBuild_NoOpenPriceProducesNoReferenceTable(t *testing.T) {
	res := analysis.Build(analysis.Input{
		TsCode: "603778.SH",
		Rows:   []spread.Bar{bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5)},
	})
	assert.Empty(t, res.ReferenceTable.Headers)
	assert.NotEmpty(t, res.ModelTable.Headers)
}

func TestBuild_ModelTableShape(t *testing.T) {
	rows := []spread.Bar{
		bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5),
		bar("20240103", 2, 1.0, 3.0, 1.0, 2.0, 1.0),
	}
	res := analysis.Build(analysis.Input{TsCode: "X", Rows: rows})
	require.Equal(t, 5, len(res.ModelTable.Rows), "4 windows + composite")
	assert.Equal(t, "综合均值", res.ModelTable.Rows[4][0])
}
```

- [ ] **Step 8: Write the parity test in `pkg/analysis/parity_test.go`**

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
	"stock/pkg/shared/spread"
)

type pythonRow struct {
	TsCode    string  `json:"ts_code"`
	TradeDate string  `json:"trade_date"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Vol       float64 `json:"vol"`
	Amount    float64 `json:"amount"`
	SpreadOH  float64 `json:"spread_oh"`
	SpreadOL  float64 `json:"spread_ol"`
	SpreadHL  float64 `json:"spread_hl"`
	SpreadOC  float64 `json:"spread_oc"`
	SpreadHC  float64 `json:"spread_hc"`
	SpreadLC  float64 `json:"spread_lc"`
}

type pythonFixture struct {
	Code               string                       `json:"code"`
	Rows               []pythonRow                  `json:"rows"`
	OpenPrice          float64                      `json:"open_price"`
	ActualHigh         *float64                     `json:"actual_high"`
	ActualLow          *float64                     `json:"actual_low"`
	ActualClose        *float64                     `json:"actual_close"`
	WindowMeans        map[string]map[string]*float64 `json:"window_means"`
	CompositeMeans     map[string]float64           `json:"composite_means"`
	ModelTableText     string                       `json:"model_table_text"`
	ReferenceTableText string                       `json:"reference_table_text"`
}

func TestParity_AgainstPythonFixtures(t *testing.T) {
	matches, err := filepath.Glob("testdata/*.json")
	require.NoError(t, err)
	require.NotEmpty(t, matches, "no parity fixtures generated yet (see tools/dump_python_fixture.py)")

	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			var fx pythonFixture
			require.NoError(t, json.Unmarshal(raw, &fx))

			bars := make([]spread.Bar, 0, len(fx.Rows))
			for _, r := range fx.Rows {
				bars = append(bars, spread.Bar{
					TsCode: r.TsCode, TradeDate: r.TradeDate,
					Open: r.Open, High: r.High, Low: r.Low, Close: r.Close,
					Vol: r.Vol, Amount: r.Amount,
					Spreads: spread.Spreads{
						OH: r.SpreadOH, OL: r.SpreadOL, HL: r.SpreadHL,
						OC: r.SpreadOC, HC: r.SpreadHC, LC: r.SpreadLC,
					},
				})
			}
			in := analysis.Input{
				TsCode:      fx.Code,
				Rows:        bars,
				OpenPrice:   ptrFloatIf(fx.OpenPrice),
				ActualHigh:  fx.ActualHigh,
				ActualLow:   fx.ActualLow,
				ActualClose: fx.ActualClose,
			}
			res := analysis.Build(in)

			// Composite means must match to 4 decimal places.
			for _, key := range analysis.ModelSpreadKeys {
				assert.InDelta(t, fx.CompositeMeans[key], res.CompositeMeans[key], 1e-4, "composite[%s]", key)
			}

			// Window means: nil-equal and value-equal.
			for w, byKey := range fx.WindowMeans {
				for k, want := range byKey {
					got := res.WindowMeans[w][k]
					if want == nil {
						assert.Nil(t, got, "%s/%s should be nil", w, k)
						continue
					}
					require.NotNil(t, got, "%s/%s should not be nil", w, k)
					assert.InDelta(t, *want, *got, 1e-4)
				}
			}

			// Rendered model table must equal byte-for-byte.
			goModel := analysis.FormatTable(res.ModelTable.Headers, res.ModelTable.Rows)
			assert.Equal(t, fx.ModelTableText, goModel, "model_table rendering differs for %s", fx.Code)

			if fx.OpenPrice != 0 {
				goRef := analysis.FormatTable(res.ReferenceTable.Headers, res.ReferenceTable.Rows)
				assert.Equal(t, fx.ReferenceTableText, goRef, "reference_table rendering differs for %s", fx.Code)
			}
		})
	}
}

func ptrFloatIf(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}
```

- [ ] **Step 9: Run the full pkg/analysis test suite**

Run: `go test ./pkg/analysis/... -v`
Expected: all PASS. If a fixture mismatch occurs:
  - First, dump the diff: `diff <(echo "$EXPECTED") <(echo "$GOT")`.
  - Common culprits: ordering of `MODEL_SPREAD_KEYS`, padding width (CJK width function), `meanOfNumericCells` over-eagerly parsing "/", or the wrong row index for `近2周` lookups.
  - DO NOT relax the test by switching to InDelta on rendered strings — the parity contract is byte-for-byte. Fix the Go side.

- [ ] **Step 10: Run the full P1 suite**

Run: `make test`
Expected: all PASS across `pkg/shared/*`, `pkg/tushare`, `pkg/analysis`.

- [ ] **Step 11: Commit**

```bash
git add tools/dump_python_fixture.py pkg/analysis/
git commit -m "feat(analysis): port window/composite means and reference tables with Python parity"
```

---

## Exit criterion

- [ ] `go test ./pkg/...` green
- [ ] Parity test in `pkg/analysis/parity_test.go` validates against at least 3 distinct stock fixtures
- [ ] `pkg/analysis.Build()` returns `AnalysisResult` consumable by both an HTTP handler (P4) and a CLI renderer (P5)

## Hand-off

Next: [P2 — Server core](./2026-05-14-p2-server-core.md). P2 wires the GORM models, config loader, auth middleware, and first-run bootstrap.
