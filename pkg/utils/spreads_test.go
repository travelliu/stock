package utils_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"stock/pkg/utils"
)

func TestComputeSpreads(t *testing.T) {
	got := utils.ComputeSpreads(100.0, 105.0, 98.0, 102.0)
	assert.InDelta(t, 5.0, got.OH, 1e-3, "spread_oh")
	assert.InDelta(t, 2.0, got.OL, 1e-3, "spread_ol")
	assert.InDelta(t, 7.0, got.HL, 1e-3, "spread_hl")
	assert.InDelta(t, 2.0, got.OC, 1e-3, "spread_oc")
	assert.InDelta(t, 3.0, got.HC, 1e-3, "spread_hc")
	assert.InDelta(t, 4.0, got.LC, 1e-3, "spread_lc")
}

func TestComputeSpreads_AllAbsolute(t *testing.T) {
	got := utils.ComputeSpreads(100.0, 100.5, 95.0, 96.0)
	assert.True(t, got.OC >= 0, "spread_oc must be absolute, got %v", got.OC)
	assert.InDelta(t, 4.0, got.OC, 1e-3)
}

func TestComputeSpreads_Zero(t *testing.T) {
	got := utils.ComputeSpreads(50.0, 50.0, 50.0, 50.0)
	assert.Equal(t, 0.0, math.Abs(got.OH+got.OL+got.HL+got.OC+got.HC+got.LC))
}
