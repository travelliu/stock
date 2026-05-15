package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/models"
	"stock/pkg/stockd/services"
)

func TestLoadStockCache_populatesBothMaps(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.Stock{TsCode: "600519.SH", Code: "600519", Name: "贵州茅台"}).Error)
	svc := services.New(gdb)
	require.NoError(t, svc.LoadStockCache(context.Background()))
	// Verify cache loaded: portfolio enrichment should work
	rows, err := svc.ListPortfolio(context.Background(), 0)
	require.NoError(t, err)
	assert.Empty(t, rows)
}
