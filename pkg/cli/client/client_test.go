package client_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/cli/client"
	"stock/pkg/models"
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
	var out struct {
		Key string `json:"key"`
	}
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

func TestGET_PortfolioFields(t *testing.T) {
	t.Run("parses camelCase fields", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"requestID":"test","code":200,"message":"ok","data":[{"id":1,"userId":2,"tsCode":"600519.SH","note":"茅台","addedAt":"2025-05-13T10:00:00Z"}]}`))
		}))
		defer srv.Close()

		c := client.New(srv.URL, "stk_test")
		var out []*models.Portfolio
		require.NoError(t, c.GET("/portfolio", &out))
		require.Len(t, out, 1)
		assert.Equal(t, uint(1), out[0].ID)
		assert.Equal(t, uint(2), out[0].UserID)
		assert.Equal(t, "600519.SH", out[0].TsCode)
		assert.Equal(t, "茅台", out[0].Note)
		assert.False(t, out[0].AddedAt.IsZero())
		assert.Equal(t, "2025-05-13T10:00:00Z", out[0].AddedAt.Format(time.RFC3339))
	})
}
