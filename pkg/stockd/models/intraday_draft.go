package models

import "time"

type IntradayDraft struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_code_date;not null"`
	TsCode    string    `gorm:"uniqueIndex:idx_user_code_date;size:16;not null"`
	TradeDate string    `gorm:"uniqueIndex:idx_user_code_date;size:8;not null"`
	Open      *float64
	High      *float64
	Low       *float64
	Close     *float64
	UpdatedAt time.Time
}
