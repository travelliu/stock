package models

import "time"

type Stock struct {
	TsCode    string `gorm:"primaryKey;size:16"`
	Code      string `gorm:"index;size:8;not null"`
	Name      string `gorm:"size:32;not null"`
	Area      string `gorm:"size:16"`
	Industry  string `gorm:"size:32"`
	Market    string `gorm:"size:16"`
	Exchange  string `gorm:"size:8"`
	ListDate  string `gorm:"size:8"`
	Delisted  bool   `gorm:"not null;default:false"`
	UpdatedAt time.Time
}
