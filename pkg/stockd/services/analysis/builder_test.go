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

func TestBuild_ModelTable(t *testing.T) {
	in := models.Input{
		TsCode:    "600519.SH",
		StockName: "贵州茅台",
		Rows:      []*models.DailyBar{bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5)},
	}
	res := Build(in)
	assert.Equal(t, "600519.SH", res.TsCode)
	assert.Equal(t, "贵州茅台", res.StockName)
	assert.NotEmpty(t, res.ModelTable.Rows)
}

func TestBuild_ReferenceTable_WithOpenPrice(t *testing.T) {
	open := 100.0
	rows := []*models.DailyBar{
		bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5),
		bar("20240103", 2, 1.0, 3.0, 1.0, 2.0, 1.0),
	}
	in := models.Input{TsCode: "X", Rows: rows, OpenPrice: &open}
	res := Build(in)
	assert.NotEmpty(t, res.ReferenceTable.Rows)
}
