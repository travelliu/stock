package render

import (
	"fmt"
	"strings"

	"stock/pkg/models"
)

// i18n keys for table headers. CLI maps these to Chinese for display.
const (
	KeyTimePeriod     = "time_period"
	KeyHistory        = "history"
	KeyLast90         = "last_90"
	KeyLast30         = "last_30"
	KeyLast15         = "last_15"
	KeyComposite      = "composite"
	KeySpreadOH       = "spread_oh"
	KeySpreadOL       = "spread_ol"
	KeySpreadHL       = "spread_hl"
	KeySpreadHC       = "spread_hc"
	KeySpreadLC       = "spread_lc"
	KeySpreadOC       = "spread_oc"
	KeyPredictHigh    = "predict_high"
	KeyPredictLow     = "predict_low"
	KeyPredictClose   = "predict_close"
	KeyReverseLow     = "reverse_low"
	KeyReverseHigh    = "reverse_high"
	KeyMean           = "mean"
	KeyDirection      = "direction"
	KeySampleCount    = "sample_count"
	KeyAvg            = "avg"
	KeyMedian         = "median"
	KeyRecommendOH    = "recommend_oh"
	KeyRecommendOL    = "recommend_ol"
	KeyDate           = "date"
	KeyPredictHighVal = "predict_high_val"
	KeyActualHighVal  = "actual_high_val"
	KeyDevHigh        = "dev_high"
	KeyPredictLowVal  = "predict_low_val"
	KeyActualLowVal   = "actual_low_val"
	KeyDevLow         = "dev_low"
	KeyByMedian       = "by_median"
	KeyByEWMA         = "by_ewma"
	KeyByRatio        = "by_ratio"
	KeyEWMA           = "ewma"
)

var zh = map[string]string{
	KeyTimePeriod:     "时段",
	KeyHistory:        "历史",
	KeyLast90:         "近3月",
	KeyLast30:         "近1月",
	KeyLast15:         "近2周",
	KeyComposite:      "综合均值",
	KeySpreadOH:       "开盘与最高价",
	KeySpreadOL:       "开盘与最低价",
	KeySpreadHL:       "最高与最低价",
	KeySpreadHC:       "最高与收盘价",
	KeySpreadLC:       "最低与收盘价",
	KeySpreadOC:       "开盘与收盘价",
	KeyPredictHigh:    "最高价预测",
	KeyPredictLow:     "最低价预测",
	KeyPredictClose:   "收盘价预测",
	KeyReverseLow:     "最低价反推(当日最低价)",
	KeyReverseHigh:    "最高价反推(当日最高价)",
	KeyMean:           "均值",
	KeyDirection:      "正负算一",
	KeySampleCount:    "样本数",
	KeyAvg:            "平均值",
	KeyMedian:         "中位数",
	KeyRecommendOH:    "高抛差价(高-开盘)",
	KeyRecommendOL:    "低吸差价(开盘-低)",
	KeyDate:           "日期",
	KeyPredictHighVal: "预测高",
	KeyActualHighVal:  "实际高",
	KeyDevHigh:        "偏差",
	KeyPredictLowVal:  "预测低",
	KeyActualLowVal:   "实际低",
	KeyDevLow:         "偏差",
	KeyByMedian:       "中位数预测",
	KeyByEWMA:         "EWMA预测",
	KeyByRatio:        "比率预测",
	KeyEWMA:           "EWMA",
}

func t(key string) string {
	if v, ok := zh[key]; ok {
		return v
	}
	return key
}

var spreadKeys = []string{
	"spread_oh", "spread_ol", "spread_hl",
	"spread_hc", "spread_lc", "spread_oc",
}

var spreadLabels = []string{
	KeySpreadOH, KeySpreadOL, KeySpreadHL,
	KeySpreadHC, KeySpreadLC, KeySpreadOC,
}

var windowIdToKey = map[string]string{
	"All":     KeyHistory,
	"last_90": KeyLast90,
	"last_30": KeyLast30,
	"last_15": KeyLast15,
}

func windowKeyName(id string) string {
	if k, ok := windowIdToKey[id]; ok {
		return t(k)
	}
	return id
}

