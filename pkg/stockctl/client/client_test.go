package client_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/stockctl/client"
)

func TestGET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer stk_test", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"requestID":"test-123","code":200,"message":"成功","data":{"key":"val"}}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "stk_test")
	var out struct{ Key string `json:"key"` }
	require.NoError(t, c.GET("/test", &out))
	assert.Equal(t, "val", out.Key)
}

func TestGET_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"requestID":"test-456","code":40002,"message":"未认证","data":null}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "")
	err := c.GET("/test", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "未认证")
}
