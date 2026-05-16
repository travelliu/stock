package services_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/models"
	"stock/pkg/stockd/config"
	"stock/pkg/stockd/services"
	"stock/pkg/tushare"
)

func TestQueryRange(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.StockDailyBar{TsCode: "X.SH", TradeDate: "20250101", Open: 1, High: 2, Low: 0.5, Close: 1.5}).Error)
	require.NoError(t, gdb.Create(&models.StockDailyBar{TsCode: "X.SH", TradeDate: "20250110", Open: 2, High: 3, Low: 1.5, Close: 2.5}).Error)
	svc := services.NewService(gdb, tushare.NewClient(), nil, &config.Config{}, logrus.New())
	out, err := svc.QueryStockDailyBar(context.Background(), "X.SH", "20250101", "20250110")
	require.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestQueryBarsPage(t *testing.T) {
	gdb := openDB(t)
	for i := 1; i <= 25; i++ {
		date := fmt.Sprintf("202501%02d", i)
		require.NoError(t, gdb.Create(&models.StockDailyBar{
			TsCode: "X.SH", TradeDate: date, Open: 1, High: 2, Low: 0.5, Close: 1.5,
		}).Error)
	}
	svc := services.NewService(gdb, tushare.NewClient(), nil, &config.Config{}, logrus.New())
	page, err := svc.QueryStockDailyBarsPage(context.Background(), "X.SH", "", "", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(25), page.Total)
	assert.Len(t, page.Items, 20)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 20, page.Limit)
	page2, _ := svc.QueryStockDailyBarsPage(context.Background(), "X.SH", "", "", 2, 20)
	assert.Len(t, page2.Items, 5)
}

func TestSync_FromTushare(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
		  "code":0,"msg":"","data":{
		    "fields":["ts_code","trade_date","open","high","low","close","vol","amount"],
		    "items":[["X.SH","20250101",1,2,0.5,1.5,1000,1000]]
		  }}`)
	}))
	defer srv.Close()
	gdb := openDB(t)
	svc := services.NewService(gdb, tushare.NewClient(tushare.WithBaseURL(srv.URL)), nil, &config.Config{}, logrus.New())
	n, err := svc.SyncDaily(context.Background(), "tok", "X.SH")
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	var bar models.StockDailyBar
	require.NoError(t, gdb.First(&bar).Error)
	assert.InDelta(t, 1.0, bar.Spreads.OH, 1e-9)
}
