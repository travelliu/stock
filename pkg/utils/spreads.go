package utils

import (
	"math"

	"stock/pkg/models"
)

func ComputeSpreads(open, high, low, close float64) models.Spreads {
	return models.Spreads{
		OH: math.Abs(high - open),
		OL: math.Abs(open - low),
		HL: math.Abs(high - low),
		OC: math.Abs(open - close),
		HC: math.Abs(high - close),
		LC: math.Abs(low - close),
	}
}
