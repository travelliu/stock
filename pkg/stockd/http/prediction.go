package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/utils"
)

func (h *handler) RecalcPredictions(c *gin.Context) {
	res, err := h.svc.Recalc(c.Request.Context(), c.DefaultQuery("code", ""))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, res)
}

func (h *handler) ListPredictions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.ListPredictionsPage(c.Request.Context(), c.Param(codeValue), c.Query("from"), c.Query("to"), page, limit)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, result)
}
