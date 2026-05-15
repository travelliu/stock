package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/services/analysis"
	"stock/pkg/stockd/utils"
)

func (h *handler) GetAnalysis(c *gin.Context) {
	u := auth.User(c)
	in := analysis.Input{UserID: u.ID, TsCode: c.Param(codeValue)}

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
	res, err := h.analysisSvc.Run(c.Request.Context(), in)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, res)
}
