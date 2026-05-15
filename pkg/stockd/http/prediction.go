package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/utils"
)

func (h *handler) RecalcPredictions(c *gin.Context) {
	tsCode := c.DefaultQuery("ts_code", "")
	res, err := h.svc.Recalc(c.Request.Context(), tsCode)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, res)
}

func (h *handler) ListPredictions(c *gin.Context) {
	tsCode := c.Param("tsCode")
	from := c.Query("from")
	to := c.Query("to")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	preds, err := h.svc.ListAnalysisPrediction(c.Request.Context(), tsCode, from, to, limit)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, preds)
}
