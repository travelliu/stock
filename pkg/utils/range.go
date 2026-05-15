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
	cum := RoundTo(float64(needed)/float64(n)*100.0, 1)
	return &Range{Low: bestLow, High: bestHigh, CumPct: cum}
}
