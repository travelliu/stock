package utils

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				if !brokenPipe {
					logrus.Errorf("[Recovery] %s panic recovered:\n%v",
						time.Now().Format("2006/01/02 - 15:04:05"), err)
				}
				c.JSON(http.StatusInternalServerError, HTTPResponse{Code: ERROR, Message: fmt.Sprintf("%v", err)})
				c.Abort()
			}
		}()
		c.Next()
	}
}
