package services

import (
	"context"
	"fmt"
	"strings"
	
	"stock/pkg/models"
)

func (s *Service) LoadStockCache(ctx context.Context) error {
	var stocks []*models.Stock
	if err := s.db.WithContext(ctx).Find(&stocks).Error; err != nil {
		return err
	}
	byCode := make(map[string]*models.Stock, len(stocks))
	byTsCode := make(map[string]*models.Stock, len(stocks))
	for _, st := range stocks {
		byCode[st.Code] = st
		byTsCode[st.TsCode] = st
	}
	s.cacheMu.Lock()
	s.stockCacheByCode = byCode
	s.stockCacheByTsCode = byTsCode
	s.cacheMu.Unlock()
	return nil
}

func (s *Service) ResolveTsCode(code string) (string, error) {
	if strings.Contains(code, ".") {
		return code, nil
	}
	s.cacheMu.RLock()
	st, ok := s.stockCacheByCode[code]
	s.cacheMu.RUnlock()
	if !ok {
		return "", fmt.Errorf("stock %q not found in cache", code)
	}
	return st.TsCode, nil
}
