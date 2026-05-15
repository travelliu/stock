package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stock/pkg/logs"
	"stock/pkg/stockd/services"
	"syscall"
	"time"

	"stock/pkg/stockd/bootstrap"
	"stock/pkg/stockd/config"
	"stock/pkg/stockd/db"
	httpkg "stock/pkg/stockd/http"
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
	fmt.Println(os.Getwd())

	cfgPath := os.Getenv("STOCKD_CONFIG")
	if cfgPath == "" {
		cfgPath = "/etc/stockd/config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("config load failed %s\n", err.Error())
		return
	}
	logger, err := logs.NewLogRusDir(cfg.Logging.Dir, "stockd.log", cfg.Logging.Level)
	if err != nil {
		fmt.Printf("config load failed %s\n", err.Error())
		return
	}

	gdb, err := db.Open(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("database open failed")
		return
	}

	if _, err := bootstrap.EnsureAdmin(gdb, logger); err != nil {
		logger.WithError(err).Fatal("bootstrap failed")
	}

	tc := tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL))
	svc := services.NewService(gdb, tc, cfg, logger)
	if err := svc.LoadStockCache(context.Background()); err != nil {
		logger.WithError(err).Warn("stock cache load failed")
	}
	if cfg.Scheduler.Enabled {
		svc.InitCron()
		svc.StartCron()
		defer svc.StopCron()
	}

	router := httpkg.NewRouter(svc, logger)

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
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
