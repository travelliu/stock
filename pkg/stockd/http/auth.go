package http

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

func (h *handler) Login(c *gin.Context) {
	var req models.LoginReq
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u, err := h.userSvc.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	sess := sessions.Default(c)
	sess.Set("uid", u.ID)
	_ = sess.Save()
	utils.HTTPRequestSuccess(c, http.StatusOK, u)
}

func (h *handler) Logout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()
	utils.HTTPRequestSuccess(c, http.StatusOK, gin.H{"message": "logged out"})
}

func (h *handler) Me(c *gin.Context) {
	u := auth.User(c)
	if u == nil {
		utils.HTTPRequestFailedV4(c, nil, utils.ErrUnauthorized)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, u)
}
