package analysis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/stockd/db"
	"stock/pkg/stockd/services/analysis"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func p(v float64) *float64 { return &v }

func TestRun_DrawsOpenFromDraftIfProvided(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250513", Open: 100, High: 102, Low: 98, Close: 101, Spreads: models.Spreads{OH: 2, OL: 2, HL: 4}}).Error)
	today := time.Now().Format("20060102")
	require.NoError(t, gdb.Create(&models.IntradayDraft{UserID: 1, TsCode: "X.SH", TradeDate: today, Open: p(105)}).Error)

	svc := analysis.New(gdb)
	res, err := svc.Run(context.Background(), analysis.Input{UserID: 1, TsCode: "X.SH", WithDraft: true})
	require.NoError(t, err)
	require.NotNil(t, res.OpenPrice)
	assert.Equal(t, 105.0, *res.OpenPrice)
}

func TestRun_ExplicitOverridesBeatDraft(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250513", Open: 100, High: 102, Low: 98, Close: 101}).Error)
	today := time.Now().Format("20060102")
	require.NoError(t, gdb.Create(&models.IntradayDraft{UserID: 1, TsCode: "X.SH", TradeDate: today, Open: p(105)}).Error)

	svc := analysis.New(gdb)
	res, err := svc.Run(context.Background(), analysis.Input{UserID: 1, TsCode: "X.SH", WithDraft: true, OpenPrice: p(110)})
	require.NoError(t, err)
	require.NotNil(t, res.OpenPrice)
	assert.Equal(t, 110.0, *res.OpenPrice)
}
