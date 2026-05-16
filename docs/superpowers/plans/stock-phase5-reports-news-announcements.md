# 个股 Phase 5: 研究报告 + 个股新闻 + 公告

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 接通三个个股端点：研究报告列表（东方财富 reportapi）、个股新闻、交易所公告（cninfo）。另加研究报告 PDF 代理端点。

**Architecture:** 扩展 `pkg/eastmoney/` 客户端（报告 + 新闻），新建 `pkg/cninfo/` 客户端（公告），报告 PDF 代理在 handler 层实现（不通过 service，直接 http.Get + io.Copy）。

**Prerequisite:** Phase 3 已完成（`pkg/eastmoney/` 已存在，`emClient` 已注入）。

**Tech Stack:** Go 1.25，标准库 `net/http`，Gin，现有 `pkg/stockd/utils`。

**API 说明:**
- 研究报告：`https://reportapi.eastmoney.com/report/list?cb=datatable&industryCode=*&pageSize=20&industry=*&rating=&ratingChange=&beginTime={date}&endTime=&pageNo=1&fields=&qType=0&orgCode=&code={code}&rtype=99&latestReport=0`
- 个股新闻：东方财富新闻 API（需 DevTools 核实 URL）
- 公告：`https://www.cninfo.com.cn/new/disclosure/stock/announcement/...`（需 DevTools 核实）

> **以上 URL 实现前均需核实，以实际 Network 请求为准。**

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `pkg/eastmoney/client.go` | 加 `ReportRecord` + `NewsRecord` 结构体 + 两个 fetch 方法 |
| Modify | `pkg/eastmoney/client_test.go` | 4 个新测试 |
| Create | `pkg/cninfo/client.go` | Client + `AnnouncementRecord` + `FetchAnnouncements` |
| Create | `pkg/cninfo/client_test.go` | 2 个单元测试 |
| Modify | `pkg/stockd/services/service.go` | 加 `cninfoClient *cninfo.Client` + `WithCninfoClient` option |
| Create | `pkg/stockd/services/reports.go` | `GetReports` / `GetStockNews` / `GetAnnouncements` 服务方法 |
| Modify | `pkg/stockd/http/external.go` | 追加 4 个 handler（含 PDF 代理） |
| Modify | `pkg/stockd/http/router.go` | 注册新路由 |
| Modify | `cmd/stockd/main.go` | 实例化 `cninfo.Client` |

---

## Task 1: pkg/eastmoney — 研究报告 + 新闻

**Files:**
- Modify: `pkg/eastmoney/client.go`
- Modify: `pkg/eastmoney/client_test.go`

- [ ] **Step 1: 在 `pkg/eastmoney/client.go` 末尾（`str` 函数之前）追加结构体 + 方法**

```go
// ReportRecord is one research report entry.
type ReportRecord struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	OrgName   string `json:"org_name"`
	Date      string `json:"date"`
	Rating    string `json:"rating"`
	PDFURL    string `json:"pdf_url"`
}

// NewsRecord is one news article for a stock.
type NewsRecord struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Source    string `json:"source"`
	Published string `json:"published"`
}

// FetchReports returns research reports for a stock from East Money reportapi.
func (c *Client) FetchReports(ctx context.Context, code string) ([]ReportRecord, error) {
	path := fmt.Sprintf(
		"/report/list?cb=datatable&industryCode=*&pageSize=20&industry=*&rating=&ratingChange=&beginTime=&endTime=&pageNo=1&fields=&qType=0&orgCode=&code=%s&rtype=99&latestReport=0",
		code,
	)
	// reportapi uses a different base URL; the Client base URL for this call is overridden.
	// During implementation, set reportBaseURL = "https://reportapi.eastmoney.com" separately.
	// TODO: verify full URL + response shape in DevTools before implementing.
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	// Response is JSONP: datatable({...}); strip callback wrapper.
	s := string(body)
	if len(s) > 10 && s[:9] == "datatable" {
		start := len("datatable(")
		end := len(s) - 1
		if end > start {
			s = s[start:end]
		}
	}

	var raw struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, fmt.Errorf("eastmoney: parse reports: %w", err)
	}

	out := make([]ReportRecord, 0, len(raw.Data))
	for _, item := range raw.Data {
		out = append(out, ReportRecord{
			ID:      str(item["reportId"]),
			Title:   str(item["title"]),
			OrgName: str(item["orgSName"]),
			Date:    str(item["publishDate"]),
			Rating:  str(item["starRating"]),
			PDFURL:  str(item["encodeUrl"]),
		})
	}
	return out, nil
}

// FetchStockNews returns recent news articles for a stock from East Money.
// NOTE: verify API URL in DevTools before implementing.
func (c *Client) FetchStockNews(ctx context.Context, code string) ([]NewsRecord, error) {
	path := fmt.Sprintf("/api/news/stock?code=%s&pageSize=20&pageNumber=1", code)
	body, err := c.get(ctx, path)
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
		return nil, fmt.Errorf("eastmoney: parse news: %w", err)
	}
	if !raw.Success {
		return nil, fmt.Errorf("eastmoney: news API returned success=false")
	}

	out := make([]NewsRecord, 0, len(raw.Result.Data))
	for _, item := range raw.Result.Data {
		out = append(out, NewsRecord{
			Title:     str(item["title"]),
			URL:       str(item["url"]),
			Source:    str(item["mediaName"]),
			Published: str(item["publishTime"]),
		})
	}
	return out, nil
}
```

