package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/utils"
)

func TestDistribution_Basic(t *testing.T) {
	bins := utils.Distribution([]float64{1, 2, 3, 4, 5}, 5)
	require.Len(t, bins, 5)
	total := 0
	for _, b := range bins {
		total += b.Count
	}
	assert.Equal(t, 5, total)
}

func TestDistribution_Empty(t *testing.T) {
	assert.Empty(t, utils.Distribution(nil, 10))
}

func TestDistribution_Single(t *testing.T) {
	bins := utils.Distribution([]float64{3.0}, 10)
	require.Len(t, bins, 1)
	assert.Equal(t, 1, bins[0].Count)
	assert.InDelta(t, 100.0, bins[0].Pct, 1e-9)
}
