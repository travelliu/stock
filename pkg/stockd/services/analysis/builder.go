package analysis

import (
	"fmt"
	"sort"
	"strconv"

	"stock/pkg/models"
)

// Build runs the full analysis pipeline and returns computed data.
func Build(in models.Input) *models.AnalysisResult {
	windows := Make(in.Rows)
	Means(windows)
	comp := Composite(windows)
	ref := BuildRefTable(windows, in.OpenPrice, in.ActualHigh, in.ActualLow)

	return &models.AnalysisResult{
		TsCode:         in.TsCode,
		StockName:      in.StockName,
		Windows:        windows,
		CompositeMeans: comp,
		RefTable:       ref,
		OpenPrice:      in.OpenPrice,
		ActualHigh:     in.ActualHigh,
		ActualLow:      in.ActualLow,
		ActualClose:    in.ActualClose,
	}
}

// BuildRefTable populates WindowData.Predict for each window and returns cross-window summaries.
// Each window gets a full PredictBreakdown (ByMean/ByMedian/ByEWMA/ByRatio/ReverseLow/ReverseHigh/Mean).
// RefTable aggregates per-window Means across all windows.
func BuildRefTable(windows []*models.WindowData, openPrice, actualHigh, actualLow *float64) *models.RefTable {
	if len(windows) == 0 {
		return nil
	}

	var highMeans, lowMeans, closeMeans []float64
	for _, w := range windows {
		if w.Means == nil {
			continue
		}
		p := &models.WindowPredict{
			High:  buildHighBreakdown(w.Means, openPrice, actualLow),
			Low:   buildLowBreakdown(w.Means, openPrice, actualHigh),
			Close: buildCloseBreakdown(w.Means, openPrice, actualHigh, actualLow),
		}
		w.Predict = p
		if p.High.Mean != 0 {
			highMeans = append(highMeans, p.High.Mean)
		}
		if p.Low.Mean != 0 {
			lowMeans = append(lowMeans, p.Low.Mean)
		}
		if p.Close.Mean != 0 {
			closeMeans = append(closeMeans, p.Close.Mean)
		}
	}

	return &models.RefTable{
		High:  models.PredictRow{Mean: avgOf(highMeans)},
		Low:   models.PredictRow{Mean: avgOf(lowMeans)},
		Close: models.PredictRow{Mean: avgOf(closeMeans)},
	}
}

// buildHighBreakdown computes all High prediction methods for one window.
// spread_oh = high - open  →  high = open + spread_oh
// ReverseLow: from actualLow + spread_hl (since high ≈ low + spread_hl)
func buildHighBreakdown(m *models.MeansData, openPrice, actualLow *float64) models.PredictBreakdown {
	var b models.PredictBreakdown
	if openPrice != nil {
		if oh := m.SpreadOH; oh != nil {
			b.ByMean = *openPrice + oh.Mean
			b.ByMedian = *openPrice + oh.Median
			if oh.EWMA > 0 {
				b.ByEWMA = *openPrice + oh.EWMA
			}
			if oh.EWMARatio > 0 {
				b.ByRatio = *openPrice + *openPrice*oh.EWMARatio
			}
		}
	}
	if actualLow != nil {
		if hl := m.SpreadHL; hl != nil && hl.Mean != 0 {
			b.ReverseLow = *actualLow + hl.Mean
		}
	}
	b.Mean = avgOfNonZero(b.ByMean, b.ByMedian, b.ByEWMA, b.ByRatio, b.ReverseLow)
	return b
}

// buildLowBreakdown computes all Low prediction methods for one window.
// spread_ol = open - low  →  low = open - spread_ol
// ReverseHigh: from actualHigh - spread_hl (since low ≈ high - spread_hl)
func buildLowBreakdown(m *models.MeansData, openPrice, actualHigh *float64) models.PredictBreakdown {
	var b models.PredictBreakdown
	if openPrice != nil {
		if ol := m.SpreadOL; ol != nil {
			b.ByMean = *openPrice - ol.Mean
			b.ByMedian = *openPrice - ol.Median
			if ol.EWMA > 0 {
				b.ByEWMA = *openPrice - ol.EWMA
			}
			if ol.EWMARatio > 0 {
				b.ByRatio = *openPrice - *openPrice*ol.EWMARatio
			}
		}
	}
	if actualHigh != nil {
		if hl := m.SpreadHL; hl != nil && hl.Mean != 0 {
			b.ReverseHigh = *actualHigh - hl.Mean
		}
	}
	b.Mean = avgOfNonZero(b.ByMean, b.ByMedian, b.ByEWMA, b.ByRatio, b.ReverseHigh)
	return b
}

