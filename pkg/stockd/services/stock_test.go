package services_test

import (
	"context"
	"stock/pkg/stockd/services"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/models"
)

func TestSearchByCodeAndName(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.Stock{TsCode: "600519.SH", Code: "600519", Name: "贵州茅台"}).Error)
	require.NoError(t, gdb.Create(&models.Stock{TsCode: "603778.SH", Code: "603778", Name: "千金药业"}).Error)
	svc := services.New(gdb)
	got, _ := svc.Search(context.Background(), "茅", 10)
	assert.Len(t, got, 1)
	got, _ = svc.Search(context.Background(), "60", 10)
	assert.Len(t, got, 2)
}

func TestImportCSV(t *testing.T) {
	gdb := openDB(t)
	svc := services.New(gdb)
	csv := strings.NewReader("ts_code,symbol,name,area,industry,market,exchange,list_date\n" +
		"600519.SH,600519,贵州茅台,贵州,白酒,主板,SSE,20010827\n" +
		"000001.SZ,000001,平安银行,深圳,银行,主板,SZSE,19910403\n")
	n, err := svc.ImportCSV(context.Background(), csv)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}
