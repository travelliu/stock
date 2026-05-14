// Package stock provides catalog search and sync.
package stock

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/tushare"
)

type Service struct {
	db *gorm.DB
	ts *tushare.Client
}

func New(db *gorm.DB) *Service { return &Service{db: db, ts: tushare.NewClient()} }

func NewWithClient(db *gorm.DB, c *tushare.Client) *Service { return &Service{db: db, ts: c} }

func (s *Service) Get(ctx context.Context, tsCode string) (*models.Stock, error) {
	var row models.Stock
	if err := s.db.WithContext(ctx).First(&row, "ts_code = ?", tsCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &row, nil
}

// Search matches by ts_code prefix, code prefix, or name substring.
func (s *Service) Search(ctx context.Context, q string, limit int) ([]models.Stock, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q = strings.TrimSpace(q)
	var out []models.Stock
	tx := s.db.WithContext(ctx).Limit(limit)
	if q == "" {
		return out, tx.Order("ts_code ASC").Find(&out).Error
	}
	pat := "%" + q + "%"
	return out, tx.Where(
		"ts_code LIKE ? OR code LIKE ? OR name LIKE ?", pat, pat, pat,
	).Order("ts_code ASC").Find(&out).Error
}

// SyncFromTushare upserts the full stock_basic catalog.
func (s *Service) SyncFromTushare(ctx context.Context, token string) (int, error) {
	rows, err := tushare.StockBasic(ctx, s.ts, token, tushare.StockBasicRequest{ListStatus: "L"})
	if err != nil {
		return 0, err
	}
	return s.upsertRows(ctx, rows)
}

func (s *Service) upsertRows(ctx context.Context, rows []tushare.StockBasicRow) (int, error) {
	n := 0
	for _, r := range rows {
		row := &models.Stock{
			TsCode: r.TsCode, Code: r.Symbol, Name: r.Name,
			Area: r.Area, Industry: r.Industry,
			Market: r.Market, Exchange: r.Exchange, ListDate: r.ListDate,
		}
		if err := s.db.WithContext(ctx).Save(row).Error; err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
