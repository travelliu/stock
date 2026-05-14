package models

import "time"

type Portfolio struct {
	ID      uint      `gorm:"primaryKey"`
	UserID  uint      `gorm:"uniqueIndex:idx_user_code;not null"`
	TsCode  string    `gorm:"uniqueIndex:idx_user_code;size:16;not null"`
	Note    string    `gorm:"size:255"`
	AddedAt time.Time
}
