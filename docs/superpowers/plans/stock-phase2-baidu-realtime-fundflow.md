# 个股 Phase 2: 百度实时资金流向

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `/api/stocks/:code/fund-flow` 增加 `?type=realtime` 支持，返回当日实时资金流向快照（主力/超大单/大单/中单/小单净流入额）。

**Architecture:** 在现有 `pkg/baidu/client.go` 新增 `FetchFundFlowRealtime` 方法（调用不同 Baidu PAE 端点），handler 通过 `type` query param 分发。不新建文件，只扩展已有文件。

**Prerequisite:** Phase 1 已完成（`pkg/baidu/` 客户端、service 方法、`/fund-flow` 历史端点全部上线）。

**Tech Stack:** Go 1.25，标准库 `net/http`，Gin，现有 `pkg/stockd/utils`。

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `pkg/baidu/client.go` | 新增 `FundFlowSnapshot` 结构体 + `FetchFundFlowRealtime` 方法 |
| Modify | `pkg/baidu/client_test.go` | 新增 `TestFetchFundFlowRealtime` + error case |
| Modify | `pkg/stockd/services/baidu.go` | 新增 `GetFundFlowRealtime` service 方法 |
| Modify | `pkg/stockd/http/external.go` | `GetFundFlow` 根据 `?type=` 分发 |

---

## Task 1: baidu 客户端 — FundFlowSnapshot + FetchFundFlowRealtime

**Files:**
- Modify: `pkg/baidu/client.go`
- Modify: `pkg/baidu/client_test.go`

> **注意：** Baidu PAE 实时资金流向端点在实现前需用浏览器 DevTools 核实 URL 和响应结构。
> 本计划使用推测路径 `/vapi/v1/fundflow`，若实际不符，请在 Step 1 前先核实并调整。

- [ ] **Step 1: 在 `pkg/baidu/client.go` 末尾追加 `FundFlowSnapshot` 结构体**

在 `FundFlowDay` 定义之后，`get` 函数之前，追加：

```go
// FundFlowSnapshot is the current-day real-time fund flow for a stock.
type FundFlowSnapshot struct {
	Date        string `json:"date"`
	Close       string `json:"close"`
	ChangePct   string `json:"change_pct"`
	SuperNetIn  string `json:"super_net_in"`
	LargeNetIn  string `json:"large_net_in"`
	MediumNetIn string `json:"medium_net_in"`
	LittleNetIn string `json:"little_net_in"`
	MainIn      string `json:"main_in"`
}
```

- [ ] **Step 2: 在 `pkg/baidu/client.go` 末尾追加 `FetchFundFlowRealtime` 方法**

在 `str` 函数之前追加：

