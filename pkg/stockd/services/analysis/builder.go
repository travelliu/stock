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

// BuildRefTable computes the prediction reference table from windows and optional actual prices.
func BuildRefTable(windows []*models.WindowData, openPrice, actualHigh, actualLow *float64) *models.RefTable {
	if len(windows) == 0 {
		return nil
	}
	lastWin := windows[len(windows)-1]

	// High: openPrice + spreadOH.mean per window; reverseLow = actualLow + spreadHL.mean
	var highVals []float64
	highWindows := make(map[string]float64)
	if openPrice != nil {
		for _, w := range windows {
			if w.Means != nil && w.Means.SpreadOH != nil && w.Means.SpreadOH.Mean != 0 {
				v := *openPrice + w.Means.SpreadOH.Mean
				highWindows[w.Info.Id] = v
				highVals = append(highVals, v)
			}
		}
	}
	var highReverseLow float64
	if actualLow != nil && lastWin.Means != nil && lastWin.Means.SpreadHL != nil && lastWin.Means.SpreadHL.Mean != 0 {
		highReverseLow = *actualLow + lastWin.Means.SpreadHL.Mean
		highVals = append(highVals, highReverseLow)
	}

	// Low: openPrice - spreadOL.mean per window; reverseHigh = actualHigh - spreadHL.mean
	var lowVals []float64
	lowWindows := make(map[string]float64)
	if openPrice != nil {
		for _, w := range windows {
			if w.Means != nil && w.Means.SpreadOL != nil && w.Means.SpreadOL.Mean != 0 {
				v := *openPrice - w.Means.SpreadOL.Mean
				lowWindows[w.Info.Id] = v
				lowVals = append(lowVals, v)
			}
		}
	}
	var lowReverseHigh float64
	if actualHigh != nil && lastWin.Means != nil && lastWin.Means.SpreadHL != nil && lastWin.Means.SpreadHL.Mean != 0 {
		lowReverseHigh = *actualHigh - lastWin.Means.SpreadHL.Mean
		lowVals = append(lowVals, lowReverseHigh)
	}

	// Close: reverseLow = actualLow + spreadLC.mean; reverseHigh = actualHigh - spreadHC.mean
	var closeVals []float64
	var closeReverseLow, closeReverseHigh float64
	if actualLow != nil && lastWin.Means != nil && lastWin.Means.SpreadLC != nil && lastWin.Means.SpreadLC.Mean != 0 {
		closeReverseLow = *actualLow + lastWin.Means.SpreadLC.Mean
		closeVals = append(closeVals, closeReverseLow)
	}
	if actualHigh != nil && lastWin.Means != nil && lastWin.Means.SpreadHC != nil && lastWin.Means.SpreadHC.Mean != 0 {
		closeReverseHigh = *actualHigh - lastWin.Means.SpreadHC.Mean
		closeVals = append(closeVals, closeReverseHigh)
	}

	return &models.RefTable{
		High: models.PredictRow{
			Windows:    highWindows,
			ReverseLow: highReverseLow,
			Mean:       avgOf(highVals),
			Direction:  "+",
		},
		Low: models.PredictRow{
			Windows:     lowWindows,
			ReverseHigh: lowReverseHigh,
			Mean:        avgOf(lowVals),
			Direction:   "-",
		},
		Close: models.PredictRow{
			Windows:     make(map[string]float64),
			ReverseLow:  closeReverseLow,
			ReverseHigh: closeReverseHigh,
			Mean:        avgOf(closeVals),
			Direction:   "-",
		},
	}
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

// Composite computes the arithmetic average across all time windows for each spread key.
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

// RecommendRange finds the narrowest contiguous range covering >= threshold% of sorted values.
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
