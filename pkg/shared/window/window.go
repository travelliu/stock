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
