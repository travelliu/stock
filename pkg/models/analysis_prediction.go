package models

import (
	"encoding/json"
	"time"
)

type AnalysisPrediction struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	TsCode         string          `gorm:"uniqueIndex:idx_pred_code_date;size:16;not null" json:"tsCode"`
	TradeDate      string          `gorm:"uniqueIndex:idx_pred_code_date;size:8;not null" json:"tradeDate"`
	SampleCounts   json.RawMessage `gorm:"type:json" json:"sampleCounts"`
	WindowMeans    json.RawMessage `gorm:"type:json" json:"windowMeans"`
	CompositeMeans json.RawMessage `gorm:"type:json" json:"compositeMeans"`
	OpenPrice      float64         `json:"openPrice"`
	PredictHigh    float64         `json:"predictHigh"`
	PredictLow     float64         `json:"predictLow"`
	PredictClose   float64         `json:"predictClose"`
	ActualHigh     float64         `json:"actualHigh"`
	ActualLow      float64         `json:"actualLow"`
	ActualClose    float64         `json:"actualClose"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}
