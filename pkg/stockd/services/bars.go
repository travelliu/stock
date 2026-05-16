// Package bars stores official daily history and incremental syncs from Tushare.
package services

import (
	"context"
	"errors"
	"stock/pkg/stockd/utils"
	"time"

	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/tushare"
)

func (s *Service) QueryStockDailyBar(ctx context.Context, tsCode, from, to string) ([]models.DailyBar, error) {
	var out []models.DailyBar
	tx := s.db.WithContext(ctx).Where("ts_code = ?", tsCode).Order("trade_date ASC")
	if from != "" {
		tx = tx.Where("trade_date >= ?", from)
	}
	if to != "" {
		tx = tx.Where("trade_date <= ?", to)
	}
	return out, tx.Find(&out).Error
}

// QueryStockDailyBarsPage returns paginated daily bars ordered newest-first.
func (s *Service) QueryStockDailyBarsPage(ctx context.Context, tsCode, from, to string, page, limit int) (*models.BarsPage, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	tx := s.db.WithContext(ctx).Model(&models.DailyBar{}).Where("ts_code = ?", tsCode)
	if from != "" {
		tx = tx.Where("trade_date >= ?", from)
	}
	if to != "" {
		tx = tx.Where("trade_date <= ?", to)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []*models.DailyBar
	err := tx.Order("trade_date DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &models.BarsPage{Items: items, Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) MaxDate(ctx context.Context, tsCode string) (string, error) {
	var row models.DailyBar
	err := s.db.WithContext(ctx).Select("trade_date").
		Where("ts_code = ?", tsCode).
		Order("trade_date DESC").Limit(1).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return row.TradeDate, err
}

// SyncDaily fetches missing bars for tsCode and upserts them. Returns the number
// of rows actually written.
func (s *Service) SyncDaily(ctx context.Context, token, tsCode string) (int, error) {
	from, err := s.MaxDate(ctx, tsCode)
	if err != nil {
		return 0, err
	}
	end := time.Now().Format("20060102")
	if from == "" {
		from = time.Now().AddDate(-10, 0, 0).Format("20060102")
	} else {
		// Inclusive of next day.
		t, _ := time.Parse("20060102", from)
		from = t.AddDate(0, 0, 1).Format("20060102")
		if from > end {
			return 0, nil
		}
	}
	rows, err := tushare.Daily(ctx, s.ts, s.cfg.Tushare.GetDefaultToken(token), tushare.DailyRequest{
		TsCode: tsCode, StartDate: from, EndDate: end,
	})
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		row := &models.DailyBar{
			TsCode: utils.TrimTsCode(r.TsCode), TradeDate: r.TradeDate,
			Open: r.Open, High: r.High, Low: r.Low, Close: r.Close,
			Vol: r.Vol, Amount: r.Amount,
			Spreads: r.Spreads,
		}
		if err := s.db.WithContext(ctx).Save(row).Error; err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
