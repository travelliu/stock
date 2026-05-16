//go:build wireable

package services

import (
	"context"
	"fmt"

	"stock/pkg/models"
)

// GetRealtimeQuote returns the cached quote for tsCode, or fetches it on-demand on a cache miss.
func (s *Service) GetRealtimeQuote(ctx context.Context, tsCode string) (*models.RealtimeQuote, error) {
	s.realtimeMu.RLock()
	q, ok := s.realtimeCache[tsCode]
	s.realtimeMu.RUnlock()
	if ok {
		return q, nil
	}

	quotes, err := s.tc.FetchQuotes(ctx, []string{tsCode})
	if err != nil {
		return nil, fmt.Errorf("fetch quote %s: %w", tsCode, err)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote data for %s", tsCode)
	}
	s.fillNames(quotes)
	s.realtimeMu.Lock()
	for _, q := range quotes {
		s.realtimeCache[q.TsCode] = q
	}
	s.realtimeMu.Unlock()
	return quotes[0], nil
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
	quotes, err := s.tc.FetchQuotes(ctx, codes)
	if err != nil {
		s.logger.WithError(err).Error("realtime: fetch quotes failed")
		return
	}
	s.fillNames(quotes)
	s.realtimeMu.Lock()
	for _, q := range quotes {
		s.realtimeCache[q.TsCode] = q
	}
	s.realtimeMu.Unlock()
}

// fillNames populates Name from the in-memory stock cache (avoids encoding issues with Tencent response).
func (s *Service) fillNames(quotes []*models.RealtimeQuote) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	for _, q := range quotes {
		if info, ok := s.stockCacheByTsCode[q.TsCode]; ok {
			q.Name = info.Name
		}
	}
}
