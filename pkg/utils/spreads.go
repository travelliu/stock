package utils

import (
	"math"

	"stock/pkg/models"
)

func ComputeSpreads(open, high, low, close float64) models.Spreads {
	return models.Spreads{
		OH: math.Round(math.Abs(high-open)*1000) / 1000,
		OL: math.Round(math.Abs(open-low)*1000) / 1000,
		HL: math.Round(math.Abs(high-low)*1000) / 1000,
		OC: math.Round(math.Abs(open-close)*1000) / 1000,
		HC: math.Round(math.Abs(high-close)*1000) / 1000,
		LC: math.Round(math.Abs(low-close)*1000) / 1000,
	}
}

func Round(float65 float64) float64 {
	return math.Round(float65*1000) / 1000
}
