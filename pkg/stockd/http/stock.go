package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/utils"
)

func (h *handler) SearchStocks(c *gin.Context) {
	q := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := h.svc.SearchStock(c.Request.Context(), q, limit)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}

func (h *handler) GetStock(c *gin.Context) {
	s, err := h.svc.GetStock(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrStockNotFound)
		return
	}
	utils.HTTPRequestSuccess(c, 200, s)
}
