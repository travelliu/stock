package services

import (
	"context"
	"fmt"
	"strings"

	"stock/pkg/models"
)

type StockInfo struct {
	TsCode   string
	Name     string
	Industry string
}

func (s *Service) LoadStockCache(ctx context.Context) error {
	var stocks []models.Stock
	if err := s.db.WithContext(ctx).Find(&stocks).Error; err != nil {
		return err
	}
	byCode := make(map[string]StockInfo, len(stocks))
	byTsCode := make(map[string]StockInfo, len(stocks))
	for _, st := range stocks {
		info := StockInfo{TsCode: st.TsCode, Name: st.Name, Industry: st.Industry}
		byCode[st.Code] = info
		byTsCode[st.TsCode] = info
	}
	s.cacheMu.Lock()
	s.stockCacheByCode = byCode
	s.stockCacheByTsCode = byTsCode
	s.cacheMu.Unlock()
	return nil
}

func (s *Service) ResolveTsCode(code string) (string, error) {
	if strings.Contains(code, ".") {
		s.cacheMu.RLock()
		_, ok := s.stockCacheByTsCode[code]
		s.cacheMu.RUnlock()
		if ok {
			return code, nil
		}
		// Accept full codes even if not in cache (e.g. empty cache scenario).
		return code, nil
	}
	s.cacheMu.RLock()
	info, ok := s.stockCacheByCode[code]
	s.cacheMu.RUnlock()
	if !ok {
		return "", fmt.Errorf("stock not found: %s", code)
	}
	return info.TsCode, nil
}