func AnalysisTable(r models.StockAnalysisResult) {

	fmt.Printf("\n=== %s (%s) 交易计划 ===\n\n", r.StockName, r.TsCode)

	// Price line
	parts := []string{}
	if r.OpenPrice != nil {
		parts = append(parts, fmt.Sprintf("开盘价: %.2f", *r.OpenPrice))
	}
	if r.ActualHigh != nil {
		parts = append(parts, fmt.Sprintf("最高价: %.2f", *r.ActualHigh))
	} else {
		parts = append(parts, "最高价: ")
	}
	if r.ActualLow != nil {
		parts = append(parts, fmt.Sprintf("最低价: %.2f", *r.ActualLow))
	} else {
		parts = append(parts, "最低价: ")
	}
	if r.ActualClose != nil {
		parts = append(parts, fmt.Sprintf("收盘价: %.2f", *r.ActualClose))
	} else {
		parts = append(parts, "收盘价: ")
	}
	fmt.Println(strings.Join(parts, "   "))

	fmt.Println()
	if r.RefTable != nil {
		fmt.Println("── 预测收盘价(历史参考价) ──")
		refTable(r)
	}

	// Model table
	fmt.Println("\n── 价差模型 ──")
	headers := []string{t(KeyTimePeriod)}
	for _, lbl := range spreadLabels {
		headers = append(headers, t(lbl))
	}
	var rows [][]string
	for _, win := range r.Windows {
		row := []string{windowKeyName(win.Info.Id)}
		if win.Means != nil {
			row = append(row, formatMeans(win.Means)...)
		} else {
			for range spreadKeys {
				row = append(row, "-")
			}
		}
		rows = append(rows, row)
	}
	compRow := []string{t(KeyComposite)}
	for _, key := range spreadKeys {
		compRow = append(compRow, fmt.Sprintf("%.2f", r.CompositeMeans[key]))
	}
	rows = append(rows, compRow)
	fmt.Println(FormatTable(headers, rows))

	// Analysis table
	fmt.Println("\n=== 价差分析 ===")
	analysisTable(r)

	// Distribution tables
	distributionTables(r)
}

func formatMeans(md *models.MeansData) []string {
	fields := []*models.MeansAvgData{
		md.SpreadOH, md.SpreadOL, md.SpreadHL,
		md.SpreadHC, md.SpreadLC, md.SpreadOC,
	}
	var row []string
	for _, m := range fields {
		if m != nil {
			row = append(row, fmt.Sprintf("%.2f", m.Mean))
		} else {
			row = append(row, "-")
		}
	}
	return row
}

func refTable(r models.StockAnalysisResult) {
	if r.RefTable == nil || len(r.Windows) == 0 {
		return
	}

	fv := func(v float64) string {
		if v == 0 {
			return "/"
		}
		return fmt.Sprintf("%.2f", v)
	}

	// Rows = prediction methods; columns = windows grouped by 最高价/最低价/收盘价.
	methods := []struct {
		label string
		get   func(*models.PredictBreakdown) float64
	}{
		{"均值", func(b *models.PredictBreakdown) float64 { return b.ByMean }},
		{t(KeyByMedian), func(b *models.PredictBreakdown) float64 { return b.ByMedian }},
		{t(KeyByEWMA), func(b *models.PredictBreakdown) float64 { return b.ByEWMA }},
		{t(KeyByRatio), func(b *models.PredictBreakdown) float64 { return b.ByRatio }},
		{t(KeyReverseLow), func(b *models.PredictBreakdown) float64 { return b.ReverseLow }},
		{t(KeyReverseHigh), func(b *models.PredictBreakdown) float64 { return b.ReverseHigh }},
		{"综合均值", func(b *models.PredictBreakdown) float64 { return b.Mean }},
	}

	picks := []func(*models.WindowPredict) *models.PredictBreakdown{
		func(p *models.WindowPredict) *models.PredictBreakdown { return &p.High },
		func(p *models.WindowPredict) *models.PredictBreakdown { return &p.Low },
		func(p *models.WindowPredict) *models.PredictBreakdown { return &p.Close },
	}

	nWin := len(r.Windows)

	// Build flat header row: method | win1..winN (high) | win1..winN (low) | win1..winN (close)
	headers := []string{"方法"}
	for range picks {
		for _, win := range r.Windows {
			headers = append(headers, windowKeyName(win.Info.Id))
		}
	}

	// Build data rows.
	var rows [][]string
	for _, m := range methods {
		row := []string{m.label}
		for _, pick := range picks {
			for _, win := range r.Windows {
				if win.Predict != nil {
					row = append(row, fv(m.get(pick(win.Predict))))
				} else {
					row = append(row, "/")
				}
			}
		}
		rows = append(rows, row)
	}

	// Compute column widths for sub-header.
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

	sectionW := func(start, end int) int {
		w := 0
		for i := start; i <= end; i++ {
			w += colW[i] + 3
		}
		return w
	}

	methodSW := colW[0] + 2
	subLine := "|" + strings.Repeat(" ", methodSW) + "|" +
		Rpad("── "+t(KeyPredictHigh)+" ──", sectionW(1, nWin)) + "|" +
		Rpad("── "+t(KeyPredictLow)+" ──", sectionW(nWin+1, 2*nWin)) + "|" +
		Rpad("── "+t(KeyPredictClose)+" ──", sectionW(2*nWin+1, 3*nWin)) + "|"

	fmt.Println()
	fmt.Println(subLine)
	fmt.Println(FormatTable(headers, rows))

	// Cross-window summary: one value per price type.
	crossHeaders := []string{"跨窗口综合均值", t(KeyPredictHigh), t(KeyPredictLow), t(KeyPredictClose)}
	crossRows := [][]string{{"", fv(r.RefTable.High.Mean), fv(r.RefTable.Low.Mean), fv(r.RefTable.Close.Mean)}}
	fmt.Println(FormatTable(crossHeaders, crossRows))
}

