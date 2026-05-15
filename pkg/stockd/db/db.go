// Package db opens the GORM connection and runs AutoMigrate for every models.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"stock/pkg/models"
	"stock/pkg/stockd/config"
)

// Open returns a configured *gorm.DB and runs AutoMigrate.
func Open(cfg *config.Config, logger *logrus.Logger) (*gorm.DB, error) {
	var dialect gorm.Dialector
	switch cfg.Database.Driver {
	case "sqlite":
		dialect = sqlite.Open(cfg.Database.DSN)
	case "mysql":
		dialect = mysql.Open(cfg.Database.DSN)
	case "postgres":
		dialect = postgres.Open(cfg.Database.DSN)
	default:
		return nil, fmt.Errorf("unknown driver %q", cfg.Database.Driver)
	}
	l := GormLogger{
		l: logger,
	}
	gdb, err := gorm.Open(dialect, &gorm.Config{
		Logger:                 l,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := AutoMigrate(gdb); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}
	return gdb, nil
}

// AutoMigrate creates/updates every table managed by stockd.
func AutoMigrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.User{},
		&models.APIToken{},
		&models.Stock{},
		&models.DailyBar{},
		&models.Portfolio{},
		&models.IntradayDraft{},
		&models.JobRun{},
		&models.AnalysisPrediction{},
	)
}

type GormLogger struct {
	l *logrus.Logger
}

func (g GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l := &GormLogger{
		l: g.l,
	}
	return l
}

func (g GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	g.l.Infof(msg, data...)
}

func (g GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	g.l.Warningf(msg, data...)
}

func (g GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	g.l.Errorf(msg, data...)
}

func (g GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil:
		sql, rows := fc()
		if rows == -1 {
			g.l.Debugf("%s [rows:%v] [%.3fms] err %s ", sql, "-", float64(elapsed.Nanoseconds())/1e6, err)
		} else {
			g.l.Debugf("%s [rows:%v] [%.3fms] err %s ", sql, rows, float64(elapsed.Nanoseconds())/1e6, err)

		}
		// case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		// 	sql, rows := fc()
		// 	slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		// 	if rows == -1 {
		// 		l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		// 	} else {
		// 		l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		// 	}
		// case l.LogLevel == Info:
		// 	sql, rows := fc()
		// 	if rows == -1 {
		// 		l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		// 	} else {
		// 		l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		// 	}
	}
	sql, rows := fc()
	if rows == -1 {
		g.l.Debugf("%s [rows:%v] [%.3fms]", sql, "-", float64(elapsed.Nanoseconds())/1e6)
	} else {
		g.l.Debugf("%s [rows:%v] [%.3fms]", sql, rows, float64(elapsed.Nanoseconds())/1e6)
	}
}
