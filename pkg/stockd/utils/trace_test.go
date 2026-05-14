package utils_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"stock/pkg/stockd/utils"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(utils.RequestID())
	r.GET("/", func(c *gin.Context) {
		assert.NotEmpty(t, utils.GetReqID(c.Request.Context()))
		c.Status(200)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, 200, w.Code)
}

func TestParseAcceptLanguage(t *testing.T) {
	assert.Equal(t, "zh-CN", utils.ParseAcceptLanguage("zh-CN,en;q=0.9"))
	assert.Equal(t, "en", utils.ParseAcceptLanguage("en"))
	assert.Equal(t, "", utils.ParseAcceptLanguage(""))
}
