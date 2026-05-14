package bars_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stock/pkg/stockd/db"
	"stock/pkg/models"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/tushare"
)

func openDB(t *testing.T) *gorm.DB {
	gdb, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(gdb))
	return gdb
}

func TestQueryRange(t *testing.T) {
	gdb := openDB(t)
	require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250101", Open: 1, High: 2, Low: 0.5, Close: 1.5}).Error)
	require.NoError(t, gdb.Create(&models.DailyBar{TsCode: "X.SH", TradeDate: "20250110", Open: 2, High: 3, Low: 1.5, Close: 2.5}).Error)
	svc := bars.New(gdb, tushare.NewClient())
	out, err := svc.Query(context.Background(), "X.SH", "20250101", "20250110")
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
	svc := bars.New(gdb, tushare.NewClient(tushare.WithBaseURL(srv.URL)))
	n, err := svc.Sync(context.Background(), "tok", "X.SH")
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	var bar models.DailyBar
	require.NoError(t, gdb.First(&bar).Error)
	assert.InDelta(t, 1.0, bar.Spreads.OH, 1e-9)
}
