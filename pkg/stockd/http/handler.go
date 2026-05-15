// Package handler implements HTTP handlers for the stockd API.
package http

import (
	"stock/pkg/stockd/services"
	"stock/pkg/stockd/services/analysis"
)

type handler struct {
	svc         *services.Service
	analysisSvc *analysis.Service
}

func NewHandler(
	svc *services.Service,
	analysisSvc *analysis.Service,
) *handler {
	return &handler{
		svc:         svc,
		analysisSvc: analysisSvc,
	}
}
