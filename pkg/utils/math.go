package utils

import "math"

func RoundTo(v float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(v*pow) / pow
}

func Clamp(v, lo, hi float64) float64 {
	return math.Min(hi, math.Max(lo, v))
}
