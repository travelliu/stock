package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/models"
	"stock/pkg/stockd/services"
)

func TestResolveTsCode_alreadyFullCode(t *testing.T) {
	svc := services.New(openDB(t))
	require.NoError(t, svc.LoadStockCache(context.Background()))
	got, err := svc.ResolveTsCode("300476.SZ")
	require.NoError(t, err)
	assert.Equal(t, "300476.SZ", got)
}

func TestResolveTsCode_shortCode(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.Stock{TsCode: "300476.SZ", Code: "300476", Name: "千方科技"}).Error)
	svc := services.New(gdb)
	require.NoError(t, svc.LoadStockCache(context.Background()))
	got, err := svc.ResolveTsCode("300476")
	require.NoError(t, err)
	assert.Equal(t, "300476.SZ", got)
}

func TestResolveTsCode_notFound(t *testing.T) {
	svc := services.New(openDB(t))
	require.NoError(t, svc.LoadStockCache(context.Background()))
	_, err := svc.ResolveTsCode("999999")
	assert.Error(t, err)
}

func TestLoadStockCache_populatesBothMaps(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.Stock{TsCode: "600519.SH", Code: "600519", Name: "贵州茅台"}).Error)
	svc := services.New(gdb)
	require.NoError(t, svc.LoadStockCache(context.Background()))
	ts, err := svc.ResolveTsCode("600519")
	require.NoError(t, err)
	assert.Equal(t, "600519.SH", ts)
	ts2, err2 := svc.ResolveTsCode("600519.SH")
	require.NoError(t, err2)
	assert.Equal(t, "600519.SH", ts2)
}