func analysisTable(r models.StockAnalysisResult) {
	uHeaders := []string{
		t(KeyTimePeriod), t(KeySampleCount), t(KeyAvg), t(KeyMedian), t(KeyMean), t(KeyEWMA), "",
		t(KeySampleCount), t(KeyAvg), t(KeyMedian), t(KeyMean), t(KeyEWMA), "",
		t(KeyRecommendOH), t(KeyRecommendOL),
	}

	ordered := make([]*models.WindowData, len(r.Windows))
	copy(ordered, r.Windows)
	for i, j := 0, len(ordered)-1; i < j; i, j = i+1, j-1 {
		ordered[i], ordered[j] = ordered[j], ordered[i]
	}

	var uTable [][]string
	for _, win := range ordered {
		row := []string{windowKeyName(win.Info.Id)}

		if win.Means != nil && win.Means.SpreadOH != nil {
			m := win.Means.SpreadOH
			row = append(row, fmt.Sprintf("%d", m.Count), fmt.Sprintf("%.2f", m.Avg), fmt.Sprintf("%.2f", m.Median), fmt.Sprintf("%.2f", m.Mean), fmt.Sprintf("%.2f", m.EWMA))
		} else {
			row = append(row, "0", "-", "-", "-", "-")
		}
		row = append(row, "")

		if win.Means != nil && win.Means.SpreadOL != nil {
			m := win.Means.SpreadOL
			row = append(row, fmt.Sprintf("%d", m.Count), fmt.Sprintf("%.2f", m.Avg), fmt.Sprintf("%.2f", m.Median), fmt.Sprintf("%.2f", m.Mean), fmt.Sprintf("%.2f", m.EWMA))
		} else {
			row = append(row, "0", "-", "-", "-", "-")
		}
		row = append(row, "")

		row = append(row, formatRecommend(win.Means, func(md *models.MeansData) *models.MeansAvgData { return md.SpreadOH }))
		row = append(row, formatRecommend(win.Means, func(md *models.MeansData) *models.MeansAvgData { return md.SpreadOL }))

		uTable = append(uTable, row)
	}

	// Compute column widths for sub-header
	colW := make([]int, len(uHeaders))
	for i, h := range uHeaders {
		colW[i] = DisplayWidth(h)
	}
	for _, row := range uTable {
		for i, cell := range row {
			if i < len(colW) && DisplayWidth(cell) > colW[i] {
				colW[i] = DisplayWidth(cell)
			}
		}
	}

	sectionW := func(start, end int) int {
		w := 0
		for i := start; i <= end; i++ {
			w += colW[i] + 3
		}
		return w
	}

	// columns: [时段] [count avg median mean ewma ""] [count avg median mean ewma ""] [recOH recOL]
	// indices:  0      1     2   3      4    5    6    7     8   9      10   11   12   13    14
	timeSW := colW[0] + 2
	ohSW := sectionW(1, 6)
	olSW := sectionW(7, 12)
	recSW := sectionW(13, 14)

	subLine := "|" + strings.Repeat(" ", timeSW) + "|" +
		Rpad("── 最高-开盘 ──", ohSW) + "|" +
		Rpad("── 开盘-最低 ──", olSW) + "|" +
		Rpad("── 高抛低吸推荐 (累计占比) ──", recSW) + "|"

	fmt.Println()
	fmt.Println(subLine)
	fmt.Println(FormatTable(uHeaders, uTable))
}

func formatRecommend(means *models.MeansData, get func(*models.MeansData) *models.MeansAvgData) string {
	if means == nil {
		return "-"
	}
	m := get(means)
	if m == nil || m.Recommend == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f~%.2f (%.1f%%)", m.Recommend.Low, m.Recommend.High, m.Recommend.CumPct)
}

