package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/db"
	"stock/pkg/stockd/models"
)

func openTestDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(0)"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestModelsRoundTrip(t *testing.T) {
	gdb := openTestDB(t)
	now := time.Now()
	cases := []any{
		&models.User{Username: "alice", PasswordHash: "x", Role: "user", CreatedAt: now, UpdatedAt: now},
		&models.APIToken{UserID: 1, Name: "cli", TokenHash: "deadbeef", CreatedAt: now},
		&models.Stock{TsCode: "600519.SH", Code: "600519", Name: "贵州茅台", Market: "主板", Exchange: "SSE", UpdatedAt: now},
		&models.DailyBar{TsCode: "600519.SH", TradeDate: "20250513", Open: 1620, High: 1655, Low: 1601, Close: 1632, Vol: 3500, Amount: 5e5},
		&models.Portfolio{UserID: 1, TsCode: "600519.SH", AddedAt: now},
		&models.IntradayDraft{UserID: 1, TsCode: "600519.SH", TradeDate: "20260514", UpdatedAt: now},
		&models.JobRun{JobName: "daily-fetch", StartedAt: now, Status: "running"},
	}
	for _, c := range cases {
		require.NoError(t, gdb.Create(c).Error)
	}
}
