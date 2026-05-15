package analysis

import (
	"encoding/json"
	"fmt"
	"strconv"

	"stock/pkg/models"
)

/*

── 价差模型 ──
+----------+--------------+--------------+--------------+--------------+--------------+--------------+
| 时段     | 开盘与最高价 | 开盘与最低价 | 最高与最低价 | 最高与收盘价 | 最低与收盘价 | 开盘与收盘价 |
+----------+--------------+--------------+--------------+--------------+--------------+--------------+
| 历史     | 1.39         | 1.07         | 2.46         | 1.20         | 1.26         | 1.25         |
| 近3月    | 8.41         | 6.12         | 14.53        | 7.41         | 7.13         | 7.61         |
| 近1月    | 11.35        | 6.36         | 17.71        | 8.26         | 9.45         | 10.16        |
| 近2周    | 12.17        | 8.61         | 20.78        | 10.36        | 10.41        | 12.51        |
| 综合均值  | 8.33         | 5.54         | 13.87        | 6.81         | 7.06         | 7.88         |
+----------+--------------+--------------+--------------+--------------+--------------+--------------+

── 预测收盘价(历史参考价) ──
+------------+------------+-------------+-------------+-------------+------------------------+------------------------+--------+----------+
|            | 历史参考价 | 近3月参考价 | 近1月参考价 | 近2周参考价 | 最低价反推(当日最低价) | 最高价反推(当日最高价) | 均值   | 正负算一 |
+------------+------------+-------------+-------------+-------------+------------------------+------------------------+--------+----------+
| 最高价预测 | 356.22     | 363.24      | 366.18      | 367.00      | /                      | /                      | 363.16 | +        |
| 最低价预测 | 353.76     | 348.71      | 348.47      | 346.22      | /                      | 334.05                 | 349.29 | -        |
| 收盘价预测 | /          | /           | /           | /           | /                      | 344.47                 | 344.47 | -        |
+------------+------------+-------------+-------------+-------------+------------------------+------------------------+--------+----------+

=== 价差分析 ===

|       |── 最高-开盘 ──                       |── 开盘-最低 ──                      |── 高抛低吸推荐 (累计占比) ──            |
+-------+--------+--------+--------+-------+--+--------+--------+--------+------+--+-------------------+--------------------+
| 时段  | 样本数 | 平均值 | 中位数 | 均值  |  | 样本数 | 平均值 | 中位数 | 均值 |  | 高抛差价(高-开盘) | 低吸差价(开盘-低)  |
+-------+--------+--------+--------+-------+--+--------+--------+--------+------+--+-------------------+--------------------+
| 近2周 | 15     | 12.17  | 8.48   | 10.32 |  | 15     | 8.61   | 6.94   | 7.77 |  | 6.92~8.88 (26.7%) | 8.99~10.72 (26.7%) |
| 近1月 | 30     | 11.35  | 8.42   | 9.89  |  | 30     | 6.36   | 4.12   | 5.24 |  | 1.89~5.38 (30.0%) | 2.00~3.93 (30.0%)  |
| 近3月 | 90     | 8.41   | 6.05   | 7.23  |  | 90     | 6.12   | 4.61   | 5.37 |  | 0.56~3.29 (30.0%) | 2.00~4.13 (30.0%)  |
| 历史  | 2421   | 1.39   | 0.42   | 0.90  |  | 2421   | 1.07   | 0.38   | 0.73 |  | 0.00~0.22 (30.0%) | 0.00~0.19 (30.0%)  |
+-------+--------+--------+--------+-------+--+--------+--------+--------+------+--+-------------------+--------------------+
*/

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
func Build(in models.Input) *models.AnalysisResult {

	windows := Make(in.Rows)
	// Means(windows)
	// comp := Composite(wmeans)
	b, _ := json.Marshal(windows)
	fmt.Println(string(b))
	result := &models.AnalysisResult{
		TsCode:      in.TsCode,
		StockName:   in.StockName,
		Windows:     Windows,
		OpenPrice:   in.OpenPrice,
		ActualHigh:  in.ActualHigh,
		ActualLow:   in.ActualLow,
		ActualClose: in.ActualClose,
		// WindowMeans: wmeans,
		// ModelTable:     buildModelTable(wmeans, comp),
	}
	// if in.OpenPrice != nil {
	// 	result.ReferenceTable = buildReferenceTable(*in.OpenPrice, in.ActualHigh, in.ActualLow, in.ActualClose, wmeans, comp)
	// }
	return result
}

