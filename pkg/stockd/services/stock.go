// Package services provides catalog search and sync.
package services

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/tushare"
)

func (s *Service) GetStock(ctx context.Context, tsCode string) (*models.StockBasicInfo, error) {
	var row models.StockBasicInfo
	if err := s.db.WithContext(ctx).First(&row, "ts_code = ? or code = ?", tsCode, tsCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// SearchStock matches by ts_code prefix, code prefix, or name substring.
func (s *Service) SearchStock(ctx context.Context, q string, limit int) ([]models.StockBasicInfo, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q = strings.TrimSpace(q)
	var out []models.StockBasicInfo
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
		row := &models.StockBasicInfo{
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
