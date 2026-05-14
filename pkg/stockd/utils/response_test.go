package utils_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockd/utils"
)

func TestHTTPRequestSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request = c.Request.WithContext(utils.AttachReqID(c.Request.Context()))
	utils.HTTPRequestSuccess(c, 200, gin.H{"key": "val"})
	assert.Equal(t, http.StatusOK, w.Code)
	var r utils.HTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &r))
	assert.Equal(t, 200, r.Code)
	assert.NotEmpty(t, r.RequestID)
}

func TestHTTPRequestFailedV4(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	utils.HTTPRequestFailedV4(c, errors.New("boom"), 500)
	assert.Equal(t, http.StatusOK, w.Code)
	var r utils.HTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &r))
	assert.Equal(t, 500, r.Code)
	assert.NotEmpty(t, r.Message)
}

func TestHTTPRequestFailedV4_Code600(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	utils.HTTPRequestFailedV4(c, errors.New("parse error"), 600)
	var r utils.HTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &r))
	assert.Equal(t, 500, r.Code)
	assert.Contains(t, r.Message, "parse error")
}
