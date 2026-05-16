package baidu_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"stock/pkg/baidu"
)

func TestFetchConceptBlocks(t *testing.T) {
	payload := map[string]any{
		"ResultCode": "0",
		"Result": []map[string]any{
			{
				"type": "行业板块",
				"list": []map[string]any{
					{"name": "工业机器人", "increase": "3.5"},
				},
			},
			{
				"type": "概念板块",
				"list": []map[string]any{
					{"name": "人形机器人", "increase": "5.1"},
					{"name": "减速器", "increase": "4.2"},
				},
			},
			{
				"type": "地域板块",
				"list": []map[string]any{
					{"name": "上海", "increase": "1.2"},
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "688017")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := baidu.NewClient(baidu.WithBaseURL(srv.URL))
	blocks, err := c.FetchConceptBlocks(context.Background(), "688017")
	require.NoError(t, err)

	require.Len(t, blocks.Industry, 1)
	assert.Equal(t, "工业机器人", blocks.Industry[0].Name)

	require.Len(t, blocks.Concept, 2)
	assert.Equal(t, "人形机器人", blocks.Concept[0].Name)

	require.Len(t, blocks.Region, 1)
	assert.Equal(t, "上海", blocks.Region[0].Name)

	assert.Equal(t, []string{"人形机器人", "减速器"}, blocks.ConceptTags)
}

func TestFetchConceptBlocks_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ResultCode": -1})
	}))
	defer srv.Close()

	c := baidu.NewClient(baidu.WithBaseURL(srv.URL))
	_, err := c.FetchConceptBlocks(context.Background(), "688017")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestFetchFundFlowHistory(t *testing.T) {
	payload := map[string]any{
		"ResultCode": "0",
		"Result": map[string]any{
			"list": []map[string]any{
				{
					"showtime":    "2026-05-16",
					"closepx":     "224.12",
					"ratio":       "4.24",
					"superNetIn":  "5000",
					"largeNetIn":  "3000",
					"mediumNetIn": "-1000",
					"littleNetIn": "-7000",
					"extMainIn":   "8000",
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "688017")
		assert.Contains(t, r.URL.String(), "rn=20")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := baidu.NewClient(baidu.WithBaseURL(srv.URL))
	rows, err := c.FetchFundFlowHistory(context.Background(), "688017", 20)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "2026-05-16", rows[0].Date)
	assert.Equal(t, "224.12", rows[0].Close)
	assert.Equal(t, "8000", rows[0].MainIn)
}

func TestFetchFundFlowHistory_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ResultCode": 1})
	}))
	defer srv.Close()

	c := baidu.NewClient(baidu.WithBaseURL(srv.URL))
	_, err := c.FetchFundFlowHistory(context.Background(), "688017", 20)
	require.Error(t, err)
}
