package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"stock/pkg/models"
)

// ImportCSV streams a `stock_basic`-shaped CSV (with header row) and upserts.
// Header columns: ts_code,symbol,name,area,industry,market,exchange,list_date
func (s *Service) ImportCSV(ctx context.Context, r io.Reader) (int, error) {
	cr := csv.NewReader(r)
	header, err := cr.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}
	required := []string{"ts_code", "symbol", "name"}
	for _, k := range required {
		if _, ok := idx[k]; !ok {
			return 0, fmt.Errorf("missing column %q", k)
		}
	}

	n := 0
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return n, err
		}
		ts := row[idx["ts_code"]]
		st := &models.StockBasicInfo{
			TsCode:   ts,
			Code:     row[idx["symbol"]],
			Name:     row[idx["name"]],
			Area:     get(row, idx, "area"),
			Industry: get(row, idx, "industry"),
			Market:   get(row, idx, "market"),
			Exchange: get(row, idx, "exchange"),
			ListDate: get(row, idx, "list_date"),
		}
		if err := s.db.WithContext(ctx).Save(st).Error; err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func get(row []string, idx map[string]int, key string) string {
	if i, ok := idx[key]; ok && i < len(row) {
		return row[i]
	}
	return ""
}
