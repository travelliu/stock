package models_test

import (
	"encoding/json"
	"fmt"
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

func TestModelJSONFields(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		val     any
		want    []string
		notWant []string
	}{
		{
			name: "User",
			val: models.User{
				ID: 1, Username: "alice", PasswordHash: "h", Role: "user",
				TushareToken: "tk", Disabled: false, CreatedAt: now, UpdatedAt: now,
			},
			want:    []string{"id", "username", "role", "tushareToken", "createdAt", "updatedAt"},
			notWant: []string{"userName", "userID"},
		},
		{
			name:    "Stock",
			val:     models.Stock{TsCode: "600519.SH", Code: "600519", Name: "贵州茅台", Market: "主板", Exchange: "SSE", ListDate: "20010827"},
			want:    []string{"tsCode", "code", "name", "market", "exchange", "listDate"},
			notWant: []string{"createdAt"},
		},
		{
			name:    "Portfolio",
			val:     models.Portfolio{ID: 1, UserID: 2, TsCode: "600519.SH", Note: "n", AddedAt: now},
			want:    []string{"id", "userId", "tsCode", "note", "addedAt"},
			notWant: []string{"userID"},
		},
		{
			name: "PortfolioReq",
			val:  models.PortfolioReq{TsCode: "600519.SH", Note: "test"},
			want: []string{"tsCode", "note"},
		},
		{
			name:    "APIToken",
			val:     models.APIToken{ID: 1, UserID: 2, Name: "cli", TokenHash: "h", CreatedAt: now},
			want:    []string{"id", "userId", "name", "tokenHash", "createdAt"},
			notWant: []string{"userID"},
		},
		{
			name:    "IntradayDraft",
			val:     models.IntradayDraft{ID: 1, UserID: 2, TsCode: "x", TradeDate: "20250513", UpdatedAt: now},
			want:    []string{"id", "userId", "tsCode", "tradeDate", "updatedAt"},
			notWant: []string{"userID"},
		},
		{
			name: "DailyBar",
			val:  models.DailyBar{TsCode: "x", TradeDate: "20250513", Open: 10, High: 11, Low: 9, Close: 10, Spreads: models.Spreads{OH: 1, HL: 2}},
			want: []string{"tsCode", "tradeDate", "open", "high", "low", "close", "spreads"},
		},
		{
			name:    "JobRun",
			val:     models.JobRun{ID: 1, JobName: "daily-fetch", StartedAt: now, Status: "success"},
			want:    []string{"id", "jobName", "startedAt", "status"},
			notWant: []string{"createdAt"},
		},
		{
			name: "LoginReq",
			val:  models.LoginReq{Username: "alice", Password: "secret"},
			want: []string{"username", "password"},
		},
		{
			name: "IssueTokenResp",
			val:  models.IssueTokenResp{Token: "tk", Metadata: &models.APIToken{ID: 1, Name: "x"}},
			want: []string{"token", "metadata"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, err := json.Marshal(c.val)
			require.NoError(t, err)
			s := string(b)
			for _, w := range c.want {
				assert.Contains(t, s, fmt.Sprintf(`"%s"`, w))
			}
			for _, nw := range c.notWant {
				assert.NotContains(t, s, fmt.Sprintf(`"%s"`, nw))
			}
		})
	}
}
