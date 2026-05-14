package spread_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/shared/spread"
)

func TestCompute(t *testing.T) {
	in := spread.OHLC{Open: 100.0, High: 105.0, Low: 98.0, Close: 102.0}
	got := spread.Compute(in)

	assert.InDelta(t, 5.0, got.OH, 1e-9, "spread_oh")
	assert.InDelta(t, 2.0, got.OL, 1e-9, "spread_ol")
	assert.InDelta(t, 7.0, got.HL, 1e-9, "spread_hl")
	assert.InDelta(t, 2.0, got.OC, 1e-9, "spread_oc") // |100-102|
	assert.InDelta(t, 3.0, got.HC, 1e-9, "spread_hc") // |105-102|
	assert.InDelta(t, 4.0, got.LC, 1e-9, "spread_lc") // |98-102|
}

func TestCompute_AllAbsolute(t *testing.T) {
	// Down day: close < open, but spread_oc must be positive (absolute).
	in := spread.OHLC{Open: 100.0, High: 100.5, Low: 95.0, Close: 96.0}
	got := spread.Compute(in)
	assert.True(t, got.OC >= 0, "spread_oc must be absolute, got %v", got.OC)
	assert.InDelta(t, 4.0, got.OC, 1e-9)
}

func TestCompute_Zero(t *testing.T) {
	in := spread.OHLC{Open: 50.0, High: 50.0, Low: 50.0, Close: 50.0}
	got := spread.Compute(in)
	assert.Equal(t, 0.0, math.Abs(got.OH+got.OL+got.HL+got.OC+got.HC+got.LC))
}
