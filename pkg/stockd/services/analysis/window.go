package analysis

import (
	"math"
	"sort"
	"stock/pkg/utils"

	"github.com/montanaflynn/stats"

	"stock/pkg/models"
)

var Windows = []*models.WindowInfo{
	{
		Id:   "All",
		Name: "全部",
		Day:  0,
	},
	{
		Id:   "last_90",
		Name: "近3月",
		Day:  90,
	},
	{
		Id:   "last_30",
		Name: "近3月",
		Day:  30,
	},
	{
		Id:   "last_15",
		Name: "近2周",
		Day:  15,
	},
}

func Make(rows []*models.DailyBar) []*models.WindowData {
	sorted := make([]*models.DailyBar, len(rows))
	copy(sorted, rows)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].TradeDate > sorted[j].TradeDate })
	var out []*models.WindowData
	for _, info := range Windows {
		if info.Day == 0 {
			out = append(out, &models.WindowData{
				Info: info,
				Rows: sorted,
			})
			continue
		}
		end := info.Day
		if end > len(sorted) {
			end = len(sorted)
		}
		out = append(out, &models.WindowData{
			Info: info,
			Rows: sorted[:end],
		})
	}
	Means(out)
	return out
}

func Means(windows []*models.WindowData) {
	for _, w := range windows {
		w.Means = &models.MeansData{}
		var (
			spreadOhList []float64
			spreadOlList []float64
			spreadHlList []float64
			spreadOcList []float64
			spreadHcList []float64
			spreadLcList []float64
			openList     []float64 // newest-first, mirrors spreadOhList/spreadOlList order
		)

		for _, row := range w.Rows {
			spreadOhList = append(spreadOhList, row.Spreads.OH)
			spreadOlList = append(spreadOlList, row.Spreads.OL)
			spreadHlList = append(spreadHlList, row.Spreads.HL)
			spreadOcList = append(spreadOcList, row.Spreads.OC)
			spreadHcList = append(spreadHcList, row.Spreads.HC)
			spreadLcList = append(spreadLcList, row.Spreads.LC)
			openList = append(openList, row.Open)
		}
		w.Means.SpreadOH = Cloc(spreadOhList)
		w.Means.SpreadOL = Cloc(spreadOlList)
		w.Means.SpreadHL = Cloc(spreadHlList)
		w.Means.SpreadOC = Cloc(spreadOcList)
		w.Means.SpreadHC = Cloc(spreadHcList)
		w.Means.SpreadLC = Cloc(spreadLcList)

		// ratio fields (spread / open) — only meaningful for OH and OL
		w.Means.SpreadOH.AvgRatio, w.Means.SpreadOH.EWMARatio = computeRatios(spreadOhList, openList)
		w.Means.SpreadOL.AvgRatio, w.Means.SpreadOL.EWMARatio = computeRatios(spreadOlList, openList)

		threshold := 30.0
		if w.Info.Id == "All" {
			threshold = 60.0
		}
		setRecommend(w.Means.SpreadOH, spreadOhList, threshold)
		setRecommend(w.Means.SpreadOL, spreadOlList, threshold)
	}
}

// computeRatios computes avg and EWMA of (spread/open) ratios.
// Both slices must be in newest-first order and have the same length.
func computeRatios(spreads, opens []float64) (avgRatio, ewmaRatio float64) {
	const lambda = 0.9
	var sum, ewmaNum, ewmaDen float64
	n := 0
	for i, s := range spreads {
		if i >= len(opens) || opens[i] <= 0 {
			continue
		}
		ratio := s / opens[i]
		sum += ratio
		w := math.Pow(lambda, float64(i))
		ewmaNum += w * ratio
		ewmaDen += w
		n++
	}
	if n > 0 {
		avgRatio = utils.Round(sum / float64(n))
	}
	if ewmaDen > 0 {
		ewmaRatio = utils.Round(ewmaNum / ewmaDen)
	}
	return
}

func setRecommend(m *models.MeansAvgData, vals []float64, threshold float64) {
	if m == nil || len(vals) == 0 {
		return
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	if lo, hi, pct, ok := RecommendRange(sorted, threshold); ok {
		m.Recommend = &models.RecommendRangeResult{Low: lo, High: hi, CumPct: pct}
	}
}

// Cloc computes descriptive statistics from list, which must be in newest-first order.
func Cloc(list []float64) *models.MeansAvgData {
	v := &models.MeansAvgData{Count: len(list)}
	if len(list) == 0 {
		return v
	}

	// EWMA (λ=0.9): computed before sorting to preserve newest-first order.
	const lambda = 0.9
	var ewmaNum, ewmaDen float64
	for i, val := range list {
		w := math.Pow(lambda, float64(i))
		ewmaNum += w * val
		ewmaDen += w
	}
	if ewmaDen > 0 {
		v.EWMA = utils.Round(ewmaNum / ewmaDen)
	}

	sorted := make([]float64, len(list))
	copy(sorted, list)
	sort.Float64s(sorted)

	v.Avg, _ = stats.Mean(sorted)
	v.Avg = utils.Round(v.Avg)
	v.Median, _ = stats.Median(sorted)
	v.Median = utils.Round(v.Median)
	v.Mean, _ = stats.Mean([]float64{v.Avg, v.Median})
	v.Mean = utils.Round(v.Mean)
	stddev, _ := stats.StandardDeviation(stats.Float64Data(sorted))
	v.StdDev = utils.Round(stddev)
	v.Distribution = Distribution(sorted, 10)
	return v
}
