package models

import "time"

type Portfolio struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	UserID  uint      `gorm:"uniqueIndex:idx_user_code;not null" json:"userId"`
	TsCode  string    `gorm:"uniqueIndex:idx_user_code;size:16;not null" json:"tsCode,omitempty"`
	Code    string    `gorm:"-" json:"code"`
	Name    string    `json:"name"`
	Note    string    `gorm:"size:255" json:"note,omitempty"`
	AddedAt time.Time `json:"addedAt"`
}

// PortfolioReq is the body for add/update operations.
// Code is the short numeric code (e.g. "300476"), not the full ts_code.
type PortfolioReq struct {
	Code   string `json:"code"`
	Note   string `json:"note,omitempty"`
	TsCode string `json:"tsCode,omitempty"`
}

func (p *PortfolioReq) GetCode() string {
	if p.Code != "" {
		return p.Code
	}
	if p.TsCode != "" {
		return p.TsCode
	}
	return ""
}
