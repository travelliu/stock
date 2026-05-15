package http

import (
	"github.com/gin-gonic/gin"
	
	"stock/pkg/stockd/utils"
)

func (h *handler) QueryBars(c *gin.Context) {
	list, err := h.svc.QueryStockDailyBar(c.Request.Context(), c.Param("tsCode"), c.Query("from"), c.Query("to"))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}
