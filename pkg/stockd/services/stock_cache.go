package services

import (
	"context"

	"stock/pkg/models"
)

func (s *Service) LoadStockCache(ctx context.Context) error {
	var stocks []*models.StockBasicInfo
	if err := s.db.WithContext(ctx).Find(&stocks).Error; err != nil {
		return err
	}
	byCode := make(map[string]*models.StockBasicInfo, len(stocks))
	byTsCode := make(map[string]*models.StockBasicInfo, len(stocks))
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
