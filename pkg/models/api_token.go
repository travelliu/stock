package models

import "time"

type APIToken struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     uint   `gorm:"index;not null"`
	Name       string `gorm:"size:64;not null"`
	TokenHash  string `gorm:"uniqueIndex;size:64;not null"`
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}