// func buildModelTable(wmeans models.MeansResult, comp map[string]float64) models.ModelTable {
// 	headers := append([]string{"时段"}, ModelSpreadLabels...)
// 	rows := make([][]string, 0, len(Names)+1)
// 	// for _, wname := range Names {
// 	// 	row := []string{wname}
// 	// 	for _, key := range ModelSpreadKeys {
// 	// 		row = append(row, formatPtr(wmeans[wname][key]))
// 	// 	}
// 	// 	rows = append(rows, row)
// 	// }
// 	// comprow := []string{"综合均值"}
// 	// for _, key := range ModelSpreadKeys {
// 	// 	comprow = append(comprow, fmt.Sprintf("%.2f", comp[key]))
// 	// }
// 	// rows = append(rows, comprow)
// 	return models.ModelTable{Headers: headers, Rows: rows}
// }

// func buildReferenceTable(openPrice float64, actualHigh, actualLow, actualClose *float64, wm models.MeansResult, _ map[string]float64) models.ReferenceTable {
// 	headers := []string{
// 		"", "历史参考价", "近3月参考价", "近1月参考价", "近2周参考价",
// 		"最低价反推(当日最低价)", "最高价反推(当日最高价)", "均值", "正负算一",
// 	}
// 	rows := [][]string{
// 		// 	highRow(openPrice, actualLow, wm),
// 		// 	lowRow(openPrice, actualHigh, wm),
// 		// 	closeRow(actualHigh, actualLow, wm),
// 	}
// 	return models.ReferenceTable{Headers: headers, Rows: rows}
// }

// func highRow(openPrice float64, actualLow *float64, wm models.MeansResult) []string {
// 	row := []string{"最高价预测"}
// 	for _, wname := range Names {
// 		v := wm[wname]["spread_oh"]
// 		if v == nil {
// 			row = append(row, "/")
// 			continue
// 		}
// 		row = append(row, fmt.Sprintf("%.2f", openPrice+*v))
// 	}
// 	hl2w := wm["近2周"]["spread_hl"]
// 	if actualLow != nil && hl2w != nil {
// 		row = append(row, fmt.Sprintf("%.2f", *actualLow+*hl2w))
// 	} else {
// 		row = append(row, "/")
// 	}
// 	row = append(row, "/")
// 	row = append(row, meanOfNumericCells(row[1:5]))
// 	row = append(row, "+")
// 	return row
// }

// func lowRow(openPrice float64, actualHigh *float64, wm models.MeansResult) []string {
// 	row := []string{"最低价预测"}
// 	for _, wname := range Names {
// 		v := wm[wname]["spread_ol"]
// 		if v == nil {
// 			row = append(row, "/")
// 			continue
// 		}
// 		row = append(row, fmt.Sprintf("%.2f", openPrice-*v))
// 	}
// 	row = append(row, "/")
// 	hl2w := wm["近2周"]["spread_hl"]
// 	if actualHigh != nil && hl2w != nil {
// 		row = append(row, fmt.Sprintf("%.2f", *actualHigh-*hl2w))
// 	} else {
// 		row = append(row, "/")
// 	}
// 	row = append(row, meanOfNumericCells(row[1:5]))
// 	row = append(row, "-")
// 	return row
// }

// func closeRow(actualHigh, actualLow *float64, wm models.MeansResult) []string {
// 	row := []string{"收盘价预测", "/", "/", "/", "/"}
// 	lc2w := wm["近2周"]["spread_lc"]
// 	if actualLow != nil && lc2w != nil {
// 		row = append(row, fmt.Sprintf("%.2f", *actualLow+*lc2w))
// 	} else {
// 		row = append(row, "/")
// 	}
// 	hc2w := wm["近2周"]["spread_hc"]
// 	if actualHigh != nil && hc2w != nil {
// 		row = append(row, fmt.Sprintf("%.2f", *actualHigh-*hc2w))
// 	} else {
// 		row = append(row, "/")
// 	}
// 	row = append(row, meanOfNumericCells(row[5:7]))
// 	row = append(row, "-")
// 	return row
// }

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
