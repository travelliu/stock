package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/services/token"
	"stock/pkg/stockd/utils"
)

func (h *handler) ListTokens(c *gin.Context) {
	u := auth.User(c)
	list, err := h.tokenSvc.List(c.Request.Context(), u.ID)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, list)
}

func (h *handler) IssueToken(c *gin.Context) {
	var req struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	plain, tok, err := h.tokenSvc.Issue(c.Request.Context(), token.IssueInput{
		UserID: u.ID, Name: req.Name, ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"token": plain, "metadata": tok})
}

func (h *handler) RevokeToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	u := auth.User(c)
	if err := h.tokenSvc.Revoke(c.Request.Context(), u.ID, uint(id)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "revoked"})
}

func (h *handler) SetTushareToken(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.userSvc.SetTushareToken(c.Request.Context(), u.ID, req.Token); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "updated"})
}

func (h *handler) ChangePassword(c *gin.Context) {
	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u := auth.User(c)
	if err := h.userSvc.ChangePassword(c.Request.Context(), u.ID, req.Old, req.New); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, gin.H{"message": "password changed"})
}
