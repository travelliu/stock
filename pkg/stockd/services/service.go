// 2026/5/15 Bin Liu <bin.liu@enmotech.com>

package services

import (
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

	stockCacheByCode   map[string]*models.Stock
	stockCacheByTsCode map[string]*models.Stock
	cacheMu            sync.RWMutex
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
