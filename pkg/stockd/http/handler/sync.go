package handler

import (
	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

func (h *handler) SyncStocklist(c *gin.Context) {
	token := auth.TushareTokenFor(c)
	n, err := h.stockSvc.SyncFromTushare(c.Request.Context(), token)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"synced": n})
}

func (h *handler) SyncBars(c *gin.Context) {
	if err := h.schedulerSvc.Trigger(c.Request.Context(), "daily-fetch"); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "daily-fetch triggered"})
}

func (h *handler) ImportCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	defer file.Close()
	n, err := h.stockSvc.ImportCSV(c.Request.Context(), file)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"imported": n})
}

func (h *handler) JobStatus(c *gin.Context) {
	name := c.Query("job")
	if name == "" {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrInvalidParam)
		return
	}
	jr, err := h.schedulerSvc.LastRun(c.Request.Context(), name)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, jr)
}
