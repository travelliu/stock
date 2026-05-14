package models

import "time"

type Portfolio struct {
	ID      uint      `gorm:"primaryKey" json:"id,omitempty"`
	UserID  uint      `gorm:"uniqueIndex:idx_user_code;not null" json:"userID,omitempty"`
	TsCode  string    `gorm:"uniqueIndex:idx_user_code;size:16;not null" json:"tsCode,omitempty"`
	Note    string    `gorm:"size:255" json:"note,omitempty"`
	AddedAt time.Time `json:"addedAt,omitempty"`
}

type PortfolioReq struct {
	TsCode string `json:"tsCode"`
	Note   string `json:"note,omitempty"`
}
