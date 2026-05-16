package baidu_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stock/pkg/baidu"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchConceptBlocks(t *testing.T) {
	payload := map[string]any{
		"ResultCode": "0",
		"Result": map[string]any{
			"688017": []map[string]any{
				{
					"name": "行业",
					"list": []map[string]any{
						{"name": "工业机器人", "ratio": "3.5%", "describe": "申万一级"},
					},
				},
				{
					"name": "概念",
					"list": []map[string]any{
						{"name": "人形机器人", "ratio": "5.1%"},
						{"name": "减速器", "ratio": "4.2%"},
					},
				},
				{
					"name": "地域",
					"list": []map[string]any{
						{"name": "上海", "ratio": "1.2%"},
					},
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
	assert.Equal(t, "申万一级", blocks.Industry[0].Describe)

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

func TestFetchFundFlow(t *testing.T) {
	payload := map[string]any{
		"ResultCode": "0",
		"Result": map[string]any{
			"content": map[string]any{
				"fundFlowBlock": map[string]any{
					"result": []map[string]any{
						{
							"belongs":    "stocklevelone",
							"updateTime": "2026-05-16 15:30:00",
							"unit":       "亿",
							"industry":   map[string]any{"name": "电力设备", "desc": "申万一级"},
							"todayMainFlow": map[string]any{
								"mainIn": "1.23", "mainOut": "0.98", "mainNetIn": "0.25",
							},
							"superGrp": map[string]any{
								"netTurnover": "0.12", "turnoverIn": "0.20", "turnoverOut": "0.08",
								"turnoverInRate": "3.5%", "turnoverOutRate": "2.1%",
							},
							"largeGrp": map[string]any{
								"netTurnover": "0.08", "turnoverIn": "0.15", "turnoverOut": "0.07",
								"turnoverInRate": "2.0%", "turnoverOutRate": "1.8%",
							},
							"mediumGrp": map[string]any{
								"netTurnover": "-0.03", "turnoverIn": "0.10", "turnoverOut": "0.13",
								"turnoverInRate": "1.5%", "turnoverOutRate": "1.8%",
							},
							"littleGrp": map[string]any{
								"netTurnover": "-0.06", "turnoverIn": "0.05", "turnoverOut": "0.11",
								"turnoverInRate": "0.8%", "turnoverOutRate": "1.5%",
							},
							"recently": []map[string]any{
								{"key": "近三日", "value": "0.5"},
								{"key": "近五日", "value": "-0.3"},
							},
						},
					},
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
	ff, err := c.FetchFundFlow(context.Background(), "688017")
	require.NoError(t, err)
	require.Len(t, ff.Levels, 1)

	lvl := ff.Levels[0]
	assert.Equal(t, "stocklevelone", lvl.Belongs)
	assert.Equal(t, "电力设备", lvl.Industry.Name)
	assert.Equal(t, "申万一级", lvl.Industry.Desc)
	assert.Equal(t, "0.25", lvl.TodayMain.MainNetIn)
	assert.Equal(t, "0.12", lvl.Super.NetTurnover)
	require.Len(t, lvl.Recently, 2)
	assert.Equal(t, "近三日", lvl.Recently[0].Key)
}

func TestFetchFundFlow_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ResultCode": 1})
	}))
	defer srv.Close()

	c := baidu.NewClient(baidu.WithBaseURL(srv.URL))
	_, err := c.FetchFundFlow(context.Background(), "688017")
	require.Error(t, err)
}