// buildCloseBreakdown computes all Close prediction methods for one window.
// spread_oc = open - close  →  close = open - spread_oc
// ReverseLow: from actualLow + spread_lc (since close ≈ low + spread_lc)
// ReverseHigh: from actualHigh - spread_hc (since close ≈ high - spread_hc)
func buildCloseBreakdown(m *models.MeansData, openPrice, actualHigh, actualLow *float64) models.PredictBreakdown {
	var b models.PredictBreakdown
	if openPrice != nil {
		if oc := m.SpreadOC; oc != nil {
			b.ByMean = *openPrice - oc.Mean
			b.ByMedian = *openPrice - oc.Median
			if oc.EWMA != 0 {
				b.ByEWMA = *openPrice - oc.EWMA
			}
			if oc.EWMARatio != 0 {
				b.ByRatio = *openPrice - *openPrice*oc.EWMARatio
			}
		}
	}
	if actualLow != nil {
		if lc := m.SpreadLC; lc != nil && lc.Mean != 0 {
			b.ReverseLow = *actualLow + lc.Mean
		}
	}
	if actualHigh != nil {
		if hc := m.SpreadHC; hc != nil && hc.Mean != 0 {
			b.ReverseHigh = *actualHigh - hc.Mean
		}
	}
	b.Mean = avgOfNonZero(b.ByMean, b.ByMedian, b.ByEWMA, b.ByRatio, b.ReverseLow, b.ReverseHigh)
	return b
}

// avgOfNonZero returns the arithmetic mean of non-zero values.
func avgOfNonZero(vals ...float64) float64 {
	var sum float64
	var n int
	for _, v := range vals {
		if v != 0 {
			sum += v
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func avgOf(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// Composite computes the arithmetic average of MeansAvgData.Mean across all time windows
// for each of the six spread keys. The result is a single blended value that treats all
// windows equally regardless of sample size — useful as a quick model summary but less
// reactive to recent changes than the per-window values in WindowData.Means.
func Composite(windows []*models.WindowData) map[string]float64 {
	spreadFields := []struct {
		key string
		get func(*models.MeansData) *models.MeansAvgData
	}{
		{"spread_oh", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadOH }},
		{"spread_ol", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadOL }},
		{"spread_hl", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadHL }},
		{"spread_hc", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadHC }},
		{"spread_lc", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadLC }},
		{"spread_oc", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadOC }},
	}

	composite := make(map[string]float64)
	for _, sf := range spreadFields {
		var vals []float64
		for _, w := range windows {
			if w.Means == nil {
				continue
			}
			m := sf.get(w.Means)
			if m != nil && m.Mean != 0 {
				vals = append(vals, m.Mean)
			}
		}
		if len(vals) > 0 {
			var sum float64
			for _, v := range vals {
				sum += v
			}
			composite[sf.key] = sum / float64(len(vals))
		}
	}
	return composite
}

// MeanOfNumericCells computes the average of parseable float strings.
func MeanOfNumericCells(cells []string) string {
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

// RecommendRange finds the narrowest contiguous band that covers at least threshold% of
// the sorted observations using a sliding window of size ceil(n * threshold / 100).
// The result is the tightest historical price-spread range a trader can expect to see
// on the specified fraction of days — used for 高抛低吸 target recommendation.
func RecommendRange(sorted []float64, threshold float64) (low, high, cumPct float64, ok bool) {
	n := len(sorted)
	if n == 0 {
		return 0, 0, 0, false
	}
	if n == 1 {
		return sorted[0], sorted[0], 100.0, true
	}

	needed := int(roundFloat(float64(n)*threshold/100, 0))
	if needed < 1 {
		needed = 1
	}

	bestLow := sorted[0]
	bestHigh := sorted[n-1]
	bestSpan := bestHigh - bestLow

	for i := 0; i <= n-needed; i++ {
		span := sorted[i+needed-1] - sorted[i]
		if span < bestSpan {
			bestSpan = span
			bestLow = sorted[i]
			bestHigh = sorted[i+needed-1]
		}
	}

	cumPct = roundFloat(float64(needed)/float64(n)*100, 1)
	return bestLow, bestHigh, cumPct, true
}

func roundFloat(v float64, places int) float64 {
	p := 1.0
	for i := 0; i < places; i++ {
		p *= 10
	}
	return float64(int(v*p+0.5)) / p
}

// ExtractSpreadValues returns non-zero values for a given spread key from bars.
func ExtractSpreadValues(rows []*models.DailyBar, key string) []float64 {
	var vals []float64
	for _, r := range rows {
		v := GetSpreadField(r.Spreads, key)
		if v != 0 {
			vals = append(vals, v)
		}
	}
	return vals
}

func GetSpreadField(s models.Spreads, key string) float64 {
	switch key {
	case "spread_oh":
		return s.OH
	case "spread_ol":
		return s.OL
	case "spread_hl":
		return s.HL
	case "spread_hc":
		return s.HC
	case "spread_lc":
		return s.LC
	case "spread_oc":
		return s.OC
	default:
		return 0
	}
}

// FormatStats returns [count, avg, median, mean] as formatted strings.
func FormatStats(vals []float64) []string {
	if len(vals) == 0 {
		return []string{"0", "-", "-", "-"}
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	avg := sum / float64(len(vals))

	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	var median float64
	if len(sorted)%2 == 1 {
		median = sorted[len(sorted)/2]
	} else {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	}
	mean := (avg + median) / 2

	return []string{
		strconv.Itoa(len(vals)),
		fmt.Sprintf("%.2f", avg),
		fmt.Sprintf("%.2f", median),
		fmt.Sprintf("%.2f", mean),
	}
}
