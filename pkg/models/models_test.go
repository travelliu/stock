package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/stockd/db"
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

func TestDailyBarHasSpreadColumns(t *testing.T) {
	gdb := openTestDB(t)
	cols := []string{"spread_oh", "spread_ol", "spread_hl", "spread_oc", "spread_hc", "spread_lc"}
	for _, col := range cols {
		assert.True(t, gdb.Migrator().HasColumn(&models.DailyBar{}, col), "column %s should exist", col)
	}

	bar := models.DailyBar{
		TsCode: "000001.SZ", TradeDate: "20250513",
		Open: 10, High: 11, Low: 9, Close: 10.5, Vol: 1000, Amount: 1e4,
		Spreads: models.Spreads{OH: 1, OL: 1, HL: 2, OC: 0.5, HC: 0.5, LC: 1.5},
	}
	require.NoError(t, gdb.Create(&bar).Error)

	var got models.DailyBar
	require.NoError(t, gdb.First(&got, "ts_code = ? AND trade_date = ?", bar.TsCode, bar.TradeDate).Error)
	assert.Equal(t, bar.Spreads, got.Spreads)
}
