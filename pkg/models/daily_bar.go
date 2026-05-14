package models

type DailyBar struct {
	TsCode    string  `gorm:"primaryKey;size:16"`
	TradeDate string  `gorm:"primaryKey;size:8"`
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Vol       float64
	Amount    float64
	Spreads   Spreads `gorm:"embedded;embeddedPrefix:spread_"`
}
