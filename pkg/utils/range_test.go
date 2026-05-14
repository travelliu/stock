package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/utils"
)

func TestRecommendedRange_Empty(t *testing.T) {
	r := utils.RecommendedRange(nil, 60.0)
	assert.Nil(t, r)
}

func TestRecommendedRange_Single(t *testing.T) {
	r := utils.RecommendedRange([]float64{3.0}, 60.0)
	require.NotNil(t, r)
	assert.InDelta(t, 3.0, r.Low, 1e-9)
	assert.InDelta(t, 3.0, r.High, 1e-9)
	assert.InDelta(t, 100.0, r.CumPct, 1e-9)
}

func TestRecommendedRange_Sliding(t *testing.T) {
	vals := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := utils.RecommendedRange(vals, 30.0)
	require.NotNil(t, r)
	assert.InDelta(t, 2.0, r.High-r.Low, 1e-9, "tightest contiguous span of 3 values is 2.0")
}

func TestRecommendedRange_SkewedTight(t *testing.T) {
	vals := []float64{0.1, 0.2, 0.15, 0.25, 0.3, 0.35, 0.18, 0.22, 1.0, 2.0}
	r := utils.RecommendedRange(vals, 60.0)
	require.NotNil(t, r)
	assert.True(t, r.CumPct >= 60.0)
	assert.True(t, r.High-r.Low < 1.0)
}
