package http

import (
	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/utils"
)

// AuthRequired aborts with ErrUnauthorized when no authenticated user is present.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.User(c) == nil {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminRequired aborts with ErrForbidden when the user is not an admin.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := auth.User(c)
		if u == nil {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrUnauthorized)
			c.Abort()
			return
		}
		if u.Role != "admin" {
			utils.HTTPRequestFailedV4(c, nil, utils.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
