package analysis

import (
	"fmt"
	"strconv"

	"stock/pkg/shared/spread"
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

func lastClose(rows []spread.Bar) *float64 {
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
