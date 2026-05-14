package tushare

import "context"

// StockBasicRow mirrors the columns used by the catalog.
type StockBasicRow struct {
	TsCode   string
	Symbol   string
	Name     string
	Area     string
	Industry string
	Market   string
	Exchange string
	ListDate string
}

// StockBasicRequest is the (optional) filter set used in this project.
type StockBasicRequest struct {
	TsCode     string
	ListStatus string // "L" listed, "D" delisted, "P" paused. Empty = all listed.
}

// StockBasic returns the catalog rows. Pagination should be added if/when
// Tushare starts capping a single response (currently 5000 rows).
func StockBasic(ctx context.Context, c *Client, token string, req StockBasicRequest) ([]StockBasicRow, error) {
	params := map[string]any{}
	if req.TsCode != "" {
		params["ts_code"] = req.TsCode
	}
	if req.ListStatus != "" {
		params["list_status"] = req.ListStatus
	}
	resp, err := c.Call(ctx, token, "stock_basic", params,
		"ts_code,symbol,name,area,industry,market,exchange,list_date")
	if err != nil {
		return nil, err
	}
	idx, err := indexFields(resp.Fields,
		"ts_code", "symbol", "name", "area", "industry", "market", "exchange", "list_date")
	if err != nil {
		return nil, err
	}
	out := make([]StockBasicRow, 0, len(resp.Items))
	for _, row := range resp.Items {
		out = append(out, StockBasicRow{
			TsCode:   strOrEmpty(row[idx["ts_code"]]),
			Symbol:   strOrEmpty(row[idx["symbol"]]),
			Name:     strOrEmpty(row[idx["name"]]),
			Area:     strOrEmpty(row[idx["area"]]),
			Industry: strOrEmpty(row[idx["industry"]]),
			Market:   strOrEmpty(row[idx["market"]]),
			Exchange: strOrEmpty(row[idx["exchange"]]),
			ListDate: strOrEmpty(row[idx["list_date"]]),
		})
	}
	return out, nil
}

func strOrEmpty(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
