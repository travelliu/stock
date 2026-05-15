package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username,omitempty"`
	PasswordHash string    `gorm:"not null" json:"passwordHash,omitempty"`
	Role         string    `gorm:"size:16;not null" json:"role,omitempty"`
	TushareToken string    `gorm:"size:128" json:"tushareToken,omitempty"`
	Disabled     bool      `gorm:"not null;default:false" json:"disabled,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
