package http

import (
	"stock/pkg/models"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

const (
	codeValue = "code"
	codeUrl   = ":code"
)

func (h *handler) ListPortfolio(c *gin.Context) {
	u := auth.User(c)
	list, err := h.svc.ListPortfolio(c.Request.Context(), u.ID)
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
	tsCode, err := h.svc.ResolveTsCode(req.GetCode())
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrStockNotFound)
		return
	}
	u := auth.User(c)
	if err := h.svc.AddPortfolio(c.Request.Context(), u.ID, tsCode, req.Note); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "added"})
}

func (h *handler) RemovePortfolio(c *gin.Context) {
	tsCode, err := h.svc.ResolveTsCode(c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrStockNotFound)
		return
	}
	u := auth.User(c)
	if err := h.svc.RemovePortfolio(c.Request.Context(), u.ID, tsCode); err != nil {
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
	tsCode, err := h.svc.ResolveTsCode(c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrStockNotFound)
		return
	}
	u := auth.User(c)
	if err := h.svc.UpdatePortfolioNote(c.Request.Context(), u.ID, tsCode, req.Note); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "updated"})
}
