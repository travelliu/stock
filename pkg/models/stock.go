package models

import "time"

type Stock struct {
	TsCode    string    `gorm:"primaryKey;size:16" json:"tsCode,omitempty"`
	Code      string    `gorm:"index;size:8;not null" json:"code,omitempty"`
	Name      string    `gorm:"size:32;not null" json:"name,omitempty"`
	Area      string    `gorm:"size:16" json:"area,omitempty"`
	Industry  string    `gorm:"size:32" json:"industry,omitempty"`
	Market    string    `gorm:"size:16" json:"market,omitempty"`
	Exchange  string    `gorm:"size:8" json:"exchange,omitempty"`
	ListDate  string    `gorm:"size:8" json:"listDate,omitempty"`
	Delisted  bool      `gorm:"not null;default:false" json:"delisted,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}
