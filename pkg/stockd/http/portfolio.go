package http

import (
	"stock/pkg/models"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

const (
	tsCodeValue = "tsCode"
	tsCodeUrl   = ":tsCode"
)

func (h *handler) ListPortfolio(c *gin.Context) {
	u := auth.User(c)
	list, err := h.portfolioSvc.List(c.Request.Context(), u.ID)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}

func (h *handler) AddPortfolio(c *gin.Context) {
	var req models.PortfolioReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.portfolioSvc.Add(c.Request.Context(), u.ID, req.TsCode, req.Note); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "added"})
}

func (h *handler) RemovePortfolio(c *gin.Context) {
	u := auth.User(c)
	if err := h.portfolioSvc.Remove(c.Request.Context(), u.ID, c.Param(tsCodeValue)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "removed"})
}

func (h *handler) UpdatePortfolioNote(c *gin.Context) {
	var req models.PortfolioReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.portfolioSvc.UpdateNote(c.Request.Context(), u.ID, c.Param(tsCodeValue), req.Note); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "updated"})
}
