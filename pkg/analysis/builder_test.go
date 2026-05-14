package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/analysis"
	"stock/pkg/shared/spread"
)

func bar(date string, oh, ol, hl, hc, lc, oc float64) spread.Bar {
	return spread.Bar{
		TradeDate: date,
		Close:     100,
		Spreads:   spread.Spreads{OH: oh, OL: ol, HL: hl, HC: hc, LC: lc, OC: oc},
	}
}

func TestBuild_NoOpenPriceProducesNoReferenceTable(t *testing.T) {
	res := analysis.Build(analysis.Input{
		TsCode: "603778.SH",
		Rows:   []spread.Bar{bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5)},
	})
	assert.Empty(t, res.ReferenceTable.Headers)
	assert.NotEmpty(t, res.ModelTable.Headers)
}

func TestBuild_ModelTableShape(t *testing.T) {
	rows := []spread.Bar{
		bar("20240104", 1, 0.5, 1.5, 0.5, 1.0, 0.5),
		bar("20240103", 2, 1.0, 3.0, 1.0, 2.0, 1.0),
	}
	res := analysis.Build(analysis.Input{TsCode: "X", Rows: rows})
	require.Equal(t, 5, len(res.ModelTable.Rows), "4 windows + composite")
	assert.Equal(t, "综合均值", res.ModelTable.Rows[4][0])
}
