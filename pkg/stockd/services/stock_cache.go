package services

import (
	"context"

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
