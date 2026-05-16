package services

import (
	"context"
	"time"

	"stock/pkg/models"
)

// isTradingHours reports whether t falls within 09:15–15:00 on a weekday in Asia/Shanghai.
func isTradingHours(t time.Time) bool {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := t.In(loc)
	wd := now.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	h, m, _ := now.Clock()
	total := h*60 + m
	return total >= 9*60+15 && total <= 15*60
}

// GetRealtimeQuote returns the cached quote for tsCode, or fetches it on-demand on a cache miss.
func (s *Service) GetRealtimeQuote(ctx context.Context, tsCode string) (*models.StockRealtimeAndAnalysis, error) {
	s.realtimeMu.RLock()
	q, ok := s.realtimeCache[tsCode]
	s.realtimeMu.RUnlock()
	if ok {
		return q, nil
	}
	s.refreshRealtimeStocks(ctx, []string{tsCode})
	s.realtimeMu.RLock()
	q, ok = s.realtimeCache[tsCode]
	s.realtimeMu.RUnlock()
	if ok {
		return q, nil
	}
	return nil, nil
}

// refreshRealtimeQuotes batch-fetches all portfolio stocks and updates the cache.
func (s *Service) refreshRealtimeQuotes(ctx context.Context) {
	codes, err := s.DistinctTsCodes(ctx)
	if err != nil {
		s.logger.WithError(err).Error("realtime: get portfolio codes failed")
		return
	}
	if len(codes) == 0 {
		return
	}

	s.refreshRealtimeStocks(ctx, codes)
}

func (s *Service) refreshRealtimeStocks(ctx context.Context, codes []string) {
	quotes, err := s.tc.FetchQuotes(ctx, codes)
	if err != nil {
		s.logger.WithError(err).Error("realtime: fetch quotes failed")
		return
	}
	s.fillNames(quotes)
	s.realtimeMu.Lock()
	for _, q := range quotes {
		a := &models.StockRealtimeAndAnalysis{
			StockRealtime: q,
		}
		in := models.AnalysisInput{
			UserID:      0,
			TsCode:      q.TsCode,
			OpenPrice:   &q.Open,
			ActualHigh:  &q.High,
			ActualLow:   &q.Low,
			ActualClose: nil,
		}
		if !isTradingHours(time.Now()) {
			in.ActualClose = &q.Price
		}
		analysisResult, err := s.RunStockAnalysis(ctx, in)
		if err == nil {
			a.StockAnalysisResult = analysisResult
		}
		s.realtimeCache[q.TsCode] = a
	}
	s.realtimeMu.Unlock()
}

// fillNames populates Name from the in-memory stock cache (avoids encoding issues with Tencent response).
func (s *Service) fillNames(quotes []*models.StockRealtime) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	for _, q := range quotes {
		if info, ok := s.stockCacheByTsCode[q.TsCode]; ok {
			q.Name = info.Name
		}
	}
}
