package http

import (
	"stock/pkg/stockd/services"
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

func (h *handler) ListTokens(c *gin.Context) {
	u := auth.User(c)
	list, err := h.svc.ListTokens(c.Request.Context(), u.ID)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}

func (h *handler) IssueToken(c *gin.Context) {
	var req models.IssueTokenReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	plain, tok, err := h.svc.Issue(c.Request.Context(), services.IssueInput{
		UserID: u.ID, Name: req.Name, ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, models.IssueTokenResp{
		Token: plain,
		Metadata: &models.APIToken{
			ID: tok.ID, UserID: tok.UserID, Name: tok.Name,
			LastUsedAt: tok.LastUsedAt, ExpiresAt: tok.ExpiresAt, CreatedAt: tok.CreatedAt,
		},
	})
}

func (h *handler) RevokeToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	u := auth.User(c)
	if err := h.svc.RevokeToken(c.Request.Context(), u.ID, uint(id)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "revoked"})
}

func (h *handler) SetTushareToken(c *gin.Context) {
	var req models.SetTushareTokenReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.svc.SetUserTushareToken(c.Request.Context(), u.ID, req.Token); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "updated"})
}

func (h *handler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.svc.ChangePassword(c.Request.Context(), u.ID, req.Old, req.New); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "password changed"})
}
