package tushare_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/pkg/tushare"
)

func TestClient_PostJSON_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "daily", req["api_name"])
		assert.Equal(t, "tok-123", req["token"])
		_, _ = io.WriteString(w, `{"code":0,"msg":"","data":{"fields":["a","b"],"items":[[1,2]]}}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL), tushare.WithTimeout(2*time.Second))
	resp, err := c.Call(context.Background(), "tok-123", "daily", map[string]any{"ts_code": "600000.SH"}, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, resp.Fields)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, []any{float64(1), float64(2)}, resp.Items[0])
}

func TestClient_RetriesOn5xxThenFails(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := tushare.NewClient(
		tushare.WithBaseURL(srv.URL),
		tushare.WithTimeout(500*time.Millisecond),
		tushare.WithMaxRetries(2),
		tushare.WithRetryDelay(1*time.Millisecond),
	)
	_, err := c.Call(context.Background(), "tok", "daily", nil, "")
	require.Error(t, err)
	assert.Equal(t, 3, calls, "1 try + 2 retries")
	assert.Contains(t, err.Error(), "500")
}

func TestClient_APIErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":40203,"msg":"token invalid","data":null}`)
	}))
	defer srv.Close()

	c := tushare.NewClient(tushare.WithBaseURL(srv.URL))
	_, err := c.Call(context.Background(), "bad", "daily", nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token invalid")
}

func TestClient_DefaultBaseURL(t *testing.T) {
	c := tushare.NewClient()
	assert.True(t, strings.HasPrefix(c.BaseURL(), "http"))
}
