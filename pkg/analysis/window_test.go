package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/analysis"
	"stock/pkg/models"
)

func bar(date string, oh, ol, hl, oc, hc, lc float64) models.DailyBar {
	return models.DailyBar{
		TradeDate: date,
		Spreads:   models.Spreads{OH: oh, OL: ol, HL: hl, OC: oc, HC: hc, LC: lc},
	}
}

func TestWindowNamesAndDaysSync(t *testing.T) {
	assert.Equal(t, []string{"历史", "近3月", "近1月", "近2周"}, analysis.Names)
	assert.Equal(t, 4, len(analysis.Days))
	assert.Nil(t, analysis.Days[0], "first window is unbounded (历史)")
	assert.Equal(t, 90, *analysis.Days[1])
	assert.Equal(t, 30, *analysis.Days[2])
	assert.Equal(t, 15, *analysis.Days[3])
}

func TestMakeWindows_SlicesByDate(t *testing.T) {
	rows := []models.DailyBar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	got := analysis.Make(rows)
	require.Len(t, got, 4)
	assert.Equal(t, "历史", got[0].Name)
	assert.Len(t, got[0].Rows, 3)
	assert.Equal(t, "近3月", got[1].Name)
	assert.Len(t, got[1].Rows, 3)
	assert.Equal(t, "近1月", got[2].Name)
	assert.Len(t, got[2].Rows, 3)
	assert.Equal(t, "近2周", got[3].Name)
	assert.Len(t, got[3].Rows, 3)
}

func TestWindowMeans_Basic(t *testing.T) {
	rows := []models.DailyBar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	means := analysis.Means(analysis.Make(rows))
	assert.InDelta(t, 2.0, *means["历史"]["spread_oh"], 1e-9)
	assert.InDelta(t, 1.0, *means["历史"]["spread_ol"], 1e-9)
}

func TestWindowMeans_Empty(t *testing.T) {
	means := analysis.Means(analysis.Make(nil))
	for _, name := range analysis.Names {
		for _, key := range analysis.SpreadKeys {
			assert.Nil(t, means[name][key], "%s/%s should be nil", name, key)
		}
	}
}

func TestCompositeMeans_NoneTreatedAsZero(t *testing.T) {
	m := func(v float64) *float64 { return &v }
	wm := models.MeansResult{
		"历史":  {"spread_oh": m(4.0), "spread_ol": nil},
		"近3月": {"spread_oh": m(2.0), "spread_ol": nil},
		"近1月": {"spread_oh": m(1.0), "spread_ol": nil},
		"近2周": {"spread_oh": m(0.5), "spread_ol": nil},
	}
	comp := analysis.Composite(wm)
	assert.InDelta(t, 1.875, comp["spread_oh"], 1e-9)
	assert.Equal(t, 0.0, comp["spread_ol"], "all-None composite must collapse to 0.0")
}
