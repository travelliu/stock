package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	ContextKeyReqID         ContextKey = "requestID"
	HTTPHeaderNameRequestID            = "X-Request-ID"
)

func genRequestID() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

func GetReqID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyReqID).(string); ok {
		return v
	}
	return ""
}

func AttachReqID(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyReqID, genRequestID())
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := AttachReqID(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetLang(c *gin.Context) string {
	lang := c.GetString("lang")
	if lang == "" {
		lang = c.GetHeader("Lang")
	}
	if lang == "" {
		lang = c.GetHeader("lang")
	}
	if lang == "" {
		lang = ParseAcceptLanguage(c.GetHeader("Accept-Language"))
	}
	return lang
}

func ParseAcceptLanguage(s string) string {
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		q := strings.Split(part, ";")
		if len(q) == 1 {
			return q[0]
		}
		qp := strings.Split(q[1], "=")
		if len(qp) >= 2 && qp[1] == "1" {
			return q[0]
		}
	}
	return ""
}

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("lang", GetLang(c))
		c.Next()
	}
}