- [ ] **Step 2: 在 `pkg/eastmoney/client_test.go` 末尾追加 4 个测试**

```go
func TestFetchReports(t *testing.T) {
	// reportapi returns JSONP: datatable({...})
	payload := `datatable({"data":[{"reportId":"123","title":"深度报告","orgSName":"中金公司","publishDate":"2026-05-15","starRating":"买入","encodeUrl":"https://pdf.example.com/123.pdf"}]})`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "688017")
		w.Write([]byte(payload))
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	rows, err := c.FetchReports(context.Background(), "688017")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "深度报告", rows[0].Title)
	assert.Equal(t, "中金公司", rows[0].OrgName)
}

func TestFetchReports_ParseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	_, err := c.FetchReports(context.Background(), "688017")
	require.Error(t, err)
}

func TestFetchStockNews(t *testing.T) {
	payload := map[string]any{
		"success": true,
		"result": map[string]any{
			"data": []map[string]any{
				{"title": "某股创新高", "url": "https://news.example.com/1", "mediaName": "东方财富", "publishTime": "2026-05-16 10:00"},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	rows, err := c.FetchStockNews(context.Background(), "688017")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "某股创新高", rows[0].Title)
}

func TestFetchStockNews_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"success": false})
	}))
	defer srv.Close()

	c := eastmoney.NewClient(eastmoney.WithBaseURL(srv.URL))
	_, err := c.FetchStockNews(context.Background(), "688017")
	require.Error(t, err)
}
```

- [ ] **Step 3: 运行测试**

```bash
go test -race ./pkg/eastmoney/...
# Expected: ok  stock/pkg/eastmoney (8 tests pass)
```

- [ ] **Step 4: Commit**

```bash
git add pkg/eastmoney/
git commit -m "feat: add FetchReports and FetchStockNews to eastmoney client"
```

---

## Task 2: pkg/cninfo — 公告客户端

**Files:**
- Create: `pkg/cninfo/client.go`
- Create: `pkg/cninfo/client_test.go`

- [ ] **Step 1: 创建 `pkg/cninfo/client.go`**

> **注意：** cninfo API URL 和请求方式（GET/POST）需在实现前用 DevTools 核实，本计划使用推测结构。