```go
// FetchFundFlowRealtime returns today's real-time fund flow snapshot for a stock.
// Endpoint: /vapi/v1/fundflow (verify against live API before shipping).
func (c *Client) FetchFundFlowRealtime(ctx context.Context, code string) (*FundFlowSnapshot, error) {
	path := fmt.Sprintf("/vapi/v1/fundflow?code=%s&market=ab&finClientType=pc", code)
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		ResultCode any `json:"ResultCode"`
		Result     struct {
			ShowTime    string `json:"showtime"`
			ClosePx     string `json:"closepx"`
			Ratio       string `json:"ratio"`
			SuperNetIn  string `json:"superNetIn"`
			LargeNetIn  string `json:"largeNetIn"`
			MediumNetIn string `json:"mediumNetIn"`
			LittleNetIn string `json:"littleNetIn"`
			ExtMainIn   string `json:"extMainIn"`
		} `json:"Result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("baidu: parse fund flow realtime: %w", err)
	}
	if fmt.Sprintf("%v", raw.ResultCode) != "0" {
		return nil, fmt.Errorf("baidu: API error ResultCode=%v", raw.ResultCode)
	}

	return &FundFlowSnapshot{
		Date:        raw.Result.ShowTime,
		Close:       raw.Result.ClosePx,
		ChangePct:   raw.Result.Ratio,
		SuperNetIn:  raw.Result.SuperNetIn,
		LargeNetIn:  raw.Result.LargeNetIn,
		MediumNetIn: raw.Result.MediumNetIn,
		LittleNetIn: raw.Result.LittleNetIn,
		MainIn:      raw.Result.ExtMainIn,
	}, nil
}
```

- [ ] **Step 3: 在 `pkg/baidu/client_test.go` 追加两个测试**

在文件末尾追加：

```go
func TestFetchFundFlowRealtime(t *testing.T) {
	payload := map[string]any{
		"ResultCode": "0",
		"Result": map[string]any{
			"showtime":    "2026-05-16",
			"closepx":     "224.12",
			"ratio":       "4.24",
			"superNetIn":  "5000",
			"largeNetIn":  "3000",
			"mediumNetIn": "-1000",
			"littleNetIn": "-7000",
			"extMainIn":   "8000",
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "688017")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := baidu.NewClient(baidu.WithBaseURL(srv.URL))
	snap, err := c.FetchFundFlowRealtime(context.Background(), "688017")
	require.NoError(t, err)
	assert.Equal(t, "2026-05-16", snap.Date)
	assert.Equal(t, "8000", snap.MainIn)
}

func TestFetchFundFlowRealtime_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ResultCode": -1})
	}))
	defer srv.Close()

	c := baidu.NewClient(baidu.WithBaseURL(srv.URL))
	_, err := c.FetchFundFlowRealtime(context.Background(), "688017")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}
```

- [ ] **Step 4: 运行测试确认全绿**

```bash
go test -race ./pkg/baidu/...
# Expected: ok  stock/pkg/baidu  (6 tests pass)
```

- [ ] **Step 5: Commit**

```bash
git add pkg/baidu/client.go pkg/baidu/client_test.go
git commit -m "feat: add FetchFundFlowRealtime to baidu client"
```

---

## Task 2: Service + Handler — type 分发

**Files:**
- Modify: `pkg/stockd/services/baidu.go`
- Modify: `pkg/stockd/http/external.go`

- [ ] **Step 1: 在 `pkg/stockd/services/baidu.go` 末尾追加 `GetFundFlowRealtime`**

```go
// GetFundFlowRealtime returns today's real-time fund flow snapshot.
func (s *Service) GetFundFlowRealtime(ctx context.Context, code string) (*baidu.FundFlowSnapshot, error) {
	if s.baiduClient == nil {
		return nil, fmt.Errorf("baidu client not configured")
	}
	return s.baiduClient.FetchFundFlowRealtime(ctx, code)
}
```

- [ ] **Step 2: 修改 `pkg/stockd/http/external.go` 中的 `GetFundFlow` 以支持 `?type=` 分发**

将现有：

```go
// GetFundFlow returns per-stock fund flow history for the last 20 trading days (Baidu PAE).
func (h *handler) GetFundFlow(c *gin.Context) {
	data, err := h.svc.GetFundFlowHistory(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}
```

替换为：

```go
// GetFundFlow returns per-stock fund flow. ?type=realtime returns today's snapshot;
// any other value (default) returns the last 20 trading days history.
func (h *handler) GetFundFlow(c *gin.Context) {
	code := c.Param(codeValue)
	if c.Query("type") == "realtime" {
		data, err := h.svc.GetFundFlowRealtime(c.Request.Context(), code)
		if err != nil {
			utils.HTTPRequestFailedV5(c, err)
			return
		}
		utils.HTTPRequestSuccess(c, 200, data)
		return
	}
	data, err := h.svc.GetFundFlowHistory(c.Request.Context(), code)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}
```

- [ ] **Step 3: 完整编译**

```bash
go build ./...
# Expected: no errors
```

- [ ] **Step 4: 运行全量测试**

```bash
make test
# Expected: all existing tests pass
```

- [ ] **Step 5: Commit**

```bash
git add pkg/stockd/services/baidu.go pkg/stockd/http/external.go
git commit -m "feat: add ?type=realtime support to /fund-flow endpoint"
```