func distributionTables(r models.StockAnalysisResult) {
	keys := []struct {
		label string
		get   func(*models.MeansData) *models.MeansAvgData
	}{
		{"最高-开盘", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadOH }},
		{"开盘-最低", func(m *models.MeansData) *models.MeansAvgData { return m.SpreadOL }},
	}

	distHeaders := []string{"区间", "数量", "占比"}

	for _, k := range keys {
		var tables []string
		for _, win := range r.Windows {
			if win.Means == nil {
				continue
			}
			m := k.get(win.Means)
			if m == nil || m.Count == 0 || len(m.Distribution) == 0 {
				continue
			}
			var rows [][]string
			for _, b := range m.Distribution {
				rows = append(rows, []string{
					fmt.Sprintf("%.2f~%.2f", b.Lower, b.Upper),
					fmt.Sprintf("%d", b.Count),
					fmt.Sprintf("%.1f%%", b.Pct),
				})
			}
			block := fmt.Sprintf("── %s 分布 (%s,%d条) ──\n", k.label, windowKeyName(win.Info.Id), m.Count) +
				FormatTable(distHeaders, rows)
			tables = append(tables, block)
		}
		if len(tables) > 0 {
			fmt.Println()
			fmt.Println(joinSideBySide(tables, 4))
		}
	}
}

func joinSideBySide(tables []string, gap int) string {
	split := make([][]string, len(tables))
	for i, tbl := range tables {
		split[i] = strings.Split(tbl, "\n")
	}

	// Normalize each block to same width
	normalized := make([][]string, len(split))
	for i, block := range split {
		maxW := 0
		for _, line := range block {
			if w := DisplayWidth(line); w > maxW {
				maxW = w
			}
		}
		padded := make([]string, len(block))
		for j, line := range block {
			padded[j] = Rpad(line, maxW)
		}
		normalized[i] = padded
	}

	maxLines := 0
	for _, b := range normalized {
		if len(b) > maxLines {
			maxLines = len(b)
		}
	}

	pad := strings.Repeat(" ", gap)
	var lines []string
	for i := 0; i < maxLines; i++ {
		var parts []string
		for _, b := range normalized {
			if i < len(b) {
				parts = append(parts, b[i])
			} else {
				parts = append(parts, strings.Repeat(" ", DisplayWidth(b[0])))
			}
		}
		lines = append(lines, strings.Join(parts, pad))
	}
	return strings.Join(lines, "\n")
}

func PortfolioTable(items []*models.StockPortfolio) {
	headers := []string{"代码", "名称", "时间", "开盘", "最高", "最低", "最新", "涨跌幅", "备注"}
	var rows [][]string
	for _, p := range items {
		q := p.Quote
		tm, open, high, low, price, changePct := "--", "--", "--", "--", "--", "--"
		if q != nil {
			tm = fmtQuoteTime(q.QuoteTime)
			open = fmt.Sprintf("%.2f", q.Open)
			high = fmt.Sprintf("%.2f", q.High)
			low = fmt.Sprintf("%.2f", q.Low)
			price = fmt.Sprintf("%.2f", q.Price)
			sign := "+"
			if q.ChangePct < 0 {
				sign = ""
			}
			changePct = fmt.Sprintf("%s%.2f%%", sign, q.ChangePct)
		}
		rows = append(rows, []string{p.Code, p.Name, tm, open, high, low, price, changePct, p.Note})
	}
	fmt.Println(FormatTable(headers, rows))
}

func fmtQuoteTime(t string) string {
	if len(t) == 14 {
		return t[8:10] + ":" + t[10:12]
	}
	if len(t) >= 16 && strings.Contains(t, " ") {
		return t[11:16]
	}
	if len(t) >= 5 {
		return t[:5]
	}
	return t
}

func BarsTable(items []*models.StockDailyBar) {
	headers := []string{"日期", "开盘", "最高", "最低", "收盘", "成交量"}
	var rows [][]string
	for _, b := range items {
		rows = append(rows, []string{
			b.TradeDate,
			fmt.Sprintf("%.2f", b.Open),
			fmt.Sprintf("%.2f", b.High),
			fmt.Sprintf("%.2f", b.Low),
			fmt.Sprintf("%.2f", b.Close),
			fmt.Sprintf("%.0f", b.Vol),
		})
	}
	fmt.Println(FormatTable(headers, rows))
}

// PredictionsTable renders prediction records.
func PredictionsTable(tsCode, stockName string, preds []models.StockAnalysisPrediction) {
	fmt.Printf("\n%s (%s) 预测记录\n\n", stockName, tsCode)

	headers := []string{
		t(KeyDate),
		t(KeyPredictHighVal), t(KeyActualHighVal), t(KeyDevHigh),
		t(KeyPredictLowVal), t(KeyActualLowVal), t(KeyDevLow),
	}
	var rows [][]string
	for _, p := range preds {
		rows = append(rows, []string{
			p.TradeDate,
			fmt.Sprintf("%.2f", p.PredictHigh),
			fmt.Sprintf("%.2f", p.ActualHigh),
			fmt.Sprintf("%+.2f", p.ActualHigh-p.PredictHigh),
			fmt.Sprintf("%.2f", p.PredictLow),
			fmt.Sprintf("%.2f", p.ActualLow),
			fmt.Sprintf("%+.2f", p.ActualLow-p.PredictLow),
		})
	}
	fmt.Println(FormatTable(headers, rows))
}
