// Package portfolio implements per-user tracked-stock CRUD.
package portfolio

import (
	"context"
	"time"

	"gorm.io/gorm"

	"stock/pkg/models"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type Entry = models.Portfolio

func (s *Service) Add(ctx context.Context, userID uint, tsCode, note string) error {
	row := &models.Portfolio{
		UserID: userID, TsCode: tsCode, Note: note, AddedAt: time.Now(),
	}
	// Upsert via ON CONFLICT (sqlite + pg + mysql all supported by GORM clause).
	return s.db.WithContext(ctx).
		Where(&models.Portfolio{UserID: userID, TsCode: tsCode}).
		Assign(map[string]any{"note": note}).
		FirstOrCreate(row).Error
}

func (s *Service) Remove(ctx context.Context, userID uint, tsCode string) error {
	return s.db.WithContext(ctx).
		Where("user_id = ? AND ts_code = ?", userID, tsCode).
		Delete(&models.Portfolio{}).Error
}

func (s *Service) UpdateNote(ctx context.Context, userID uint, tsCode, note string) error {
	return s.db.WithContext(ctx).Model(&models.Portfolio{}).
		Where("user_id = ? AND ts_code = ?", userID, tsCode).
		Update("note", note).Error
}

func (s *Service) List(ctx context.Context, userID uint) ([]Entry, error) {
	var rows []models.Portfolio
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("added_at DESC").Find(&rows).Error
	return rows, err
}

// DistinctTsCodes returns every ts_code referenced by any portfolio (used by
// the daily-fetch scheduler).
func (s *Service) DistinctTsCodes(ctx context.Context) ([]string, error) {
	var out []string
	err := s.db.WithContext(ctx).Model(&models.Portfolio{}).
		Distinct("ts_code").Order("ts_code ASC").Pluck("ts_code", &out).Error
	return out, err
}
