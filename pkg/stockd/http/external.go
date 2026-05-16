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

// GetFundFlow returns industry-level fund flow (申万一级/二级) for a stock (Baidu PAE).
func (h *handler) GetFundFlow(c *gin.Context) {
	data, err := h.svc.GetFundFlow(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}
