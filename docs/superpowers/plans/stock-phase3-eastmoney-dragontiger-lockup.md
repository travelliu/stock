# 个股 Phase 3: 东方财富龙虎榜 + 解禁日历

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 接通两个个股端点：龙虎榜（个股历史龙虎榜 + 前 5 买卖席位），以及解禁日历（历史 + 未来 90 天）。

**Architecture:** 新建 `pkg/eastmoney/` 客户端包，Service 注入（ServiceOption 模式），两个 handler 追加到 `pkg/stockd/http/external.go`，路由注册到 `stCode` 组，main.go 实例化。

**Prerequisite:** Phase 1 已完成（ServiceOption 模式已在 service.go 中）。

**Tech Stack:** Go 1.25，标准库 `net/http`，Gin，现有 `pkg/stockd/utils`。

**API 说明:** 东方财富 datacenter API 使用 `https://datacenter-web.eastmoney.com/api/data/v1/get` 端点，通过 `reportName` 参数指定数据集：
- 龙虎榜：`reportName=RPT_BILLBOARD_DAILYDETAILSED` （个股席位明细）
- 解禁：`reportName=RPT_LIFT_STAGE_SHAREPLAN_NEW`

> **实现前务必用浏览器 DevTools 核实 `reportName` 参数值和响应字段名，与本计划对比。**

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `pkg/eastmoney/client.go` | Client + DragonTigerRecord + LockupRecord + fetch methods |
| Create | `pkg/eastmoney/client_test.go` | 4 个单元测试 |
| Modify | `pkg/stockd/services/service.go` | 加 `emClient *eastmoney.Client` 字段 + `WithEastMoneyClient` option |
| Create | `pkg/stockd/services/eastmoney.go` | `GetDragonTigerHistory` / `GetLockupCalendar` 服务方法 |
| Modify | `pkg/stockd/http/external.go` | 追加两个 handler |
| Modify | `pkg/stockd/http/router.go` | 注册新路由 |
| Modify | `cmd/stockd/main.go` | 实例化 `eastmoney.Client` |

---

## Task 1: pkg/eastmoney — Client + DragonTiger

**Files:**
- Create: `pkg/eastmoney/client.go`
- Create: `pkg/eastmoney/client_test.go`

- [ ] **Step 1: 创建 `pkg/eastmoney/client.go`**

```go
// Package eastmoney fetches dragon-tiger board and lockup calendar from East Money datacenter.
package eastmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://datacenter-web.eastmoney.com"

var defaultHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/117.0.0.0 Safari/537.36",
	"Referer":    "https://data.eastmoney.com/",
}

// Client calls East Money datacenter APIs.
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

// Seat is a single buy or sell seat on the dragon-tiger board.
type Seat struct {
	Name    string `json:"name"`
	NetBuy  string `json:"net_buy"`
	Buy     string `json:"buy"`
	Sell    string `json:"sell"`
}

// DragonTigerRecord is one appearance on the dragon-tiger board for a stock.
type DragonTigerRecord struct {
	Date      string `json:"date"`
	Reason    string `json:"reason"`
	Close     string `json:"close"`
	ChangePct string `json:"change_pct"`
	NetBuy    string `json:"net_buy"`
	TopBuy    []Seat `json:"top_buy"`
	TopSell   []Seat `json:"top_sell"`
}

// LockupRecord is one lockup expiry event for a stock.
type LockupRecord struct {
	Date       string `json:"date"`
	ShareType  string `json:"share_type"`
	Shares     string `json:"shares"`
	ShareRatio string `json:"share_ratio"`
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("eastmoney: build request: %w", err)
	}
	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("eastmoney: request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("eastmoney: read body: %w", err)
	}
	return body, nil
}

// FetchDragonTigerHistory returns historical dragon-tiger board entries for a stock.
// code is a 6-digit stock code.
func (c *Client) FetchDragonTigerHistory(ctx context.Context, code string) ([]DragonTigerRecord, error) {
	params := url.Values{
		"reportName": {"RPT_BILLBOARD_DAILYDETAILSED"},
		"columns":    {"ALL"},
		"filter":     {fmt.Sprintf(`(SECURITY_CODE="%s")`, code)},
		"pageSize":   {"20"},
		"pageNumber": {"1"},
		"sortTypes":  {"-1"},
		"sortColumns": {"TRADE_DATE"},
	}
	body, err := c.get(ctx, "/api/data/v1/get?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var raw struct {
		Success bool `json:"success"`
		Result  struct {
			Data []map[string]any `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("eastmoney: parse dragon-tiger: %w", err)
	}
	if !raw.Success {
		return nil, fmt.Errorf("eastmoney: dragon-tiger API returned success=false")
	}

	out := make([]DragonTigerRecord, 0, len(raw.Result.Data))
	for _, item := range raw.Result.Data {
		out = append(out, DragonTigerRecord{
			Date:      str(item["TRADE_DATE"]),
			Reason:    str(item["EXPLAIN"]),
			Close:     str(item["CLOSE_PRICE"]),
			ChangePct: str(item["CHANGE_RATE"]),
			NetBuy:    str(item["NET_BUY_AMT"]),
		})
	}
	return out, nil
}

