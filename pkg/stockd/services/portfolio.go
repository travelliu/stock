// Package services implements per-user tracked-stock CRUD.
package services

import (
	"context"
	"strings"
	"time"

	"stock/pkg/models"
)

func (s *Service) AddPortfolio(ctx context.Context, userID uint, tsCode, note string) error {
	row := &models.Portfolio{
		UserID: userID, TsCode: tsCode, Note: note, AddedAt: time.Now(),
	}
	// Upsert via ON CONFLICT (sqlite + pg + mysql all supported by GORM clause).
	return s.db.WithContext(ctx).
		Where(&models.Portfolio{UserID: userID, TsCode: tsCode}).
		Assign(map[string]any{"note": note}).
		FirstOrCreate(row).Error
}

func (s *Service) RemovePortfolio(ctx context.Context, userID uint, tsCode string) error {
	return s.db.WithContext(ctx).
		Where("user_id = ? AND ts_code = ?", userID, tsCode).
		Delete(&models.Portfolio{}).Error
}

func (s *Service) UpdatePortfolioNote(ctx context.Context, userID uint, tsCode, note string) error {
	return s.db.WithContext(ctx).Model(&models.Portfolio{}).
		Where("user_id = ? AND ts_code = ?", userID, tsCode).
		Update("note", note).Error
}

func (s *Service) ListPortfolio(ctx context.Context, userID uint) ([]*models.Portfolio, error) {
	var rows []*models.Portfolio
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("added_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	s.cacheMu.RLock()
	for _, r := range rows {
		if strings.Contains(r.TsCode, ".") {
			if info, ok := s.stockCacheByTsCode[r.TsCode]; ok {
				r.Name = info.Name
				r.Code = info.Code
			}
		} else {
			if info, ok := s.stockCacheByCode[r.TsCode]; ok {
				r.Name = info.Name
				r.Code = info.Code
			}
		}
	}
	s.cacheMu.RUnlock()
	return rows, nil
}

// DistinctTsCodes returns every ts_code referenced by any portfolio (used by
// the daily-fetch scheduler).
func (s *Service) DistinctTsCodes(ctx context.Context) ([]string, error) {
	var out []string
	err := s.db.WithContext(ctx).Model(&models.Portfolio{}).
		Distinct("ts_code").Order("ts_code ASC").Pluck("ts_code", &out).Error
	return out, err
}
