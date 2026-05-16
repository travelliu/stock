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

// GetFundFlow returns industry-level fund flow (申万一级/二级) for a stock.
func (s *Service) GetFundFlow(ctx context.Context, code string) (*baidu.FundFlow, error) {
	if s.baiduClient == nil {
		return nil, fmt.Errorf("baidu client not configured")
	}
	return s.baiduClient.FetchFundFlow(ctx, code)
}
