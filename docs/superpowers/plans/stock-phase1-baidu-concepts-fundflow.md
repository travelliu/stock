# 个股 Phase 1: 百度概念板块 + 资金流向历史

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 接通两个个股端点：概念/行业/地域板块归属，以及最近 20 交易日资金流向（主力/散户/超大单/大单）。

**Architecture:** `pkg/baidu/` 客户端已完成。本阶段只需：Service 注入（ServiceOption 模式，不破坏现有 `NewService` 签名）、两个 Gin handler、路由注册、main.go 实例化。

**Prerequisite:** `pkg/baidu/client.go` 已存在，4 个测试全部通过。

**Tech Stack:** Go 1.25，标准库 `net/http`，Gin，现有 `pkg/stockd/utils` 响应辅助函数。

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `pkg/stockd/services/service.go` | 加 `baiduClient` 字段 + `ServiceOption` 模式 |
| Create | `pkg/stockd/services/baidu.go` | `GetConceptBlocks` / `GetFundFlowHistory` 服务方法 |
| Create | `pkg/stockd/http/external.go` | 两个 handler |
| Modify | `pkg/stockd/http/router.go` | 注册新路由 |
| Modify | `cmd/stockd/main.go` | 实例化 `baidu.Client` |

---

## Task 1: Service — ServiceOption + baidu 字段

**Files:**
- Modify: `pkg/stockd/services/service.go`

- [ ] **Step 1: 在 `service.go` 顶部 import 块中加入 baidu 包**

在现有 import 后追加：
```go
"stock/pkg/baidu"
```

- [ ] **Step 2: 在 `Service` struct 的 `realtimeMu` 字段后追加**

```go
	baiduClient *baidu.Client
```

- [ ] **Step 3: 在 `NewService` 函数定义之前（`func NewService` 上方）插入**

```go
// ServiceOption configures Service after construction.
type ServiceOption func(*Service)

// WithBaiduClient injects a Baidu PAE client.
func WithBaiduClient(c *baidu.Client) ServiceOption {
	return func(s *Service) { s.baiduClient = c }
}
```

- [ ] **Step 4: 将 `NewService` 签名改为接受可变参数 opts，并在返回前应用**

将现有：
```go
func NewService(db *gorm.DB, ts *tushare.Client, tc *tencent.Client, cfg *config.Config,
	logger *logrus.Logger) *Service {
	return &Service{
		db:            db,
		ts:            ts,
		tc:            tc,
		cfg:           cfg,
		cron:          cron.New(cron.WithLocation(time.Local)),
		jobs:          map[string]JobFunc{},
		logger:        logger,
		realtimeCache: make(map[string]*models.StockRealtimeAndAnalysis),
	}
}
```

替换为：
```go
func NewService(db *gorm.DB, ts *tushare.Client, tc *tencent.Client, cfg *config.Config,
	logger *logrus.Logger, opts ...ServiceOption) *Service {
	svc := &Service{
		db:            db,
		ts:            ts,
		tc:            tc,
		cfg:           cfg,
		cron:          cron.New(cron.WithLocation(time.Local)),
		jobs:          map[string]JobFunc{},
		logger:        logger,
		realtimeCache: make(map[string]*models.StockRealtimeAndAnalysis),
	}
	for _, o := range opts {
		o(svc)
	}
	return svc
}
```

- [ ] **Step 5: 验证编译**

```bash
go build ./...
# Expected: no errors（现有调用方传 0 个 opts，向后兼容）
```

- [ ] **Step 6: Commit**

```bash
git add pkg/stockd/services/service.go
git commit -m "refactor: add ServiceOption pattern to NewService for optional client injection"
```

---

## Task 2: Service — GetConceptBlocks / GetFundFlowHistory

**Files:**
- Create: `pkg/stockd/services/baidu.go`

- [ ] **Step 1: 创建 `pkg/stockd/services/baidu.go`**

