package analysis

import (
	"math"
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
