// Package analysis orchestrates the analysis pipeline:
//   - load all daily_bars for the ts_code
//   - merge today's intraday_draft (when WithDraft=true and a row exists)
//   - apply explicit param overrides (always win)
//   - call pkg/analysis.Build
package analysis

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"stock/pkg/models"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type Input struct {
	UserID      uint
	TsCode      string
	OpenPrice   *float64
	ActualHigh  *float64
	ActualLow   *float64
	ActualClose *float64
	WithDraft   bool
}

func (s *Service) Run(ctx context.Context, in Input) (*models.AnalysisResult, error) {
	var bars []*models.DailyBar
	err := s.db.WithContext(ctx).Where("ts_code = ?", in.TsCode).Order("trade_date ASC").Find(&bars).Error
	if err != nil {
		return &models.AnalysisResult{}, err
	}

	if in.WithDraft {
		today := time.Now().Format("20060102")
		var d models.IntradayDraft
		err := s.db.WithContext(ctx).
			Where("user_id = ? AND ts_code = ? AND trade_date = ?", in.UserID, in.TsCode, today).
			First(&d).Error
		if err == nil {
			if in.OpenPrice == nil {
				in.OpenPrice = d.Open
			}
			if in.ActualHigh == nil {
				in.ActualHigh = d.High
			}
			if in.ActualLow == nil {
				in.ActualLow = d.Low
			}
			if in.ActualClose == nil {
				in.ActualClose = d.Close
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.AnalysisResult{}, err
		}
	}

	var name string
	var st models.Stock
	if s.db.WithContext(ctx).First(&st, "ts_code = ? or code = ?", in.TsCode, in.TsCode).Error == nil {
		name = st.Name
	}

	res := Build(models.Input{
		TsCode: in.TsCode, StockName: name,
		Rows:        bars,
		OpenPrice:   in.OpenPrice,
		ActualHigh:  in.ActualHigh,
		ActualLow:   in.ActualLow,
		ActualClose: in.ActualClose,
	})
	return res, nil
}