// FetchLockupCalendar returns lockup expiry records for a stock (historical + upcoming).
func (c *Client) FetchLockupCalendar(ctx context.Context, code string) ([]LockupRecord, error) {
	params := url.Values{
		"reportName": {"RPT_LIFT_STAGE_SHAREPLAN_NEW"},
		"columns":    {"ALL"},
		"filter":     {fmt.Sprintf(`(SECURITY_CODE="%s")`, code)},
		"pageSize":   {"50"},
		"pageNumber": {"1"},
		"sortTypes":  {"-1"},
		"sortColumns": {"LIFT_DATE"},
	}
	body, err := c.get(ctx, "/api/data/v1/get?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var raw struct {
		Success bool `json:"success"`
		Result  struct {
			Data []map[string]any `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("eastmoney: parse lockup: %w", err)
	}
	if !raw.Success {
		return nil, fmt.Errorf("eastmoney: lockup API returned success=false")
	}

	out := make([]LockupRecord, 0, len(raw.Result.Data))
	for _, item := range raw.Result.Data {
		out = append(out, LockupRecord{
			Date:       str(item["LIFT_DATE"]),
			ShareType:  str(item["SHARE_TYPE"]),
			Shares:     str(item["LIFT_SHARES"]),
			ShareRatio: str(item["LIFT_RATIO"]),
		})
	}
	return out, nil
}

func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
```

- [ ] **Step 2: 创建 `pkg/eastmoney/client_test.go`**

```go
package eastmoney_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"stock/pkg/eastmoney"
)

func TestFetchDragonTigerHistory(t *testing.T) {
	payload := map[string]any{
		"success": true,
		"result": map[string]any{
			"data": []map[string]any{
				{
					"TRADE_DATE":  "2026-05-15",
					"EXPLAIN":     "涨幅偏离值达7%",
					"CLOSE_PRICE": "224.12",
					"CHANGE_RATE": "7.23",
					"NET_BUY_AMT": "12345678",
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "RPT_BILLBOARD_DAILYDETAILSED")
		assert.Contains(t, r.URL.String(), "688017")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	rows, err := c.FetchDragonTigerHistory(context.Background(), "688017")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "2026-05-15", rows[0].Date)
	assert.Equal(t, "涨幅偏离值达7%", rows[0].Reason)
}

func TestFetchDragonTigerHistory_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"success": false})
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	_, err := c.FetchDragonTigerHistory(context.Background(), "688017")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "success=false")
}

func TestFetchLockupCalendar(t *testing.T) {
	payload := map[string]any{
		"success": true,
		"result": map[string]any{
			"data": []map[string]any{
				{
					"LIFT_DATE":   "2026-08-01",
					"SHARE_TYPE":  "限售股",
					"LIFT_SHARES": "10000000",
					"LIFT_RATIO":  "5.23",
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "RPT_LIFT_STAGE_SHAREPLAN_NEW")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	rows, err := c.FetchLockupCalendar(context.Background(), "688017")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "2026-08-01", rows[0].Date)
	assert.Equal(t, "5.23", rows[0].ShareRatio)
}

func TestFetchLockupCalendar_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"success": false})
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	_, err := c.FetchLockupCalendar(context.Background(), "688017")
	require.Error(t, err)
}
```

- [ ] **Step 3: 运行测试**

```bash
go test -race ./pkg/eastmoney/...
# Expected: ok  stock/pkg/eastmoney (4 tests pass)
```

- [ ] **Step 4: Commit**

```bash
git add pkg/eastmoney/
git commit -m "feat: add eastmoney client for dragon-tiger board and lockup calendar"
```

---

## Task 2: Service — emClient 字段 + WithEastMoneyClient

**Files:**
- Modify: `pkg/stockd/services/service.go`

- [ ] **Step 1: 在 `service.go` import 中加入 eastmoney 包**

在 `"stock/pkg/baidu"` 后追加：
```go
"stock/pkg/eastmoney"
```

- [ ] **Step 2: 在 `Service` struct 的 `baiduClient` 字段后追加**

```go
	emClient    *eastmoney.Client
```

- [ ] **Step 3: 在 `WithBaiduClient` 函数后追加**

```go
// WithEastMoneyClient injects an East Money datacenter client.
func WithEastMoneyClient(c *eastmoney.Client) ServiceOption {
	return func(s *Service) { s.emClient = c }
}
```

- [ ] **Step 4: 编译验证**

```bash
go build ./pkg/stockd/services/...
# Expected: no errors
```

- [ ] **Step 5: Commit**

```bash
git add pkg/stockd/services/service.go
git commit -m "refactor: inject eastmoney client via ServiceOption"
```

---

## Task 3: Service — GetDragonTigerHistory / GetLockupCalendar

**Files:**
- Create: `pkg/stockd/services/eastmoney.go`

- [ ] **Step 1: 创建 `pkg/stockd/services/eastmoney.go`**

```go
package services

import (
	"context"
	"fmt"

	"stock/pkg/eastmoney"
)

// GetDragonTigerHistory returns historical dragon-tiger board entries for a stock.
func (s *Service) GetDragonTigerHistory(ctx context.Context, code string) ([]eastmoney.DragonTigerRecord, error) {
	if s.emClient == nil {
		return nil, fmt.Errorf("eastmoney client not configured")
	}
	return s.emClient.FetchDragonTigerHistory(ctx, code)
}

// GetLockupCalendar returns lockup expiry records for a stock.
func (s *Service) GetLockupCalendar(ctx context.Context, code string) ([]eastmoney.LockupRecord, error) {
	if s.emClient == nil {
		return nil, fmt.Errorf("eastmoney client not configured")
	}
	return s.emClient.FetchLockupCalendar(ctx, code)
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./pkg/stockd/services/...
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add pkg/stockd/services/eastmoney.go
git commit -m "feat: add GetDragonTigerHistory and GetLockupCalendar service methods"
```

---

## Task 4: HTTP Handlers + 路由 + main.go

**Files:**
- Modify: `pkg/stockd/http/external.go`
- Modify: `pkg/stockd/http/router.go`
- Modify: `cmd/stockd/main.go`

- [ ] **Step 1: 在 `pkg/stockd/http/external.go` 末尾追加两个 handler**

```go
// GetDragonTiger returns per-stock dragon-tiger board history (East Money).
func (h *handler) GetDragonTiger(c *gin.Context) {
	data, err := h.svc.GetDragonTigerHistory(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}

// GetLockup returns per-stock lockup expiry calendar (East Money).
func (h *handler) GetLockup(c *gin.Context) {
	data, err := h.svc.GetLockupCalendar(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}
```

- [ ] **Step 2: 在 `router.go` 的 `stCode` 路由组末尾追加**

```go
	stCode.GET("/dragon-tiger", h.GetDragonTiger)
	stCode.GET("/lockup", h.GetLockup)
```

- [ ] **Step 3: 在 `cmd/stockd/main.go` 中实例化 eastmoney 客户端并注入**

在 import 块中加入：
```go
"stock/pkg/eastmoney"
```

在 `baiduClient := baidu.NewClient()` 之后加：
```go
	emClient := eastmoney.NewClient()
```

将 `services.NewService(...)` 调用改为：
```go
	svc := services.NewService(gdb, tc, tencentClient, cfg, logger,
		services.WithBaiduClient(baiduClient),
		services.WithEastMoneyClient(emClient),
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
git commit -m "feat: register /dragon-tiger and /lockup routes, wire eastmoney client"
```