```go
package services

import (
	"context"
	"fmt"

	"stock/pkg/baidu"
)

// GetConceptBlocks returns concept/industry/region block memberships for a stock.
// code is a 6-digit stock code.
func (s *Service) GetConceptBlocks(ctx context.Context, code string) (*baidu.ConceptBlocks, error) {
	if s.baiduClient == nil {
		return nil, fmt.Errorf("baidu client not configured")
	}
	return s.baiduClient.FetchConceptBlocks(ctx, code)
}

// GetFundFlowHistory returns daily fund flow for the last 20 trading days.
func (s *Service) GetFundFlowHistory(ctx context.Context, code string) ([]baidu.FundFlowDay, error) {
	if s.baiduClient == nil {
		return nil, fmt.Errorf("baidu client not configured")
	}
	return s.baiduClient.FetchFundFlowHistory(ctx, code, 20)
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./pkg/stockd/services/...
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add pkg/stockd/services/baidu.go
git commit -m "feat: add GetConceptBlocks and GetFundFlowHistory service methods"
```

---

## Task 3: HTTP Handlers

**Files:**
- Create: `pkg/stockd/http/external.go`

- [ ] **Step 1: 创建 `pkg/stockd/http/external.go`**

```go
package http

import (
	"github.com/gin-gonic/gin"
	"stock/pkg/stockd/utils"
)

// GetConceptBlocks returns concept/industry/region block memberships (Baidu PAE).
func (h *handler) GetConceptBlocks(c *gin.Context) {
	data, err := h.svc.GetConceptBlocks(c.Request.Context(), c.Param(codeValue))
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, data)
}

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

- [ ] **Step 2: 编译验证**

```bash
go build ./pkg/stockd/http/...
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add pkg/stockd/http/external.go
git commit -m "feat: add GetConceptBlocks and GetFundFlow HTTP handlers"
```

---

## Task 4: 路由 + main.go 接线

**Files:**
- Modify: `pkg/stockd/http/router.go`
- Modify: `cmd/stockd/main.go`

- [ ] **Step 1: 在 `router.go` 的 `stCode` 路由组中追加两条路由**

找到 `stCode` 路由组（`stCode.GET("/analysis", ...)`附近），在末尾追加：

```go
	stCode.GET("/concepts", h.GetConceptBlocks)
	stCode.GET("/fund-flow", h.GetFundFlow)
```

- [ ] **Step 2: 在 `cmd/stockd/main.go` 中实例化 baidu 客户端并注入**

在 import 块中加入：
```go
"stock/pkg/baidu"
```

在 `tencentClient := tencent.NewClient()` 之后加：
```go
	baiduClient := baidu.NewClient()
```

将 `services.NewService(...)` 调用改为：
```go
	svc := services.NewService(gdb, tc, tencentClient, cfg, logger,
		services.WithBaiduClient(baiduClient),
	)
```

- [ ] **Step 3: 完整编译**

```bash
go build ./cmd/stockd/
# Expected: no errors
```

- [ ] **Step 4: 运行全量测试**

```bash
make test
# Expected: all existing tests pass
```

- [ ] **Step 5: 手动冒烟测试**

```bash
go run ./cmd/stockd &
sleep 2

# 登录获取 session（替换实际用户名密码）
curl -s -c /tmp/c.jar -X POST http://localhost:8443/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq '.code'
# Expected: 200

# 概念板块
curl -s -b /tmp/c.jar http://localhost:8443/api/stocks/600519/concepts | jq '{industry: .data.industry[0].name, concept_count: (.data.concept_tags | length)}'
# Expected: industry name + concept count > 0

# 资金流向
curl -s -b /tmp/c.jar http://localhost:8443/api/stocks/600519/fund-flow | jq '.data | length'
# Expected: up to 20 rows
```

- [ ] **Step 6: Final commit**

```bash
git add pkg/stockd/http/router.go cmd/stockd/main.go
git commit -m "feat: register /concepts and /fund-flow routes, wire baidu client"
```
