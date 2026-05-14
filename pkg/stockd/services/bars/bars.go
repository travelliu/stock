// Package bars stores official daily history and incremental syncs from Tushare.
package bars

import (
	"context"
	"errors"
	"stock/pkg/stockd/utils"
	"time"

	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/tushare"
)

type Service struct {
	db *gorm.DB
	ts *tushare.Client
}

func New(db *gorm.DB, ts *tushare.Client) *Service { return &Service{db: db, ts: ts} }

func (s *Service) Query(ctx context.Context, tsCode, from, to string) ([]models.DailyBar, error) {
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

// Sync fetches missing bars for tsCode and upserts them. Returns the number
// of rows actually written.
func (s *Service) Sync(ctx context.Context, token, tsCode string) (int, error) {
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
	rows, err := tushare.Daily(ctx, s.ts, token, tushare.DailyRequest{
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
