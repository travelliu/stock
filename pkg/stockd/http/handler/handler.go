// Package handler implements HTTP handlers for the stockd API.
package handler

import (
	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/services/bars"
	"stock/pkg/stockd/services/draft"
	"stock/pkg/stockd/services/portfolio"
	"stock/pkg/stockd/services/scheduler"
	"stock/pkg/stockd/services/stock"
	"stock/pkg/stockd/services/token"
	"stock/pkg/stockd/services/user"
)

type handler struct {
	userSvc      *user.Service
	tokenSvc     *token.Service
	stockSvc     *stock.Service
	portfolioSvc *portfolio.Service
	draftSvc     *draft.Service
	barsSvc      *bars.Service
	analysisSvc  *analysis.Service
	schedulerSvc *scheduler.Service
}

func NewHandler(
	userSvc *user.Service,
	tokenSvc *token.Service,
	stockSvc *stock.Service,
	portfolioSvc *portfolio.Service,
	draftSvc *draft.Service,
	barsSvc *bars.Service,
	analysisSvc *analysis.Service,
	schedulerSvc *scheduler.Service,
) *handler {
	return &handler{
		userSvc: userSvc, tokenSvc: tokenSvc, stockSvc: stockSvc,
		portfolioSvc: portfolioSvc, draftSvc: draftSvc, barsSvc: barsSvc,
		analysisSvc: analysisSvc, schedulerSvc: schedulerSvc,
	}
}
