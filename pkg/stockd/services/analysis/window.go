package analysis

import (
	"sort"
	"stock/pkg/utils"

	"github.com/montanaflynn/stats"

	"stock/pkg/models"
	// "gonum.org/v1/gonum/stat"
	// "gonum.org/v1/gonum/stat"
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

var Names = []string{"历史", "近3月", "近1月", "近2周"}

var Days = []*int{nil, ptr(90), ptr(30), ptr(15)}

func ptr(v int) *int { return &v }

var SpreadKeys = []string{
	"spread_oh", "spread_ol", "spread_hl",
	"spread_hc", "spread_lc", "spread_oc",
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
		)

		for _, row := range w.Rows {
			spreadOhList = append(spreadOhList, row.Spreads.OH)
			spreadOlList = append(spreadOlList, row.Spreads.OL)
			spreadHlList = append(spreadHlList, row.Spreads.HL)
			spreadOcList = append(spreadOcList, row.Spreads.OC)
			spreadHcList = append(spreadHcList, row.Spreads.HC)
			spreadLcList = append(spreadLcList, row.Spreads.LC)
		}
		w.Means.SpreadOH = Cloc(spreadOhList)
		w.Means.SpreadOL = Cloc(spreadOlList)
		w.Means.SpreadHL = Cloc(spreadHlList)
		w.Means.SpreadOC = Cloc(spreadOcList)
		w.Means.SpreadHC = Cloc(spreadHcList)
		w.Means.SpreadLC = Cloc(spreadLcList)

		threshold := 30.0
		if w.Info.Id == "All" {
			threshold = 60.0
		}
		setRecommend(w.Means.SpreadOH, spreadOhList, threshold)
		setRecommend(w.Means.SpreadOL, spreadOlList, threshold)
	}
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

func Cloc(list []float64) *models.MeansAvgData {
	sorted := make([]float64, len(list))
	copy(sorted, list)
	sort.Float64s(sorted)
	v := &models.MeansAvgData{Count: len(list)}

	v.Avg, _ = stats.Mean(sorted)
	v.Avg = utils.Round(v.Avg)
	v.Median, _ = stats.Median(sorted)
	v.Median = utils.Round(v.Median)
	v.Mean, _ = stats.Mean([]float64{v.Avg, v.Median})
	v.Mean = utils.Round(v.Mean)
	v.Distribution = Distribution(sorted, 10)
	return v
}
