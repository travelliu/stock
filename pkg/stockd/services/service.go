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

// WithTushareClient overrides the default Tushare client (used in tests).
func WithTushareClient(c *tushare.Client) ServiceOption {
	return func(s *Service) { s.ts = c }
}

// WithBaiduClient overrides the default Baidu PAE client (used in tests).
func WithBaiduClient(c *baidu.Client) ServiceOption {
	return func(s *Service) { s.baiduClient = c }
}

func NewService(db *gorm.DB, cfg *config.Config, logger *logrus.Logger, opts ...ServiceOption) *Service {
	svc := &Service{
		db:            db,
		cfg:           cfg,
		cron:          cron.New(cron.WithLocation(time.Local)),
		jobs:          map[string]JobFunc{},
		logger:        logger,
		realtimeCache: make(map[string]*models.StockRealtimeAndAnalysis),
		tc:            tencent.NewClient(),
		baiduClient:   baidu.NewClient(),
	}
	if cfg != nil {
		svc.ts = tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL))
	}
	for _, o := range opts {
		o(svc)
	}
	return svc
}

func (s *Service) GetDB() *gorm.DB           { return s.db }
func (s *Service) GetConfig() *config.Config { return s.cfg }
