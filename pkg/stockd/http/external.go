package http

import (
	"github.com/gin-gonic/gin"
	"stock/pkg/stockd/utils"
)

// GetConceptBlocks returns concept/industry/region block memberships (Baidu PAE).
func (h *handler) GetConceptBlocks(c *gin.Context) {
	data, err := h.svc.GetConceptBlocks(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}

// GetFundFlow returns per-stock fund flow history for the last 20 trading days (Baidu PAE).
func (h *handler) GetFundFlow(c *gin.Context) {
	data, err := h.svc.GetFundFlowHistory(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}
