package models

import "time"

type JobRun struct {
	ID         uint      `gorm:"primaryKey"`
	JobName    string    `gorm:"size:64;index;not null"`
	StartedAt  time.Time `gorm:"not null"`
	FinishedAt *time.Time
	Status     string `gorm:"size:16;not null"` // "running" | "success" | "error"
	Message    string `gorm:"type:text"`
}
