package http

import (
	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/utils"
)

func (h *handler) GetQuote(c *gin.Context) {
	q, err := h.svc.GetRealtimeQuote(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, q)
}
