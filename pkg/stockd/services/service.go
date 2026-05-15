// 2026/5/15 Bin Liu <bin.liu@enmotech.com>

package services

import (
	"context"
	"stock/pkg/models"
	"stock/pkg/stockd/config"
	"stock/pkg/tushare"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	ts     *tushare.Client
	cfg    *config.Config
	cron   *cron.Cron
	mu     sync.Mutex
	jobs   map[string]JobFunc
	sf     singleflight.Group
	logger *logrus.Logger
}

func NewService(db *gorm.DB, ts *tushare.Client, cfg *config.Config,
	logger *logrus.Logger) *Service {
	return &Service{db: db, ts: ts, cfg: cfg,
		cron:   cron.New(cron.WithLocation(time.Local)),
		jobs:   map[string]JobFunc{},
		logger: logger,
	}
}

func (s *Service) GetDB() *gorm.DB {
	return s.db
}

func (s *Service) GetTS() *tushare.Client {
	return s.ts
}

func (s *Service) GetConfig() *config.Config {
	return s.cfg
}

// New creates a minimal Service for unit tests (no scheduler, tushare, or logger).
func New(db *gorm.DB) *Service {
	return &Service{db: db, jobs: map[string]JobFunc{}}
}

// Search is a test-friendly alias for SearchStock.
func (s *Service) Search(ctx context.Context, q string, limit int) ([]models.Stock, error) {
	return s.SearchStock(ctx, q, limit)
}

// Add is a test-friendly alias for AddPortfolio.
func (s *Service) Add(ctx context.Context, userID uint, tsCode, note string) error {
	return s.AddPortfolio(ctx, userID, tsCode, note)
}

// List is a test-friendly alias for ListPortfolio.
func (s *Service) List(ctx context.Context, userID uint) ([]*models.Portfolio, error) {
	return s.ListPortfolio(ctx, userID)
}

// Remove is a test-friendly alias for RemovePortfolio.
func (s *Service) Remove(ctx context.Context, userID uint, tsCode string) error {
	return s.RemovePortfolio(ctx, userID, tsCode)
}

// Upsert is a test-friendly alias for UpsertDraft.
func (s *Service) Upsert(ctx context.Context, in UpsertInput) (*models.IntradayDraft, error) {
	return s.UpsertDraft(ctx, in)
}

// GetByDate is a test-friendly alias for GetDraftByDate.
func (s *Service) GetByDate(ctx context.Context, userID uint, tsCode, date string) (*models.IntradayDraft, error) {
	return s.GetDraftByDate(ctx, userID, tsCode, date)
}

// Revoke is a test-friendly alias for RevokeToken.
func (s *Service) Revoke(ctx context.Context, userID, id uint) error {
	return s.RevokeToken(ctx, userID, id)
}

// Create is a test-friendly alias for CreateUser.
func (s *Service) Create(ctx context.Context, in CreateInput) (*User, error) {
	return s.CreateUser(ctx, in)
}

// SetDisabled is a test-friendly alias for SetUserDisabled.
func (s *Service) SetDisabled(ctx context.Context, id uint, disabled bool) error {
	return s.SetUserDisabled(ctx, id, disabled)
}
