package http

import (
	"stock/pkg/models"
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

func (h *handler) GetAnalysis(c *gin.Context) {
	u := auth.User(c)
	in := models.AnalysisInput{UserID: u.ID, TsCode: c.Param(codeValue)}

	if v := c.Query("actual_open"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.OpenPrice = &f
	}
	if v := c.Query("actual_high"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.ActualHigh = &f
	}
	if v := c.Query("actual_low"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.ActualLow = &f
	}
	if v := c.Query("actual_close"); v != "" {
		f, _ := strconv.ParseFloat(v, 64)
		in.ActualClose = &f
	}
	res, err := h.svc.RunStockAnalysis(c.Request.Context(), in)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, res)
}
