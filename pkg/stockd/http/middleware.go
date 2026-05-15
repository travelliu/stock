package http

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

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

// Logger 日志记录到文件
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		// 打印请求开始时间
		logFields := logrus.Fields{
			// 请求IP
			"clientIp": c.ClientIP(),
			// 请求方式
			"reqMethod": c.Request.Method,
			// 请求路由
			"reqUri":                      c.Request.RequestURI,
			string(utils.ContextKeyReqID): utils.GetReqID(c.Request.Context()),
			"userID":                      c.GetInt64("userID"),
			"clientUserAgent":             c.Request.UserAgent(),
		}
		logger.WithFields(logFields).Info("http Request begin")
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		bodyStr := string(body)
		// req body
		bodyStr = pdReg.ReplaceAllString(bodyStr, "${1}\"*****\"")
		logger.WithFields(logFields).Debugf("reqbody [%s]", bodyStr)
		// 处理请求
		c.Next()
		// 结束时间
		endTime := time.Now()
		// 状态码
		logFields["statusCode"] = c.Writer.Status()
		// 执行时间
		logFields["latencyTime"] = endTime.Sub(startTime)
		logger.WithFields(logFields).Info("http Request end")
	}
}

// 脱敏密码表达式
var pdReg = regexp.MustCompile(`(?i)(\"password\":\s*)\"(?:.*?)\"`)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

// Recovery returns a middleware for a given writer that recovers from any panics and calls the provided handle func to handle it.
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				if logger != nil {
					stack := utils.Stack(3)
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}
					headersToStr := strings.Join(headers, "\r\n")
					if brokenPipe {
						logger.Errorf("%s\n%s", err, headersToStr)
					} else {
						logger.Errorf("[Recovery] %s panic recovered:\n%s\n%s",
							TimeFormat(time.Now()), err, stack)
					}
				}
				c.JSON(http.StatusInternalServerError, utils.HTTPResponse{Code: utils.ERROR, Message: err.(error).Error()})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// TimeFormat returns a customized time string for logger.
func TimeFormat(t time.Time) string {
	return t.Format("2006/01/02 - 15:04:05")
}
