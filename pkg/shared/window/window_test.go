package window_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/shared/spread"
	"stock/pkg/shared/window"
)

func bar(date string, oh, ol, hl, oc, hc, lc float64) spread.Bar {
	return spread.Bar{
		TradeDate: date,
		Spreads:   spread.Spreads{OH: oh, OL: ol, HL: hl, OC: oc, HC: hc, LC: lc},
	}
}

func TestWindowNamesAndDaysSync(t *testing.T) {
	assert.Equal(t, []string{"历史", "近3月", "近1月", "近2周"}, window.Names)
	assert.Equal(t, 4, len(window.Days))
	assert.Nil(t, window.Days[0], "first window is unbounded (历史)")
	assert.Equal(t, 90, *window.Days[1])
	assert.Equal(t, 30, *window.Days[2])
	assert.Equal(t, 15, *window.Days[3])
}

func TestMakeWindows_SlicesByDate(t *testing.T) {
	rows := []spread.Bar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	got := window.Make(rows)
	require.Len(t, got, 4)
	assert.Equal(t, "历史", got[0].Name)
	assert.Len(t, got[0].Rows, 3)
	assert.Equal(t, "近3月", got[1].Name)
	assert.Len(t, got[1].Rows, 3) // fewer than 90 rows, so all included
	assert.Equal(t, "近1月", got[2].Name)
	assert.Len(t, got[2].Rows, 3)
	assert.Equal(t, "近2周", got[3].Name)
	assert.Len(t, got[3].Rows, 3)
}

func TestWindowMeans_Basic(t *testing.T) {
	rows := []spread.Bar{
		bar("2024-01-04", 1, 0.5, 1.5, 0.5, 0.5, 1.0),
		bar("2024-01-03", 2, 1.0, 3.0, 1.0, 1.0, 2.0),
		bar("2024-01-02", 3, 1.5, 4.5, 1.5, 1.5, 3.0),
	}
	means := window.Means(window.Make(rows))
	assert.InDelta(t, 2.0, *means["历史"]["spread_oh"], 1e-9)
	assert.InDelta(t, 1.0, *means["历史"]["spread_ol"], 1e-9)
}

func TestWindowMeans_Empty(t *testing.T) {
	means := window.Means(window.Make(nil))
	for _, name := range window.Names {
		for _, key := range window.SpreadKeys {
			assert.Nil(t, means[name][key], "%s/%s should be nil", name, key)
		}
	}
}

func TestCompositeMeans_NoneTreatedAsZero(t *testing.T) {
	// Python: composite = mean of non-None window means; all-None -> 0.0
	m := func(v float64) *float64 { return &v }
	wm := window.MeansResult{
		"历史":  {"spread_oh": m(4.0), "spread_ol": nil},
		"近3月": {"spread_oh": m(2.0), "spread_ol": nil},
		"近1月": {"spread_oh": m(1.0), "spread_ol": nil},
		"近2周": {"spread_oh": m(0.5), "spread_ol": nil},
	}
	comp := window.Composite(wm)
	assert.InDelta(t, 1.875, comp["spread_oh"], 1e-9)
	assert.Equal(t, 0.0, comp["spread_ol"], "all-None composite must collapse to 0.0")
}

func TestRecommendedRange_Empty(t *testing.T) {
	r := window.RecommendedRange(nil, 60.0)
	assert.Nil(t, r)
}

func TestRecommendedRange_Single(t *testing.T) {
	r := window.RecommendedRange([]float64{3.0}, 60.0)
	require.NotNil(t, r)
	assert.InDelta(t, 3.0, r.Low, 1e-9)
	assert.InDelta(t, 3.0, r.High, 1e-9)
	assert.InDelta(t, 100.0, r.CumPct, 1e-9)
}

func TestRecommendedRange_Sliding(t *testing.T) {
	vals := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := window.RecommendedRange(vals, 30.0)
	require.NotNil(t, r)
	assert.InDelta(t, 2.0, r.High-r.Low, 1e-9, "tightest contiguous span of 3 values is 2.0")
}

func TestRecommendedRange_SkewedTight(t *testing.T) {
	vals := []float64{0.1, 0.2, 0.15, 0.25, 0.3, 0.35, 0.18, 0.22, 1.0, 2.0}
	r := window.RecommendedRange(vals, 60.0)
	require.NotNil(t, r)
	assert.True(t, r.CumPct >= 60.0)
	assert.True(t, r.High-r.Low < 1.0)
}

func TestDistribution_Basic(t *testing.T) {
	bins := window.Distribution([]float64{1, 2, 3, 4, 5}, 5)
	require.Len(t, bins, 5)
	total := 0
	for _, b := range bins {
		total += b.Count
	}
	assert.Equal(t, 5, total)
}

func TestDistribution_Empty(t *testing.T) {
	assert.Empty(t, window.Distribution(nil, 10))
}

func TestDistribution_Single(t *testing.T) {
	bins := window.Distribution([]float64{3.0}, 10)
	require.Len(t, bins, 1)
	assert.Equal(t, 1, bins[0].Count)
	assert.InDelta(t, 100.0, bins[0].Pct, 1e-9)
}
