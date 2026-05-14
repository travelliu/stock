package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/services/user"
	"stock/pkg/stockd/utils"
)

func (h *handler) CreateUser(c *gin.Context) {
	var req struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Role         string `json:"role"`
		TushareToken string `json:"tushare_token,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	u, err := h.userSvc.Create(c.Request.Context(), user.CreateInput{
		Username: req.Username, Password: req.Password, Role: req.Role, TushareToken: req.TushareToken,
	})
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, u)
}

func (h *handler) ListUsers(c *gin.Context) {
	list, err := h.userSvc.List(c.Request.Context())
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, list)
}

func (h *handler) PatchUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Role         *string `json:"role,omitempty"`
		Disabled     *bool   `json:"disabled,omitempty"`
		TushareToken *string `json:"tushare_token,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		utils.HTTPRequestFailedV4(c, err, 600)
		return
	}
	if req.Role != nil {
		_ = h.userSvc.SetRole(c.Request.Context(), uint(id), *req.Role)
	}
	if req.Disabled != nil {
		_ = h.userSvc.SetDisabled(c.Request.Context(), uint(id), *req.Disabled)
	}
	if req.TushareToken != nil {
		_ = h.userSvc.SetTushareToken(c.Request.Context(), uint(id), *req.TushareToken)
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, gin.H{"message": "updated"})
}

func (h *handler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.userSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, http.StatusOK, gin.H{"message": "deleted"})
}
