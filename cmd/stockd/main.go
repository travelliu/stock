package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"stock/pkg/stockd/bootstrap"
	"stock/pkg/stockd/config"
	"stock/pkg/stockd/db"
	httpkg "stock/pkg/stockd/http"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/stockd/services/portfolio"
	"stock/pkg/stockd/services/scheduler"
	"stock/pkg/stockd/services/stock"
	"stock/pkg/tushare"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("stockd %s (built %s)\n", Version, BuildTime)
		return
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfgPath := os.Getenv("STOCKD_CONFIG")
	if cfgPath == "" {
		cfgPath = "/etc/stockd/config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.WithError(err).Fatal("config load failed")
	}

	lvl, _ := logrus.ParseLevel(cfg.Logging.Level)
	logger.SetLevel(lvl)

	gdb, err := db.Open(cfg)
	if err != nil {
		logger.WithError(err).Fatal("database open failed")
	}

	if _, err := bootstrap.EnsureAdmin(gdb, logger); err != nil {
		logger.WithError(err).Fatal("bootstrap failed")
	}

	sched := scheduler.New(gdb)
	if cfg.Scheduler.Enabled {
		tc := tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL))
		barsSvc := bars.New(gdb, tc)
		stockSvc := stock.New(gdb)
		portfolioSvc := portfolio.New(gdb)

		sched.RegisterCron("daily-fetch", cfg.Scheduler.DailyFetchCron, func(ctx context.Context) error {
			codes, err := portfolioSvc.DistinctTsCodes(ctx)
			if err != nil {
				return err
			}
			for _, code := range codes {
				if _, err := barsSvc.Sync(ctx, cfg.Tushare.DefaultToken, code); err != nil {
					logger.WithError(err).WithField("ts_code", code).Error("daily sync failed")
				}
			}
			return nil
		})
		sched.RegisterCron("stocklist-sync", cfg.Scheduler.StocklistSyncCron, func(ctx context.Context) error {
			_, err := stockSvc.SyncFromTushare(ctx, cfg.Tushare.DefaultToken)
			return err
		})
		sched.Start()
		defer sched.Stop()
	}

	router := httpkg.NewRouter(gdb, cfg, sched)

	srv := &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: router,
	}

	go func() {
		logger.WithField("addr", cfg.Server.Listen).Info("starting server")
		var err error
		if cfg.Server.TLS.Enabled {
			err = srv.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("shutdown error")
	}
}
