// Package http implements HTTP handlers for the stockd API.
package http

import (
	"stock/pkg/stockd/services"
)

type handler struct {
	svc *services.Service
}

func NewHandler(
	svc *services.Service,
) *handler {
	return &handler{
		svc: svc,
	}
}
