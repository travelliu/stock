package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/utils"
)

func (h *handler) RecalcPredictions(c *gin.Context) {
	code := c.DefaultQuery("code", "")
	tsCode := ""
	if code != "" {
		var err error
		tsCode, err = h.svc.ResolveTsCode(code)
		if err != nil {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrStockNotFound)
			return
		}
	}
	res, err := h.svc.Recalc(c.Request.Context(), tsCode)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, res)
}

func (h *handler) ListPredictions(c *gin.Context) {
	tsCode, err := h.svc.ResolveTsCode(c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrStockNotFound)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.ListPredictionsPage(c.Request.Context(), tsCode, c.Query("from"), c.Query("to"), page, limit)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, result)
}
