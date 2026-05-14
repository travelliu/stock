package tushare_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/tushare"
)

func TestStockBasic_ParsesItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
		  "code":0,"msg":"","data":{
		    "fields":["ts_code","symbol","name","area","industry","market","exchange","list_date"],
		    "items":[
		      ["600519.SH","600519","贵州茅台","贵州","白酒","主板","SSE","20010827"],
		      ["000001.SZ","000001","平安银行","深圳","银行","主板","SZSE","19910403"]
		    ]
		  }
		}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL))
	out, err := tushare.StockBasic(context.Background(), c, "tok", tushare.StockBasicRequest{})
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "600519.SH", out[0].TsCode)
	assert.Equal(t, "贵州茅台", out[0].Name)
	assert.Equal(t, "SSE", out[0].Exchange)
}
