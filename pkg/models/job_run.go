package models

import "time"

type JobRun struct {
	ID         uint       `gorm:"primaryKey" json:"id,omitempty"`
	JobName    string     `gorm:"size:64;index;not null" json:"jobName,omitempty"`
	StartedAt  time.Time  `gorm:"not null" json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Status     string     `gorm:"size:16;not null" json:"status,omitempty"` // "running" | "success" | "error"
	Message    string     `gorm:"type:text" json:"message,omitempty"`
}
