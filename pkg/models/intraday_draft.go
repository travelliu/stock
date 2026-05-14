package models

import "time"

type IntradayDraft struct {
	ID        uint      `gorm:"primaryKey" json:"id,omitempty"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_code_date;not null" json:"userID,omitempty"`
	TsCode    string    `gorm:"uniqueIndex:idx_user_code_date;size:16;not null" json:"tsCode,omitempty"`
	TradeDate string    `gorm:"uniqueIndex:idx_user_code_date;size:8;not null" json:"tradeDate,omitempty"`
	Open      *float64  `json:"open,omitempty"`
	High      *float64  `json:"high,omitempty"`
	Low       *float64  `json:"low,omitempty"`
	Close     *float64  `json:"close,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}
