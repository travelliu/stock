# 个股 Phase 4: 估值层（EPS 预测 + 正向 PE / PEG）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 接通两个个股端点：机构一致预期 EPS（同花顺 THS API），以及基于现价和 EPS 计算的前瞻 PE / PEG / PE 消化年数。

**Architecture:** 新建 `pkg/ths/` 客户端包（EPS forecast 方法），Service 注入（ServiceOption 模式），EPS 端点直通，valuation 端点在 service 层聚合（调用腾讯现价 + THS EPS）并计算。

**Prerequisite:** Phase 1 已完成（ServiceOption 模式已在 service.go 中）。Phase 1 注入的腾讯客户端已提供实时现价。

**Tech Stack:** Go 1.25，标准库 `net/http`/`math`，Gin，现有 `pkg/stockd/utils`。

**API 说明:** 同花顺 THS API URL 在实现前需用浏览器 DevTools 核实。推测路径为
`https://basic.10jqka.com.cn/api/stock/profit.json?code={code}`，字段 `yoyjlr`（净利润预测）或专用预期 EPS 字段。实现前请对照核实。

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `pkg/ths/client.go` | Client + EpsForecast 结构体 + FetchEpsForecast |
| Create | `pkg/ths/client_test.go` | 2 个单元测试 |
| Modify | `pkg/stockd/services/service.go` | 加 `thsClient *ths.Client` + `WithThsClient` option |
| Create | `pkg/stockd/services/ths.go` | `GetEpsForecast` / `GetValuation` 服务方法 |
| Modify | `pkg/stockd/http/external.go` | 追加两个 handler |
| Modify | `pkg/stockd/http/router.go` | 注册新路由 |
| Modify | `cmd/stockd/main.go` | 实例化 `ths.Client` |

---

## Task 1: pkg/ths — Client + EpsForecast

**Files:**
- Create: `pkg/ths/client.go`
- Create: `pkg/ths/client_test.go`

- [ ] **Step 1: 创建 `pkg/ths/client.go`**

```go
// Package ths fetches institutional EPS forecasts from Tonghuashun (10jqka.com.cn).
package ths

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://basic.10jqka.com.cn"

var defaultHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/117.0.0.0 Safari/537.36",
	"Referer":    "https://basic.10jqka.com.cn/",
}

// Client calls Tonghuashun APIs.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures Client.
type Option func(*Client)

// WithBaseURL overrides the base URL (used in tests).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = u } }

// NewClient returns a Client with sensible defaults.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// EpsYear is institutional consensus EPS for one fiscal year.
type EpsYear struct {
	Year  string  `json:"year"`
	Count int     `json:"count"`
	Min   float64 `json:"min"`
	Mean  float64 `json:"mean"`
	Max   float64 `json:"max"`
}

// EpsForecast groups consensus EPS forecasts by year (usually current + next 2 years).
type EpsForecast struct {
	Code  string    `json:"code"`
	Years []EpsYear `json:"years"`
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("ths: build request: %w", err)
	}
	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ths: request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ths: read body: %w", err)
	}
	return body, nil
}

// FetchEpsForecast returns institutional consensus EPS forecasts for a stock.
// code is a 6-digit stock code.
// NOTE: verify the API URL and response shape in DevTools before implementation.
func (c *Client) FetchEpsForecast(ctx context.Context, code string) (*EpsForecast, error) {
	path := fmt.Sprintf("/api/stock/profit.json?code=%s", code)
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	// Expected shape (verify before implementing):
	// {"status": 0, "data": {"forecast": [{"year":"2026","count":15,"min":1.2,"avg":1.5,"max":1.8}, ...]}}
	var raw struct {
		Status int `json:"status"`
		Data   struct {
			Forecast []struct {
				Year  string  `json:"year"`
				Count int     `json:"count"`
				Min   float64 `json:"min"`
				Avg   float64 `json:"avg"`
				Max   float64 `json:"max"`
			} `json:"forecast"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("ths: parse eps forecast: %w", err)
	}
	if raw.Status != 0 {
		return nil, fmt.Errorf("ths: API error status=%d", raw.Status)
	}

	out := &EpsForecast{Code: code}
	for _, f := range raw.Data.Forecast {
		out.Years = append(out.Years, EpsYear{
			Year:  f.Year,
			Count: f.Count,
			Min:   f.Min,
			Mean:  f.Avg,
			Max:   f.Max,
		})
	}
	return out, nil
}
```

- [ ] **Step 2: 创建 `pkg/ths/client_test.go`**

```go
package ths_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"stock/pkg/ths"
)

func TestFetchEpsForecast(t *testing.T) {
	payload := map[string]any{
		"status": 0,
		"data": map[string]any{
			"forecast": []map[string]any{
				{"year": "2026", "count": 15, "min": 1.2, "avg": 1.5, "max": 1.8},
				{"year": "2027", "count": 12, "min": 1.8, "avg": 2.1, "max": 2.5},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "688017")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := ths.NewClient(ths.WithBaseURL(srv.URL))
	fc, err := c.FetchEpsForecast(context.Background(), "688017")
	require.NoError(t, err)
	require.Len(t, fc.Years, 2)
	assert.Equal(t, "2026", fc.Years[0].Year)
	assert.Equal(t, 15, fc.Years[0].Count)
	assert.InDelta(t, 1.5, fc.Years[0].Mean, 0.001)
}

func TestFetchEpsForecast_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"status": -1})
	}))
	defer srv.Close()

	c := ths.NewClient(ths.WithBaseURL(srv.URL))
	_, err := c.FetchEpsForecast(context.Background(), "688017")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}
