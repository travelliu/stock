package draft_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/db"
	"stock/pkg/stockd/services/draft"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func p(v float64) *float64 { return &v }

func TestUpsertAndGet(t *testing.T) {
	svc := draft.New(openDB(t))
	ctx := context.Background()
	_, err := svc.Upsert(ctx, draft.UpsertInput{
		UserID: 1, TsCode: "600519.SH", TradeDate: "20260514",
		Open: p(1620), High: p(1655), Low: p(1601), Close: p(1632),
	})
	require.NoError(t, err)
	got, err := svc.GetByDate(ctx, 1, "600519.SH", "20260514")
	require.NoError(t, err)
	assert.NotNil(t, got.High)
	assert.Equal(t, 1655.0, *got.High)
}

func TestUpsertOverwrites(t *testing.T) {
	svc := draft.New(openDB(t))
	ctx := context.Background()
	_, err := svc.Upsert(ctx, draft.UpsertInput{
		UserID: 1, TsCode: "X.SH", TradeDate: "20260514",
		Open: p(10), High: p(11), Low: p(9), Close: p(10.5),
	})
	require.NoError(t, err)
	_, err = svc.Upsert(ctx, draft.UpsertInput{
		UserID: 1, TsCode: "X.SH", TradeDate: "20260514",
		High: p(12),
	})
	require.NoError(t, err)
	got, _ := svc.GetByDate(ctx, 1, "X.SH", "20260514")
	assert.Equal(t, 12.0, *got.High)
	assert.Equal(t, 10.0, *got.Open, "unspecified fields preserved")
}

func TestUpsertRejectsInvalidHighLow(t *testing.T) {
	svc := draft.New(openDB(t))
	_, err := svc.Upsert(context.Background(), draft.UpsertInput{
		UserID: 1, TsCode: "Y.SH", TradeDate: "20260514",
		Open: p(10), High: p(8), Low: p(9), Close: p(10),
	})
	assert.ErrorIs(t, err, draft.ErrInvalid)
}
