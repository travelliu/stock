package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/models"
	"stock/pkg/stockd/db"
	"stock/pkg/stockd/services"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestAddListRemove(t *testing.T) {
	svc := services.New(openDB(t))
	ctx := context.Background()
	require.NoError(t, svc.Add(ctx, 1, "600519.SH", "茅台仓"))
	require.NoError(t, svc.Add(ctx, 1, "000001.SZ", ""))
	list, err := svc.List(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	
	require.NoError(t, svc.Remove(ctx, 1, "600519.SH"))
	list, _ = svc.List(ctx, 1)
	assert.Len(t, list, 1)
}

func TestAddDuplicateIsIdempotent(t *testing.T) {
	svc := services.New(openDB(t))
	ctx := context.Background()
	require.NoError(t, svc.Add(ctx, 1, "600519.SH", "v1"))
	require.NoError(t, svc.Add(ctx, 1, "600519.SH", "v2"), "duplicate add overwrites note")
	list, _ := svc.List(ctx, 1)
	require.Len(t, list, 1)
	assert.Equal(t, "v2", list[0].Note)
}

func TestRemoveOnlyAffectsOwner(t *testing.T) {
	svc := services.New(openDB(t))
	ctx := context.Background()
	require.NoError(t, svc.Add(ctx, 1, "600519.SH", ""))
	require.NoError(t, svc.Add(ctx, 2, "600519.SH", ""))
	require.NoError(t, svc.Remove(ctx, 1, "600519.SH"))
	list2, _ := svc.List(ctx, 2)
	assert.Len(t, list2, 1)
}

func TestListPortfolioEnrichesName(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.Stock{TsCode: "600519.SH", Code: "600519", Name: "贵州茅台"}).Error)
	svc := services.New(gdb)
	require.NoError(t, svc.LoadStockCache(context.Background()))
	require.NoError(t, svc.Add(context.Background(), 1, "600519.SH", ""))
	list, err := svc.List(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "贵州茅台", list[0].Name)
	assert.Equal(t, "600519", list[0].Code)
}