```

- [ ] **Step 3: 运行测试**

```bash
go test -race ./pkg/ths/...
# Expected: ok  stock/pkg/ths (2 tests pass)
```

- [ ] **Step 4: Commit**

```bash
git add pkg/ths/
git commit -m "feat: add ths client for institutional EPS forecasts"
```

---

## Task 2: Service — thsClient + 估值计算

**Files:**
- Modify: `pkg/stockd/services/service.go`
- Create: `pkg/stockd/services/ths.go`

- [ ] **Step 1: 在 `service.go` import 中加入 ths 包**

在 `"stock/pkg/eastmoney"` 后追加：
```go
"stock/pkg/ths"
```

- [ ] **Step 2: 在 `Service` struct 的 `emClient` 字段后追加**

```go
	thsClient   *ths.Client
```

- [ ] **Step 3: 在 `WithEastMoneyClient` 函数后追加**

```go
// WithThsClient injects a Tonghuashun client.
func WithThsClient(c *ths.Client) ServiceOption {
	return func(s *Service) { s.thsClient = c }
}
```

- [ ] **Step 4: 编译验证**

```bash
go build ./pkg/stockd/services/...
# Expected: no errors
```

- [ ] **Step 5: 创建 `pkg/stockd/services/ths.go`**

```go
package services

import (
	"context"
	"fmt"
	"math"

	"stock/pkg/ths"
)

// Valuation holds forward-looking valuation metrics.
type Valuation struct {
	Code       string         `json:"code"`
	Price      float64        `json:"price"`
	EpsForecast *ths.EpsForecast `json:"eps_forecast"`
	ForwardPE  *float64       `json:"forward_pe"`
	CAGR       *float64       `json:"cagr"`
	PEG        *float64       `json:"peg"`
	DigestYears *float64      `json:"digest_years"`
}

// GetEpsForecast returns institutional consensus EPS forecasts.
func (s *Service) GetEpsForecast(ctx context.Context, code string) (*ths.EpsForecast, error) {
	if s.thsClient == nil {
		return nil, fmt.Errorf("ths client not configured")
	}
	return s.thsClient.FetchEpsForecast(ctx, code)
}

// GetValuation returns forward valuation metrics using real-time price + EPS forecast.
func (s *Service) GetValuation(ctx context.Context, code string) (*Valuation, error) {
	if s.thsClient == nil {
		return nil, fmt.Errorf("ths client not configured")
	}

	// Fetch current price via existing tencent client.
	quotes, err := s.tc.GetStockQuotes(ctx, []string{code})
	if err != nil || len(quotes) == 0 {
		return nil, fmt.Errorf("valuation: fetch price: %w", err)
	}
	price := quotes[0].Price

	fc, err := s.thsClient.FetchEpsForecast(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("valuation: fetch eps: %w", err)
	}

	val := &Valuation{Code: code, Price: price, EpsForecast: fc}

	if len(fc.Years) >= 1 && fc.Years[0].Mean > 0 {
		fpe := price / fc.Years[0].Mean
		val.ForwardPE = &fpe

		if len(fc.Years) >= 2 && fc.Years[0].Mean > 0 {
			cagr := fc.Years[1].Mean/fc.Years[0].Mean - 1
			val.CAGR = &cagr
			if cagr > 0 {
				peg := fpe / (cagr * 100)
				val.PEG = &peg
				// PE digest years: how many years at this CAGR to bring PE down to 30x.
				dy := math.Log(fpe/30) / math.Log(1+cagr)
				val.DigestYears = &dy
			}
		}
	}
	return val, nil
}
```

- [ ] **Step 6: 编译验证**

```bash
go build ./pkg/stockd/services/...
# Expected: no errors
```

- [ ] **Step 7: Commit**

```bash
git add pkg/stockd/services/service.go pkg/stockd/services/ths.go
git commit -m "feat: add GetEpsForecast and GetValuation service methods"
```

---

## Task 3: HTTP Handlers + 路由 + main.go

**Files:**
- Modify: `pkg/stockd/http/external.go`
- Modify: `pkg/stockd/http/router.go`
- Modify: `cmd/stockd/main.go`

- [ ] **Step 1: 在 `pkg/stockd/http/external.go` 末尾追加两个 handler**

```go
// GetEpsForecast returns institutional consensus EPS forecasts (Tonghuashun).
func (h *handler) GetEpsForecast(c *gin.Context) {
	data, err := h.svc.GetEpsForecast(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}

// GetValuation returns forward PE, PEG, and PE digest years for a stock.
func (h *handler) GetValuation(c *gin.Context) {
	data, err := h.svc.GetValuation(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}
```

- [ ] **Step 2: 在 `router.go` 的 `stCode` 路由组末尾追加**

```go
	stCode.GET("/eps-forecast", h.GetEpsForecast)
	stCode.GET("/valuation", h.GetValuation)
```

- [ ] **Step 3: 在 `cmd/stockd/main.go` 中实例化 ths 客户端并注入**

在 import 块中加入：
```go
"stock/pkg/ths"
```

在 `emClient := eastmoney.NewClient()` 之后加：
```go
	thsClient := ths.NewClient()
```

将 `services.NewService(...)` 调用改为：
```go
	svc := services.NewService(gdb, tc, tencentClient, cfg, logger,
		services.WithBaiduClient(baiduClient),
		services.WithEastMoneyClient(emClient),
		services.WithThsClient(thsClient),
	)
```

- [ ] **Step 4: 完整编译**

```bash
go build ./cmd/stockd/
# Expected: no errors
```

- [ ] **Step 5: 运行全量测试**

```bash
make test
# Expected: all existing tests pass
```

- [ ] **Step 6: Final commit**

```bash
git add pkg/stockd/http/external.go pkg/stockd/http/router.go cmd/stockd/main.go
git commit -m "feat: register /eps-forecast and /valuation routes, wire ths client"
```
