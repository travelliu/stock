package services

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm/clause"

	"stock/pkg/models"
	"stock/pkg/stockd/services/analysis"
)

type RecalcResult struct {
	Upserted int `json:"updated"`
}

func (s *Service) Recalc(ctx context.Context, tsCode string) (*RecalcResult, error) {
	var portfolios []models.Portfolio
	q := s.db.WithContext(ctx)
	if tsCode != "" {
		q = q.Where("ts_code = ?", tsCode)
	}
	if err := q.Find(&portfolios).Error; err != nil {
		return nil, fmt.Errorf("list portfolio: %w", err)
	}

	total := 0
	for _, p := range portfolios {
		n, err := s.recalcStock(ctx, p.TsCode)
		if err != nil {
			return nil, fmt.Errorf("recalc %s: %w", p.TsCode, err)
		}
		total += n
	}
	return &RecalcResult{Upserted: total}, nil
}

func (s *Service) recalcStock(ctx context.Context, tsCode string) (int, error) {
	var bars []*models.DailyBar
	if err := s.db.WithContext(ctx).
		Where("ts_code = ?", tsCode).
		Order("trade_date ASC").
		Find(&bars).Error; err != nil {
		return 0, err
	}
	if len(bars) < 16 {
		return 0, nil
	}

	count := 0
	for i := 15; i < len(bars); i++ {
		historical := bars[:i]
		today := bars[i]
		if today.Open == 0 {
			continue
		}

		windows := analysis.Make(historical)
		analysis.Means(windows)

		var ohMean, olMean float64
		for _, w := range windows {
			if w.Info.Id == "last_15" && w.Means != nil {
				if w.Means.SpreadOH != nil {
					ohMean = w.Means.SpreadOH.Mean
				}
				if w.Means.SpreadOL != nil {
					olMean = w.Means.SpreadOL.Mean
				}
			}
		}
		if ohMean == 0 {
			continue
		}

		sampleCounts := make(map[string]int)
		for _, w := range windows {
			sampleCounts[w.Info.Id] = len(w.Rows)
		}
		sampleJSON, _ := json.Marshal(sampleCounts)
		wmJSON, _ := json.Marshal(windows)
		comp := analysis.Composite(windows)
		compJSON, _ := json.Marshal(comp)

		p := models.AnalysisPrediction{
			TsCode:         tsCode,
			TradeDate:      today.TradeDate,
			SampleCounts:   sampleJSON,
			WindowMeans:    wmJSON,
			CompositeMeans: compJSON,
			OpenPrice:      today.Open,
			PredictHigh:    today.Open + ohMean,
			PredictLow:     today.Open - olMean,
			ActualHigh:     today.High,
			ActualLow:      today.Low,
			ActualClose:    today.Close,
		}

		err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "ts_code"}, {Name: "trade_date"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"sample_counts", "window_means", "composite_means",
				"open_price", "predict_high", "predict_low", "predict_close",
				"actual_high", "actual_low", "actual_close", "updated_at",
			}),
		}).Create(&p).Error
		if err != nil {
			return count, fmt.Errorf("upsert %s %s: %w", tsCode, today.TradeDate, err)
		}
		count++
	}
	return count, nil
}

// PredictionsPage is the paginated response from ListPredictionsPage.
type PredictionsPage struct {
	Items []models.AnalysisPrediction `json:"items"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Limit int                         `json:"limit"`
}

// ListPredictionsPage returns paginated predictions ordered newest-first.
func (s *Service) ListPredictionsPage(ctx context.Context, tsCode, from, to string, page, limit int) (*PredictionsPage, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q := s.db.WithContext(ctx).Model(&models.AnalysisPrediction{}).Where("ts_code = ?", tsCode)
	if from != "" {
		q = q.Where("trade_date >= ?", from)
	}
	if to != "" {
		q = q.Where("trade_date <= ?", to)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []models.AnalysisPrediction
	err := q.Order("trade_date DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PredictionsPage{Items: items, Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) ListAnalysisPrediction(ctx context.Context, tsCode, from, to string, limit int) ([]models.AnalysisPrediction, error) {
	if limit <= 0 {
		limit = 30
	}
	q := s.db.WithContext(ctx).Where("ts_code = ?", tsCode).Order("trade_date DESC")
	if from != "" {
		q = q.Where("trade_date >= ?", from)
	}
	if to != "" {
		q = q.Where("trade_date <= ?", to)
	}
	var preds []models.AnalysisPrediction
	if err := q.Limit(limit).Find(&preds).Error; err != nil {
		return nil, err
	}
	return preds, nil
}
