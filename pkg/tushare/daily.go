package tushare

import (
	"context"
	"fmt"

	"stock/pkg/models"
	"stock/pkg/utils"
)

// DailyRequest mirrors Tushare `daily` parameters used by the project.
type DailyRequest struct {
	TsCode    string
	StartDate string // YYYYMMDD
	EndDate   string // YYYYMMDD
}

// Daily fetches OHLCV rows and returns models.StockDailyBar with spreads pre-computed.
// Rows are returned in the order Tushare provides them (newest first); callers
// re-sort as needed.
func Daily(ctx context.Context, c *Client, token string, req DailyRequest) ([]models.StockDailyBar, error) {
	params := map[string]any{
		"ts_code":    utils.ToTushareCode(req.TsCode),
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	}
	resp, err := c.Call(ctx, token, "daily", params,
		"ts_code,trade_date,open,high,low,close,vol,amount")
	if err != nil {
		return nil, err
	}
	idx, err := indexFields(resp.Fields,
		"ts_code", "trade_date", "open", "high", "low", "close", "vol", "amount")
	if err != nil {
		return nil, err
	}
	bars := make([]models.StockDailyBar, 0, len(resp.Items))
	for _, row := range resp.Items {
		tsCode, _ := row[idx["ts_code"]].(string)
		tradeDate, _ := row[idx["trade_date"]].(string)
		open, _ := toFloat(row[idx["open"]])
		high, _ := toFloat(row[idx["high"]])
		low, _ := toFloat(row[idx["low"]])
		close, _ := toFloat(row[idx["close"]])
		vol, _ := toFloat(row[idx["vol"]])
		amount, _ := toFloat(row[idx["amount"]])
		bar := models.StockDailyBar{
			TsCode:    tsCode,
			TradeDate: tradeDate,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Vol:       vol,
			Amount:    amount,
		}
		bar.Spreads = utils.ComputeSpreads(open, high, low, close)
		bars = append(bars, bar)
	}
	return bars, nil
}

func indexFields(fields []string, want ...string) (map[string]int, error) {
	out := make(map[string]int, len(want))
	for i, f := range fields {
		out[f] = i
	}
	for _, w := range want {
		if _, ok := out[w]; !ok {
			return nil, fmt.Errorf("tushare response missing field %q (got %v)", w, fields)
		}
	}
	return out, nil
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case nil:
		return 0, false
	}
	return 0, false
}
