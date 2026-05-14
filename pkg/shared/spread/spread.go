// Package spread computes the six absolute price spreads used by the analysis pipeline.
package spread

import "math"

// OHLC is the input for spread computation.
type OHLC struct {
	Open, High, Low, Close float64
}

// Spreads holds the six absolute spreads.
type Spreads struct {
	OH float64 // |high - open|
	OL float64 // |open - low|
	HL float64 // |high - low|
	OC float64 // |open - close|
	HC float64 // |high - close|
	LC float64 // |low  - close|
}

// Compute returns the six absolute spreads for one bar.
func Compute(b OHLC) Spreads {
	return Spreads{
		OH: math.Abs(b.High - b.Open),
		OL: math.Abs(b.Open - b.Low),
		HL: math.Abs(b.High - b.Low),
		OC: math.Abs(b.Open - b.Close),
		HC: math.Abs(b.High - b.Close),
		LC: math.Abs(b.Low - b.Close),
	}
}

// Bar combines OHLCV with trade date and computed spreads. Used by downstream
// packages (analysis, db) as the canonical row representation.
type Bar struct {
	TsCode    string
	TradeDate string // YYYYMMDD (Tushare style); DB layer may also accept YYYY-MM-DD
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Vol       float64
	Amount    float64
	Spreads   Spreads
}
