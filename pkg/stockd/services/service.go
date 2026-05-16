package services

import (
	"stock/pkg/baidu"
	"stock/pkg/models"
	"stock/pkg/stockd/config"
	"stock/pkg/tencent"
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
	tc     *tencent.Client
	cfg    *config.Config
	cron   *cron.Cron
	mu     sync.Mutex
	jobs   map[string]JobFunc
	sf     singleflight.Group
	logger *logrus.Logger

	stockCacheByCode   map[string]*models.StockBasicInfo
	stockCacheByTsCode map[string]*models.StockBasicInfo
	cacheMu            sync.RWMutex

	realtimeCache map[string]*models.StockRealtimeAndAnalysis
	realtimeMu    sync.RWMutex

	baiduClient *baidu.Client
}

// ServiceOption configures Service after construction.
type ServiceOption func(*Service)

// WithBaiduClient injects a Baidu PAE client.
func WithBaiduClient(c *baidu.Client) ServiceOption {
	return func(s *Service) { s.baiduClient = c }
}

func NewService(db *gorm.DB, ts *tushare.Client, tc *tencent.Client, cfg *config.Config,
	logger *logrus.Logger, opts ...ServiceOption) *Service {
	svc := &Service{
		db:            db,
		ts:            ts,
		tc:            tc,
		cfg:           cfg,
		cron:          cron.New(cron.WithLocation(time.Local)),
		jobs:          map[string]JobFunc{},
		logger:        logger,
		realtimeCache: make(map[string]*models.StockRealtimeAndAnalysis),
	}
	for _, o := range opts {
		o(svc)
	}
	return svc
}

func (s *Service) GetDB() *gorm.DB           { return s.db }
func (s *Service) GetTS() *tushare.Client    { return s.ts }
func (s *Service) GetConfig() *config.Config { return s.cfg }
