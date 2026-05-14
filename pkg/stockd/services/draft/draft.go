// Package draft stores user-entered intraday OHLC drafts.
package draft

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"stock/pkg/stockd/models"
)

var (
	ErrInvalid  = errors.New("invalid draft (need at least one field; high>=open,close; low<=open,close)")
	ErrNotFound = errors.New("draft not found")
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type UpsertInput struct {
	UserID    uint
	TsCode    string
	TradeDate string
	Open      *float64
	High      *float64
	Low       *float64
	Close     *float64
}

func (s *Service) Upsert(ctx context.Context, in UpsertInput) (*models.IntradayDraft, error) {
	if in.Open == nil && in.High == nil && in.Low == nil && in.Close == nil {
		return nil, ErrInvalid
	}
	if !validate(in) {
		return nil, ErrInvalid
	}
	var row models.IntradayDraft
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND ts_code = ? AND trade_date = ?", in.UserID, in.TsCode, in.TradeDate).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.IntradayDraft{
			UserID: in.UserID, TsCode: in.TsCode, TradeDate: in.TradeDate,
			Open: in.Open, High: in.High, Low: in.Low, Close: in.Close,
			UpdatedAt: time.Now(),
		}
		return &row, s.db.WithContext(ctx).Create(&row).Error
	}
	if err != nil {
		return nil, err
	}
	if in.Open != nil {
		row.Open = in.Open
	}
	if in.High != nil {
		row.High = in.High
	}
	if in.Low != nil {
		row.Low = in.Low
	}
	if in.Close != nil {
		row.Close = in.Close
	}
	row.UpdatedAt = time.Now()
	return &row, s.db.WithContext(ctx).Save(&row).Error
}

func (s *Service) GetByDate(ctx context.Context, userID uint, tsCode, tradeDate string) (*models.IntradayDraft, error) {
	var row models.IntradayDraft
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND ts_code = ? AND trade_date = ?", userID, tsCode, tradeDate).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &row, err
}

func (s *Service) Delete(ctx context.Context, userID, id uint) error {
	return s.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).Delete(&models.IntradayDraft{}).Error
}

func validate(in UpsertInput) bool {
	hi := bestFloat(in.High)
	lo := bestFloat(in.Low)
	op := bestFloat(in.Open)
	cl := bestFloat(in.Close)
	if in.High != nil {
		if in.Open != nil && hi < op {
			return false
		}
		if in.Close != nil && hi < cl {
			return false
		}
	}
	if in.Low != nil {
		if in.Open != nil && lo > op {
			return false
		}
		if in.Close != nil && lo > cl {
			return false
		}
	}
	return true
}

func bestFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
