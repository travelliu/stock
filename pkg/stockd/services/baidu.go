package services

import (
	"context"
	"fmt"

	"stock/pkg/baidu"
)

// GetConceptBlocks returns concept/industry/region block memberships for a stock.
func (s *Service) GetConceptBlocks(ctx context.Context, code string) (*baidu.ConceptBlocks, error) {
	if s.baiduClient == nil {
		return nil, fmt.Errorf("baidu client not configured")
	}
	return s.baiduClient.FetchConceptBlocks(ctx, code)
}

// GetFundFlowHistory returns daily fund flow for the last 20 trading days.
func (s *Service) GetFundFlowHistory(ctx context.Context, code string) ([]baidu.FundFlowDay, error) {
	if s.baiduClient == nil {
		return nil, fmt.Errorf("baidu client not configured")
	}
	return s.baiduClient.FetchFundFlowHistory(ctx, code, 20)
}
