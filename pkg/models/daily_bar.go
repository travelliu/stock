package models

type DailyBar struct {
	TsCode    string  `gorm:"primaryKey;size:16"  json:"tsCode,omitempty"`
	TradeDate string  `gorm:"primaryKey;size:8" json:"tradeDate,omitempty"`
	Open      float64 `json:"open" json:"open,omitempty"`
	High      float64 `json:"high,omitempty"`
	Low       float64 `json:"low,omitempty"`
	Close     float64 `json:"close,omitempty"`
	Vol       float64 `json:"vol,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Spreads   Spreads `gorm:"embedded;embeddedPrefix:spread_" json:"spreads"`
}
