package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/models"
)

func bar(date string, oh, ol, hl, hc, lc, oc float64) *models.DailyBar {
	return &models.DailyBar{
		TradeDate: date,
		Spreads:   models.Spreads{OH: oh, OL: ol, HL: hl, HC: hc, LC: lc, OC: oc},
	}
}

func TestBuild_WindowsAndComposite(t *testing.T) {
	in := models.Input{
		TsCode:    "600519.SH",
		StockName: "贵州茅台",
		Rows:      []*models.DailyBar{bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5)},
	}
	res := Build(in)
	assert.Equal(t, "600519.SH", res.TsCode)
	assert.Equal(t, "贵州茅台", res.StockName)
	assert.NotEmpty(t, res.Windows)
	assert.NotNil(t, res.CompositeMeans)
}

func TestBuild_WithOpenPrice(t *testing.T) {
	open := 100.0
	rows := []*models.DailyBar{
		bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5),
		bar("20240103", 2, 1.0, 3.0, 1.0, 2.0, 1.0),
	}
	in := models.Input{TsCode: "X", Rows: rows, OpenPrice: &open}
	res := Build(in)
	assert.NotNil(t, res.OpenPrice)
	assert.Equal(t, 100.0, *res.OpenPrice)
	assert.NotEmpty(t, res.Windows)
}

func TestComposite(t *testing.T) {
	windows := []*models.WindowData{
		{
			Info: &models.WindowInfo{Id: "All"},
			Means: &models.MeansData{
				SpreadOH: &models.MeansAvgData{Mean: 1.0},
				SpreadOL: &models.MeansAvgData{Mean: 0.5},
			},
		},
		{
			Info: &models.WindowInfo{Id: "last_90"},
			Means: &models.MeansData{
				SpreadOH: &models.MeansAvgData{Mean: 2.0},
				SpreadOL: &models.MeansAvgData{Mean: 1.0},
			},
		},
	}
	comp := Composite(windows)
	assert.InDelta(t, 1.5, comp["spread_oh"], 0.001)
	assert.InDelta(t, 0.75, comp["spread_ol"], 0.001)
}

func TestRecommendRange(t *testing.T) {
	sorted := []float64{1.0, 1.5, 2.0, 2.5, 3.0}
	lo, hi, pct, ok := RecommendRange(sorted, 60)
	assert.True(t, ok)
	assert.LessOrEqual(t, hi-lo, 2.0)
	assert.GreaterOrEqual(t, pct, 60.0)
}

func TestExtractSpreadValues(t *testing.T) {
	rows := []*models.DailyBar{
		bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5),
		bar("20240103", 2, 1.0, 3.0, 1.0, 2.0, 1.0),
	}
	vals := ExtractSpreadValues(rows, "spread_oh")
	assert.Equal(t, []float64{1, 2}, vals)
}

func TestFormatStats(t *testing.T) {
	vals := []float64{1.0, 3.0}
	stats := FormatStats(vals)
	assert.Equal(t, "2", stats[0])
	assert.Equal(t, "2.00", stats[1])
	assert.Equal(t, "2.00", stats[2])
	assert.Equal(t, "2.00", stats[3])
}

func TestFormatStats_Empty(t *testing.T) {
	stats := FormatStats(nil)
	assert.Equal(t, []string{"0", "-", "-", "-"}, stats)
}

func TestMeanOfNumericCells(t *testing.T) {
	result := MeanOfNumericCells([]string{"1.50", "2.50", "abc"})
	assert.Equal(t, "2.00", result)
}

func TestMeanOfNumericCells_Empty(t *testing.T) {
	result := MeanOfNumericCells(nil)
	assert.Equal(t, "/", result)
}
