package services_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"stock/pkg/stockd/config"
	"stock/pkg/stockd/services"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"stock/pkg/models"
	
	"stock/pkg/tushare"
)

func TestQueryRange(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250101", Open: 1, High: 2, Low: 0.5, Close: 1.5}).Error)
	require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250110", Open: 2, High: 3, Low: 1.5, Close: 2.5}).Error)
	svc := services.NewService(gdb, tushare.NewClient(), &config.Config{})
	out, err := svc.QueryStockDailyBar(context.Background(), "X.SH", "20250101", "20250110")
	require.NoError(t, err)
	assert.Len(t, out, 2)
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
	svc := services.NewService(gdb, tushare.NewClient(tushare.WithBaseURL(srv.URL)), &config.Config{})
	n, err := svc.SyncDaily(context.Background(), "tok", "X.SH")
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	var bar models.DailyBar
	require.NoError(t, gdb.First(&bar).Error)
	assert.InDelta(t, 1.0, bar.Spreads.OH, 1e-9)
}
