package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/services/draft"
	"stock/pkg/stockd/utils"
)

func (h *handler) GetDraftToday(c *gin.Context) {
	u := auth.User(c)
	tradeDate := c.DefaultQuery("trade_date", time.Now().Format("20060102"))
	d, err := h.draftSvc.GetByDate(c.Request.Context(), u.ID, c.Query("ts_code"), tradeDate)
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrInvalidParam)
		return
	}
	utils.HTTPRequestSuccess(c, 200, d)
}

func (h *handler) UpsertDraft(c *gin.Context) {
	var req struct {
		TsCode    string   `json:"ts_code"`
		TradeDate string   `json:"trade_date"`
		Open      *float64 `json:"open,omitempty"`
		High      *float64 `json:"high,omitempty"`
		Low       *float64 `json:"low,omitempty"`
		Close     *float64 `json:"close,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	d, err := h.draftSvc.Upsert(c.Request.Context(), draft.UpsertInput{
		UserID: u.ID, TsCode: req.TsCode, TradeDate: req.TradeDate,
		Open: req.Open, High: req.High, Low: req.Low, Close: req.Close,
	})
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, d)
}

func (h *handler) DeleteDraft(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	u := auth.User(c)
	if err := h.draftSvc.Delete(c.Request.Context(), u.ID, uint(id)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "deleted"})
}
