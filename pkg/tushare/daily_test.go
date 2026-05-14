package tushare_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/tushare"
)

func TestDaily_ParsesItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "daily", req["api_name"])
		params := req["params"].(map[string]any)
		assert.Equal(t, "600519.SH", params["ts_code"])

		_, _ = io.WriteString(w, `{
		  "code":0,"msg":"","data":{
		    "fields":["ts_code","trade_date","open","high","low","close","vol","amount"],
		    "items":[
		      ["600519.SH","20250513",1620.0,1655.0,1601.0,1632.0,3500.0,500000.0],
		      ["600519.SH","20250512",1610.0,1640.0,1590.0,1620.0,3300.0,480000.0]
		    ]
		  }
		}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL))
	bars, err := tushare.Daily(context.Background(), c, "tok", tushare.DailyRequest{
		TsCode:    "600519.SH",
		StartDate: "20250101",
		EndDate:   "20250513",
	})
	require.NoError(t, err)
	require.Len(t, bars, 2)
	assert.Equal(t, "20250513", bars[0].TradeDate)
	assert.InDelta(t, 1620.0, bars[0].Open, 1e-9)
	assert.InDelta(t, 35.0, bars[0].Spreads.OH, 1e-9) // |1655 - 1620|
}
