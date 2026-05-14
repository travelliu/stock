// Package db opens the GORM connection and runs AutoMigrate for every models.
package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"stock/pkg/stockd/config"
	"stock/pkg/models"
)

// Open returns a configured *gorm.DB and runs AutoMigrate.
func Open(cfg *config.Config) (*gorm.DB, error) {
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
	gdb, err := gorm.Open(dialect, &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(gormlogger.Warn),
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
	)
}