```go
// Package cninfo fetches exchange announcements from cninfo.com.cn.
package cninfo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://www.cninfo.com.cn"

var defaultHeaders = map[string]string{
	"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/117.0.0.0 Safari/537.36",
	"Referer":      "https://www.cninfo.com.cn/",
	"Content-Type": "application/x-www-form-urlencoded",
}

// Client calls cninfo announcement APIs.
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

// AnnouncementRecord is one exchange announcement entry.
type AnnouncementRecord struct {
	Title     string `json:"title"`
	Date      string `json:"date"`
	URL       string `json:"url"`
	Category  string `json:"category"`
}

// FetchAnnouncements returns recent exchange announcements for a stock.
// NOTE: cninfo uses POST with form-encoded body; verify before implementing.
func (c *Client) FetchAnnouncements(ctx context.Context, code string) ([]AnnouncementRecord, error) {
	form := url.Values{
		"stock":      {code + ","},
		"tabName":    {"fulltext"},
		"pageSize":   {"20"},
		"pageNum":    {"1"},
		"column":     {"szse"},
		"category":   {""},
		"plate":      {""},
		"seDate":     {""},
		"searchkey":  {""},
		"secid":      {""},
		"sortName":   {""},
		"sortType":   {""},
		"isHLtitle":  {"true"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/new/disclosure/hisAnnouncement/query",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("cninfo: build request: %w", err)
	}
	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cninfo: request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cninfo: read body: %w", err)
	}

	var raw struct {
		Announcements []map[string]any `json:"announcements"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("cninfo: parse announcements: %w", err)
	}

	out := make([]AnnouncementRecord, 0, len(raw.Announcements))
	for _, item := range raw.Announcements {
		annoURL := ""
		if adjunctURL, ok := item["adjunctUrl"].(string); ok {
			annoURL = "https://static.cninfo.com.cn/" + adjunctURL
		}
		out = append(out, AnnouncementRecord{
			Title:    str(item["announcementTitle"]),
			Date:     str(item["announcementTime"]),
			URL:      annoURL,
			Category: str(item["announcementType"]),
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

- [ ] **Step 2: 创建 `pkg/cninfo/client_test.go`**

```go
package cninfo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"stock/pkg/cninfo"
)

func TestFetchAnnouncements(t *testing.T) {
	payload := map[string]any{
		"announcements": []map[string]any{
			{
				"announcementTitle": "2026年一季报",
				"announcementTime":  "1747324800000",
				"adjunctUrl":        "finalpage/2026-05-15/123.pdf",
				"announcementType":  "定期报告",
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := cninfo.NewClient(cninfo.WithBaseURL(srv.URL))
	rows, err := c.FetchAnnouncements(context.Background(), "688017")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "2026年一季报", rows[0].Title)
	assert.Contains(t, rows[0].URL, "123.pdf")
}

func TestFetchAnnouncements_ParseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := cninfo.NewClient(cninfo.WithBaseURL(srv.URL))
	_, err := c.FetchAnnouncements(context.Background(), "688017")
	require.Error(t, err)
}
```

- [ ] **Step 3: 运行测试**

```bash
go test -race ./pkg/cninfo/...
# Expected: ok  stock/pkg/cninfo (2 tests pass)
```

- [ ] **Step 4: Commit**

```bash
git add pkg/cninfo/
git commit -m "feat: add cninfo client for exchange announcements"
```

---

## Task 3: Service — cninfoClient + 三个服务方法

**Files:**
- Modify: `pkg/stockd/services/service.go`
- Create: `pkg/stockd/services/reports.go`

- [ ] **Step 1: 在 `service.go` import 中加入 cninfo 包**

在 `"stock/pkg/ths"` 后追加：
```go
"stock/pkg/cninfo"
```

- [ ] **Step 2: 在 `Service` struct 的 `thsClient` 字段后追加**

```go
	cninfoClient *cninfo.Client
```

- [ ] **Step 3: 在 `WithThsClient` 函数后追加**

```go
// WithCninfoClient injects a cninfo announcements client.
func WithCninfoClient(c *cninfo.Client) ServiceOption {
	return func(s *Service) { s.cninfoClient = c }
}
```

- [ ] **Step 4: 创建 `pkg/stockd/services/reports.go`**

```go
package services

import (
	"context"
	"fmt"

	"stock/pkg/cninfo"
	"stock/pkg/eastmoney"
)

// GetReports returns research report list for a stock.
func (s *Service) GetReports(ctx context.Context, code string) ([]eastmoney.ReportRecord, error) {
	if s.emClient == nil {
		return nil, fmt.Errorf("eastmoney client not configured")
	}
	return s.emClient.FetchReports(ctx, code)
}

// GetStockNews returns recent news for a stock.
func (s *Service) GetStockNews(ctx context.Context, code string) ([]eastmoney.NewsRecord, error) {
	if s.emClient == nil {
		return nil, fmt.Errorf("eastmoney client not configured")
	}
	return s.emClient.FetchStockNews(ctx, code)
}

// GetAnnouncements returns exchange announcements for a stock.
func (s *Service) GetAnnouncements(ctx context.Context, code string) ([]cninfo.AnnouncementRecord, error) {
	if s.cninfoClient == nil {
		return nil, fmt.Errorf("cninfo client not configured")
	}
	return s.cninfoClient.FetchAnnouncements(ctx, code)
}
```

- [ ] **Step 5: 编译验证**

```bash
go build ./pkg/stockd/services/...
# Expected: no errors
```

- [ ] **Step 6: Commit**

```bash
git add pkg/stockd/services/service.go pkg/stockd/services/reports.go
git commit -m "feat: add GetReports, GetStockNews, GetAnnouncements service methods"
```

---

## Task 4: HTTP Handlers（含 PDF 代理）+ 路由 + main.go

**Files:**
- Modify: `pkg/stockd/http/external.go`
- Modify: `pkg/stockd/http/router.go`
- Modify: `cmd/stockd/main.go`

- [ ] **Step 1: 在 `pkg/stockd/http/external.go` 追加 import `"io"` 和 `"net/http"`（若未存在）**

确认 import 块包含：
```go
import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"stock/pkg/stockd/utils"
)
```

- [ ] **Step 2: 在 `pkg/stockd/http/external.go` 末尾追加 4 个 handler**

```go
// GetReports returns research report list for a stock (East Money reportapi).
func (h *handler) GetReports(c *gin.Context) {
	data, err := h.svc.GetReports(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}

// GetReportPDF proxies the PDF download from East Money CDN to avoid browser CORS.
func (h *handler) GetReportPDF(c *gin.Context) {
	pdfURL := c.Query("url")
	if pdfURL == "" {
		c.JSON(400, gin.H{"message": "missing url param"})
		return
	}
	resp, err := http.Get(pdfURL) //nolint:noctx
	if err != nil {
		c.JSON(502, gin.H{"message": "failed to fetch PDF"})
		return
	}
	defer resp.Body.Close()
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline")
	c.Status(200)
	io.Copy(c.Writer, resp.Body) //nolint:errcheck
}

// GetStockNews returns recent news for a stock (East Money).
func (h *handler) GetStockNews(c *gin.Context) {
	data, err := h.svc.GetStockNews(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}

// GetAnnouncements returns exchange announcements for a stock (cninfo).
func (h *handler) GetAnnouncements(c *gin.Context) {
	data, err := h.svc.GetAnnouncements(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}
```

- [ ] **Step 3: 在 `router.go` 的 `stCode` 路由组末尾追加**

```go
	stCode.GET("/reports", h.GetReports)
	stCode.GET("/reports/pdf", h.GetReportPDF)
	stCode.GET("/news", h.GetStockNews)
	stCode.GET("/announcements", h.GetAnnouncements)
```

- [ ] **Step 4: 在 `cmd/stockd/main.go` 中实例化 cninfo 客户端并注入**

在 import 块中加入：
```go
"stock/pkg/cninfo"
```

在 `thsClient := ths.NewClient()` 之后加：
```go
	cninfoClient := cninfo.NewClient()
```

将 `services.NewService(...)` 调用改为：
```go
	svc := services.NewService(gdb, tc, tencentClient, cfg, logger,
		services.WithBaiduClient(baiduClient),
		services.WithEastMoneyClient(emClient),
		services.WithThsClient(thsClient),
		services.WithCninfoClient(cninfoClient),
	)
```

- [ ] **Step 5: 完整编译**

```bash
go build ./cmd/stockd/
# Expected: no errors
```

- [ ] **Step 6: 运行全量测试**

```bash
make test
# Expected: all existing tests pass
```

- [ ] **Step 7: Final commit**

```bash
git add pkg/stockd/http/external.go pkg/stockd/http/router.go cmd/stockd/main.go
git commit -m "feat: register /reports, /reports/pdf, /news, /announcements routes, wire cninfo client"
```
