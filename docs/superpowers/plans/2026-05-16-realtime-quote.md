# Realtime Quote Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add intraday realtime price data by polling Tencent Finance (`qt.gtimg.cn`) every 30 seconds during trading hours, caching results in memory, and exposing them via a new REST endpoint and updated frontend views.

**Architecture:** Backend polls `qt.gtimg.cn` in a cron job (every 30s, 09:15–15:00 Asia/Shanghai, weekdays only). Quotes are cached in `Service.realtimeCache` (RWMutex-guarded map). On cache miss, a single on-demand fetch is performed. Frontend calls `GET /api/quotes/:code` and displays realtime price in `StockBasicCard` (detail page) and `StockListView` (portfolio page).

**Tech Stack:** Go 1.25, Gin, robfig/cron v3, net/http (standard), Vue 3, TypeScript, Element Plus.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `pkg/models/realtime_quote.go` | Create | `RealtimeQuote` struct |
| `pkg/tencent/client.go` | Create | HTTP client + parser for qt.gtimg.cn |
| `pkg/tencent/client_test.go` | Create | Fixture parsing + code conversion tests |
| `pkg/stockd/services/realtime.go` | Create | Cache CRUD, `isTradingHours`, on-demand fetch, refresh |
| `pkg/stockd/services/realtime_test.go` | Create | `isTradingHours` table tests, cache hit/miss tests |
| `pkg/stockd/services/service.go` | Modify | Add `tc`, `realtimeCache`, `realtimeMu` fields; update `NewService` |
| `pkg/stockd/services/scheduler.go` | Modify | Register realtime refresh cron job in `InitCron` |
| `pkg/stockd/http/quote.go` | Create | `GetQuote` HTTP handler |
| `pkg/stockd/http/router.go` | Modify | Register `GET /api/quotes/:code` |
| `cmd/stockd/main.go` | Modify | Construct `tencent.Client`, pass to `NewService` |
| `web/src/types/api.ts` | Modify | Add `RealtimeQuote` interface |
| `web/src/apis/stocks.ts` | Modify | Add `getQuote(code)` |
| `web/src/components/StockBasicCard.vue` | Modify | Optional `quote` prop; realtime price display with badge |
| `web/src/views/stock/StockDetailView.vue` | Modify | Fetch quote on mount, pass to `StockBasicCard` |
| `web/src/views/StockListView.vue` | Modify | Replace bar-based prices with realtime quotes |

---

## Task 1: RealtimeQuote Model

**Files:**
- Create: `pkg/models/realtime_quote.go`

- [ ] **Step 1: Create the model file**

```go
package models

import "time"

type RealtimeQuote struct {
	TsCode    string    `json:"tsCode"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`      // [3]  当前价
	PrevClose float64   `json:"prevClose"`  // [4]  昨收
	Open      float64   `json:"open"`       // [5]  今开
	Vol       float64   `json:"vol"`        // [6]  成交量（手）
	High      float64   `json:"high"`       // [33] 最高
	Low       float64   `json:"low"`        // [34] 最低
	Amount    float64   `json:"amount"`     // [37] 成交额（万元）
	Change    float64   `json:"change"`     // [31] 涨跌
	ChangePct float64   `json:"changePct"`  // [32] 涨跌%
	LimitUp   float64   `json:"limitUp"`    // [47] 涨停价
	LimitDown float64   `json:"limitDown"`  // [48] 跌停价
	QuoteTime string    `json:"quoteTime"`  // [30] 行情时间（原始字符串）
	UpdatedAt time.Time `json:"updatedAt"`
}
```

- [ ] **Step 2: Verify the build**

```bash
go build ./pkg/models/...
```

Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add pkg/models/realtime_quote.go
git commit -m "feat: add RealtimeQuote model"
```

---

## Task 2: Tencent Finance Client

**Files:**
- Create: `pkg/tencent/client.go`
- Create: `pkg/tencent/client_test.go`

### Tencent API format reference

Request: `GET https://qt.gtimg.cn/q=sh600519,sz000858`

Response (one line per stock, semicolon-terminated):
```
v_sh600519="1~贵州茅台~600519~1780.00~1775.00~1778.00~...~14:58:30~5.00~0.28~1785.00~1760.00~...~12345~5000.00~...~1815.80~1738.60";
```

The content between quotes is `~`-delimited. Key field indices:
```
[1]  name        [3]  price     [4]  prevClose  [5]  open
[6]  vol(手)     [30] quoteTime [31] change      [32] changePct
[33] high        [34] low       [37] amount(万)
[47] limitUp     [48] limitDown
```
Minimum of 49 fields required for a valid record.

- [ ] **Step 1: Write the failing test**

Create `pkg/tencent/client_test.go`:

```go
package tencent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixture has exactly 49 fields (indices 0-48)
const fixtureResponse = `v_sh600519="1~贵州茅台~600519~1780.00~1775.00~1778.00~12345~200~100~1779.00~100~1779.50~200~1779.80~300~1780.00~100~1780.00~50~1780.10~100~1780.20~200~1780.30~300~1780.40~100~1780.50~50~recent~14:58:30~5.00~0.28~1785.00~1760.00~stuff~12345~5000.00~1.23~28.5~field40~1800.00~1600.00~1.50~2000.00~2500.00~5.20~1815.80~1738.60";`

func TestFetchQuotes_ParsesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fixtureResponse))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/q="))
	quotes, err := c.FetchQuotes(context.Background(), []string{"600519.SH"})
	require.NoError(t, err)
	require.Len(t, quotes, 1)

	q := quotes[0]
	assert.Equal(t, "600519.SH", q.TsCode)
	assert.InDelta(t, 1780.00, q.Price, 0.001)
	assert.InDelta(t, 1775.00, q.PrevClose, 0.001)
	assert.InDelta(t, 1778.00, q.Open, 0.001)
	assert.InDelta(t, 12345.0, q.Vol, 0.001)
	assert.InDelta(t, 1785.00, q.High, 0.001)
	assert.InDelta(t, 1760.00, q.Low, 0.001)
	assert.InDelta(t, 5000.00, q.Amount, 0.001)
	assert.InDelta(t, 5.00, q.Change, 0.001)
	assert.InDelta(t, 0.28, q.ChangePct, 0.001)
	assert.InDelta(t, 1815.80, q.LimitUp, 0.001)
	assert.InDelta(t, 1738.60, q.LimitDown, 0.001)
	assert.Equal(t, "14:58:30", q.QuoteTime)
}

func TestTsToCodes(t *testing.T) {
	got := tsToCodes([]string{"600519.SH", "000858.SZ"})
	assert.Equal(t, []string{"sh600519", "sz000858"}, got)
}

func TestTencentToTs(t *testing.T) {
	cases := []struct{ in, want string }{
		{"sh600519", "600519.SH"},
		{"sz000858", "000858.SZ"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tencentToTs(tc.in))
	}
}

func TestFetchQuotes_EmptyInput(t *testing.T) {
	c := NewClient()
	quotes, err := c.FetchQuotes(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, quotes)
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./pkg/tencent/... -run TestFetchQuotes_ParsesFields -v 2>&1 | head -20
```

Expected: compile error (package does not exist yet).

- [ ] **Step 3: Implement the client**

Create `pkg/tencent/client.go`:

```go
// Package tencent fetches intraday quotes from Tencent Finance (qt.gtimg.cn).
package tencent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock/pkg/models"
)

const (
	defaultBaseURL = "https://qt.gtimg.cn/q="

	idxName      = 1
	idxPrice     = 3
	idxPrevClose = 4
	idxOpen      = 5
	idxVol       = 6
	idxQuoteTime = 30
	idxChange    = 31
	idxChangePct = 32
	idxHigh      = 33
	idxLow       = 34
	idxAmount    = 37
	idxLimitUp   = 47
	idxLimitDown = 48
	idxMinFields = 49
)

// Client fetches realtime quotes from Tencent Finance.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures Client.
type Option func(*Client)

// WithBaseURL overrides the base URL (used in tests to point at httptest.Server).
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

// FetchQuotes retrieves realtime quotes for the given tsCode list (e.g. "600519.SH").
// Returns partial results on per-stock parse errors.
func (c *Client) FetchQuotes(ctx context.Context, tsCodes []string) ([]*models.RealtimeQuote, error) {
	if len(tsCodes) == 0 {
		return nil, nil
	}
	url := c.baseURL + strings.Join(tsToCodes(tsCodes), ",")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Referer", "https://finance.qq.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch quotes: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return parseQuotes(body), nil
}

func parseQuotes(body []byte) []*models.RealtimeQuote {
	now := time.Now()
	var out []*models.RealtimeQuote
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "v_") {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}
		tencentCode := line[2:eqIdx] // e.g. "sh600519"

		start := strings.Index(line, `"`)
		end := strings.LastIndex(line, `"`)
		if start < 0 || end <= start {
			continue
		}
		content := line[start+1 : end]
		fields := strings.Split(content, "~")
		if len(fields) < idxMinFields {
			continue
		}
		out = append(out, &models.RealtimeQuote{
			TsCode:    tencentToTs(tencentCode),
			Price:     parseFloat(fields[idxPrice]),
			PrevClose: parseFloat(fields[idxPrevClose]),
			Open:      parseFloat(fields[idxOpen]),
			Vol:       parseFloat(fields[idxVol]),
			High:      parseFloat(fields[idxHigh]),
			Low:       parseFloat(fields[idxLow]),
			Amount:    parseFloat(fields[idxAmount]),
			Change:    parseFloat(fields[idxChange]),
			ChangePct: parseFloat(fields[idxChangePct]),
			LimitUp:   parseFloat(fields[idxLimitUp]),
			LimitDown: parseFloat(fields[idxLimitDown]),
			QuoteTime: fields[idxQuoteTime],
			UpdatedAt: now,
		})
	}
	return out
}

// tsToCodes converts ["600519.SH","000858.SZ"] → ["sh600519","sz000858"].
func tsToCodes(tsCodes []string) []string {
	out := make([]string, 0, len(tsCodes))
	for _, ts := range tsCodes {
		parts := strings.SplitN(ts, ".", 2)
		if len(parts) != 2 {
			continue
		}
		out = append(out, strings.ToLower(parts[1])+parts[0])
	}
	return out
}

// tencentToTs converts "sh600519" → "600519.SH".
func tencentToTs(code string) string {
	if len(code) < 3 {
		return code
	}
	prefix := strings.ToUpper(code[:2])
	num := code[2:]
	return num + "." + prefix
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/tencent/... -v -race
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/tencent/client.go pkg/tencent/client_test.go
git commit -m "feat: add Tencent Finance quote client"
```

---

## Task 3: Service Realtime Layer

**Files:**
- Create: `pkg/stockd/services/realtime.go`
- Create: `pkg/stockd/services/realtime_test.go`

- [ ] **Step 1: Write the failing tests**

Create `pkg/stockd/services/realtime_test.go`:

```go
package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsTradingHours(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")

	cases := []struct {
		name string
		t    time.Time
		want bool
	}{
		{
			name: "weekday at 9:15 (open)",
			t:    time.Date(2026, 5, 12, 9, 15, 0, 0, loc), // Monday
			want: true,
		},
		{
			name: "weekday at 15:00 (close)",
			t:    time.Date(2026, 5, 12, 15, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "weekday at 9:14 (before open)",
			t:    time.Date(2026, 5, 12, 9, 14, 0, 0, loc),
			want: false,
		},
		{
			name: "weekday at 15:01 (after close)",
			t:    time.Date(2026, 5, 12, 15, 1, 0, 0, loc),
			want: false,
		},
		{
			name: "Saturday",
			t:    time.Date(2026, 5, 16, 10, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "Sunday",
			t:    time.Date(2026, 5, 17, 10, 0, 0, 0, loc),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isTradingHours(tc.t))
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test ./pkg/stockd/services/... -run TestIsTradingHours -v 2>&1 | head -10
```

Expected: compile error (function not defined).

- [ ] **Step 3: Implement realtime.go**

Create `pkg/stockd/services/realtime.go`:

```go
package services

import (
	"context"
	"fmt"
	"time"

	"stock/pkg/models"
)

// isTradingHours reports whether t falls within 09:15–15:00 on a weekday in Asia/Shanghai.
func isTradingHours(t time.Time) bool {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := t.In(loc)
	wd := now.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	h, m, _ := now.Clock()
	total := h*60 + m
	return total >= 9*60+15 && total <= 15*60
}

// GetRealtimeQuote returns the cached quote for tsCode, or fetches it on-demand on a cache miss.
func (s *Service) GetRealtimeQuote(ctx context.Context, tsCode string) (*models.RealtimeQuote, error) {
	s.realtimeMu.RLock()
	q, ok := s.realtimeCache[tsCode]
	s.realtimeMu.RUnlock()
	if ok {
		return q, nil
	}

	quotes, err := s.tc.FetchQuotes(ctx, []string{tsCode})
	if err != nil {
		return nil, fmt.Errorf("fetch quote %s: %w", tsCode, err)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote data for %s", tsCode)
	}
	s.fillNames(quotes)
	s.realtimeMu.Lock()
	for _, q := range quotes {
		s.realtimeCache[q.TsCode] = q
	}
	s.realtimeMu.Unlock()
	return quotes[0], nil
}

// refreshRealtimeQuotes batch-fetches all portfolio stocks and updates the cache.
func (s *Service) refreshRealtimeQuotes(ctx context.Context) {
	codes, err := s.DistinctTsCodes(ctx)
	if err != nil {
		s.logger.WithError(err).Error("realtime: get portfolio codes failed")
		return
	}
	if len(codes) == 0 {
		return
	}
	quotes, err := s.tc.FetchQuotes(ctx, codes)
	if err != nil {
		s.logger.WithError(err).Error("realtime: fetch quotes failed")
		return
	}
	s.fillNames(quotes)
	s.realtimeMu.Lock()
	for _, q := range quotes {
		s.realtimeCache[q.TsCode] = q
	}
	s.realtimeMu.Unlock()
}

// fillNames populates Name from the in-memory stock cache (avoids encoding issues with Tencent response).
func (s *Service) fillNames(quotes []*models.RealtimeQuote) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	for _, q := range quotes {
		if info, ok := s.stockCacheByTsCode[q.TsCode]; ok {
			q.Name = info.Name
		}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/stockd/services/... -run TestIsTradingHours -v -race
```

Expected: 6 subtests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/stockd/services/realtime.go pkg/stockd/services/realtime_test.go
git commit -m "feat: add realtime quote cache and trading hours logic"
```

---

## Task 4: Wire Up — Service Fields, Scheduler, main.go

**Files:**
- Modify: `pkg/stockd/services/service.go`
- Modify: `pkg/stockd/services/scheduler.go`
- Modify: `cmd/stockd/main.go`

- [ ] **Step 1: Update service.go — add fields and update NewService**

Open `pkg/stockd/services/service.go`. Replace the entire file content:

```go
package services

import (
	"stock/pkg/models"
	"stock/pkg/stockd/config"
	"stock/pkg/tencent"
	"stock/pkg/tushare"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	ts     *tushare.Client
	tc     *tencent.Client
	cfg    *config.Config
	cron   *cron.Cron
	mu     sync.Mutex
	jobs   map[string]JobFunc
	sf     singleflight.Group
	logger *logrus.Logger

	stockCacheByCode   map[string]*models.Stock
	stockCacheByTsCode map[string]*models.Stock
	cacheMu            sync.RWMutex

	realtimeCache map[string]*models.RealtimeQuote
	realtimeMu    sync.RWMutex
}

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
		realtimeCache: make(map[string]*models.RealtimeQuote),
	}
}

func (s *Service) GetDB() *gorm.DB         { return s.db }
func (s *Service) GetTS() *tushare.Client  { return s.ts }
func (s *Service) GetConfig() *config.Config { return s.cfg }
```

- [ ] **Step 2: Update scheduler.go — add realtime cron job**

Open `pkg/stockd/services/scheduler.go`. Add the realtime job at the **end of `InitCron`**, just before the final `return nil`:

```go
	if _, err := s.cron.AddFunc("@every 30s", func() {
		if !isTradingHours(time.Now()) {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		s.refreshRealtimeQuotes(ctx)
	}); err != nil {
		return err
	}
	return nil
```

The complete `InitCron` function should look like this after the edit:

```go
func (s *Service) InitCron() error {
	err := s.RegisterCron("daily-fetch", s.cfg.Scheduler.DailyFetchCron, func(ctx context.Context) error {
		codes, err := s.DistinctTsCodes(ctx)
		if err != nil {
			return err
		}
		for _, code := range codes {
			s.logger.Infof("sync %s ", code)
			if _, err := s.SyncDaily(ctx, s.cfg.Tushare.GetDefaultToken(""), code); err != nil {
				s.logger.WithError(err).WithField("ts_code", code).Error("daily sync failed")
			}
		}
		for _, code := range codes {
			_, err := s.recalcStock(ctx, code)
			if err != nil {
				s.logger.WithError(err).WithField("ts_code", code).Error("recalcStock sync failed")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = s.RegisterCron("stocklist-sync", s.cfg.Scheduler.StocklistSyncCron, func(ctx context.Context) error {
		_, err := s.SyncFromTushare(ctx, s.cfg.Tushare.DefaultToken)
		return err
	})
	if err != nil {
		return err
	}
	if _, err := s.cron.AddFunc("@every 30s", func() {
		if !isTradingHours(time.Now()) {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		s.refreshRealtimeQuotes(ctx)
	}); err != nil {
		return err
	}
	return nil
}
```

Also add `"time"` to the imports of `scheduler.go` if it is not already present.

- [ ] **Step 3: Update main.go — construct Tencent client and pass to NewService**

Open `cmd/stockd/main.go`. Find these two lines:

```go
	tc := tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL))
	svc := services.NewService(gdb, tc, cfg, logger)
```

Replace them with:

```go
	tc := tushare.NewClient(tushare.WithBaseURL(cfg.Tushare.BaseURL))
	tencentClient := tencent.NewClient()
	svc := services.NewService(gdb, tc, tencentClient, cfg, logger)
```

Add the import at the top of the file:

```go
	"stock/pkg/tencent"
```

- [ ] **Step 4: Build and run existing tests**

```bash
go build ./... && go test -race ./pkg/stockd/... 2>&1 | tail -20
```

Expected: build succeeds, existing tests still pass.

- [ ] **Step 5: Commit**

```bash
git add pkg/stockd/services/service.go pkg/stockd/services/scheduler.go cmd/stockd/main.go
git commit -m "feat: wire tencent client into service and add realtime cron job"
```

---

## Task 5: HTTP Handler and Route

**Files:**
- Create: `pkg/stockd/http/quote.go`
- Modify: `pkg/stockd/http/router.go`

- [ ] **Step 1: Create the handler**

Create `pkg/stockd/http/quote.go`:

```go
package http

import (
	"github.com/gin-gonic/gin"

	"stock/pkg/stockd/utils"
	utilspkg "stock/pkg/utils"
)

func (h *handler) GetQuote(c *gin.Context) {
	tsCode := utilspkg.ToTushareCode(c.Param(codeValue))
	q, err := h.svc.GetRealtimeQuote(c.Request.Context(), tsCode)
	if err != nil {
		utils.HTTPRequestFailedV5(c, err)
		return
	}
	utils.HTTPRequestSuccess(c, 200, q)
}
```

- [ ] **Step 2: Register the route**

Open `pkg/stockd/http/router.go`. After the line:

```go
	api.GET("/stocks/"+codeUrl, h.GetStock)
```

Add:

```go
	api.GET("/quotes/"+codeUrl, h.GetQuote)
```

- [ ] **Step 3: Build to verify no compile errors**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add pkg/stockd/http/quote.go pkg/stockd/http/router.go
git commit -m "feat: add GET /api/quotes/:code endpoint"
```

---

## Task 6: Frontend — Types and API Client

**Files:**
- Modify: `web/src/types/api.ts`
- Modify: `web/src/apis/stocks.ts`

- [ ] **Step 1: Add RealtimeQuote to api.ts**

Open `web/src/types/api.ts`. Append at the end of the file:

```typescript
export interface RealtimeQuote {
  tsCode: string
  name: string
  price: number
  prevClose: number
  open: number
  vol: number       // 手
  high: number
  low: number
  amount: number    // 万元
  change: number
  changePct: number
  limitUp: number
  limitDown: number
  quoteTime: string
  updatedAt: string
}
```

- [ ] **Step 2: Add getQuote to stocks.ts**

Open `web/src/apis/stocks.ts`. Add the import and function.

The updated file should look like:

```typescript
import type { Stock, DailyBar, PageResult, RealtimeQuote } from '@/types/api'
import { $http } from './axios'

export const searchStocks = (q: string, limit = 20): Promise<Stock[]> =>
  $http.get('/stocks', { params: { q, limit } }) as any

export const getStock = (code: string): Promise<Stock> => $http.get(`/stocks/${code}`) as any

export const queryBars = (
  code: string,
  params?: { from?: string; to?: string; page?: number; limit?: number },
): Promise<PageResult<DailyBar>> => $http.get(`/bars/${code}`, { params }) as any

export const getQuote = (code: string): Promise<RealtimeQuote> =>
  $http.get(`/quotes/${code}`) as any
```

- [ ] **Step 3: Run TypeScript type check**

```bash
cd web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add web/src/types/api.ts web/src/apis/stocks.ts
git commit -m "feat: add RealtimeQuote type and getQuote API"
```

---

## Task 7: StockBasicCard — Realtime Price Display

**Files:**
- Modify: `web/src/components/StockBasicCard.vue`

The component currently shows `lastBar.close` and a percent change calculated from `prevClose`. When a `quote` prop is provided, it should show `quote.price` and `quote.changePct` instead.

- [ ] **Step 1: Update StockBasicCard.vue**

Replace the entire file with:

```vue
<script setup lang="ts">
import { computed } from 'vue'
import type { Stock, DailyBar, RealtimeQuote } from '@/types/api'
import { fmtPrice, fmtPct, priceClass } from '@/utils/format'

const props = defineProps<{
  stock: Stock
  lastBar?: DailyBar
  prevClose?: number
  quote?: RealtimeQuote
}>()
const emit = defineEmits<{ back: [] }>()

const displayPrice = computed(() => {
  if (props.quote) return props.quote.price
  return props.lastBar?.close ?? 0
})

const changeClass = computed(() => {
  if (props.quote) {
    return props.quote.changePct > 0 ? 'g-up' : props.quote.changePct < 0 ? 'g-down' : 'g-flat'
  }
  if (!props.lastBar) return 'g-flat'
  const base = props.prevClose || props.lastBar.open
  return priceClass(props.lastBar.close, base)
})

const changePct = computed(() => {
  if (props.quote) {
    const sign = props.quote.changePct > 0 ? '+' : ''
    return `${sign}${props.quote.changePct.toFixed(2)}%`
  }
  if (!props.lastBar) return '--'
  const base = props.prevClose || props.lastBar.open
  return fmtPct(props.lastBar.close, base)
})

function fmtDate(d: string): string {
  if (d?.length === 8) return `${d.slice(0, 4)}-${d.slice(4, 6)}-${d.slice(6)}`
  return d
}
</script>

<template>
  <el-card body-style="padding: 10px 16px">
    <div class="stock-row">
      <el-button link size="small" class="back-btn" @click="emit('back')">←</el-button>
      <span class="name">{{ stock.name }}</span>
      <span class="code">{{ stock.tsCode }}</span>
      <el-tag size="small" type="info">{{ stock.industry }}</el-tag>
      <span class="list-date">上市 {{ fmtDate(stock.listDate) }}</span>
      <template v-if="displayPrice">
        <!-- push price group to the right; badge sits before price -->
        <el-tag v-if="quote" size="small" type="success" class="realtime-badge">实时</el-tag>
        <span class="price" :class="[changeClass, { 'price--no-badge': !quote }]">
          {{ fmtPrice(displayPrice) }}
        </span>
        <span class="pct" :class="changeClass">{{ changePct }}</span>
      </template>
    </div>
    <div v-if="quote" class="quote-row">
      <span>开 {{ fmtPrice(quote.open) }}</span>
      <span>高 <span class="g-up">{{ fmtPrice(quote.high) }}</span></span>
      <span>低 <span class="g-down">{{ fmtPrice(quote.low) }}</span></span>
      <span>量 {{ (quote.vol / 10000).toFixed(2) }}万手</span>
      <span>额 {{ (quote.amount / 10000).toFixed(2) }}亿</span>
    </div>
  </el-card>
</template>

<!-- g-up / g-down / g-flat are defined globally in index.scss -->
<style scoped lang="scss">
.stock-row {
  display: flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
}
.quote-row {
  display: flex;
  gap: 16px;
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.back-btn {
  padding: 0;
  color: #909399;
  font-size: 15px;
}
.name {
  font-size: 16px;
  font-weight: 600;
}
.code {
  font-size: 13px;
  color: #909399;
}
.list-date {
  font-size: 12px;
  color: #909399;
}
/* Push price group to the right side */
.realtime-badge {
  margin-left: auto;
}
.price {
  font-size: 20px;
  font-weight: bold;
}
/* When no badge is shown, price itself takes the auto margin */
.price--no-badge {
  margin-left: auto;
}
.pct {
  font-size: 13px;
}
</style>
```

- [ ] **Step 2: Run TypeScript type check**

```bash
cd web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/StockBasicCard.vue
git commit -m "feat: show realtime quote in StockBasicCard"
```

---

## Task 8: StockDetailView — Fetch Quote on Mount

**Files:**
- Modify: `web/src/views/stock/StockDetailView.vue`

- [ ] **Step 1: Update StockDetailView.vue**

Replace the entire file with:

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { getStock, queryBars, getQuote } from '@/apis/stocks'
import type { Stock, DailyBar, RealtimeQuote } from '@/types/api'
import StockBasicCard from '@/components/StockBasicCard.vue'

const route = useRoute()
const router = useRouter()
const code = route.params.code as string

const stock = ref<Stock | null>(null)
const lastBar = ref<DailyBar | undefined>(undefined)
const prevClose = ref(0)
const quote = ref<RealtimeQuote | null>(null)
const loading = ref(false)

const activeTab = computed(() => {
  if (route.path.endsWith('/bars')) return 'bars'
  if (route.path.endsWith('/predictions')) return 'predictions'
  return 'prediction'
})

function onTabClick(paneName: string) {
  if (paneName === 'bars') router.push(`/stocks/${code}/bars`)
  else if (paneName === 'predictions') router.push(`/stocks/${code}/predictions`)
  else router.push(`/stocks/${code}`)
}

onMounted(async () => {
  loading.value = true
  try {
    const [s, barsPage, q] = await Promise.all([
      getStock(code),
      queryBars(code, { limit: 2 }),
      getQuote(code).catch(() => null),
    ])
    stock.value = s
    lastBar.value = barsPage.items[0]
    prevClose.value = barsPage.items[1]?.close ?? 0
    quote.value = q
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-loading="loading">
    <div v-if="stock">
      <StockBasicCard
        :stock="stock"
        :last-bar="lastBar"
        :prev-close="prevClose"
        :quote="quote ?? undefined"
        @back="router.push('/stocks')"
      />
    </div>
    <el-tabs
      :model-value="activeTab"
      style="margin-top: 8px"
      @tab-click="(tab: any) => onTabClick(tab.paneName)"
    >
      <el-tab-pane label="预测" name="prediction" />
      <el-tab-pane label="日K数据" name="bars" />
      <el-tab-pane label="预测记录" name="predictions" />
    </el-tabs>
    <router-view />
  </div>
</template>

<style scoped lang="scss">
.detail-header {
  display: flex;
  align-items: center;
}
</style>
```

- [ ] **Step 2: Run TypeScript type check**

```bash
cd web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/views/stock/StockDetailView.vue
git commit -m "feat: fetch realtime quote in stock detail view"
```

---

## Task 9: StockListView — Replace Bar Prices with Realtime Quotes

**Files:**
- Modify: `web/src/views/StockListView.vue`

The portfolio list currently uses `queryBars` to get the last two bars and computes change%. Replace this with `getQuote` calls for each portfolio item.

- [ ] **Step 1: Update StockListView.vue**

Replace the `<script setup>` block (lines 1–103) with:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { usePortfolioStore } from '@/stores/portfolio'
import { searchStocks, getQuote } from '@/apis/stocks'
import type { Stock, Portfolio, RealtimeQuote } from '@/types/api'

const router = useRouter()
const portfolioStore = usePortfolioStore()
const showAdd = ref(false)
const selectedCode = ref('')
const note = ref('')
const stockOptions = ref<{ value: string; label: string }[]>([])
const loadingAdd = ref(false)

const quoteMap = ref<Record<string, RealtimeQuote>>({})
const loadingQuotes = ref(false)

async function loadQuotes(items: Portfolio[]) {
  if (!items.length) return
  loadingQuotes.value = true
  try {
    const results = await Promise.allSettled(items.map(p => getQuote(p.code)))
    results.forEach((res, i) => {
      if (res.status === 'fulfilled') {
        quoteMap.value[items[i].code] = res.value
      }
    })
  } finally {
    loadingQuotes.value = false
  }
}

onMounted(async () => {
  await portfolioStore.fetch()
  loadQuotes(portfolioStore.items)
})

async function searchStockOptions(query: string) {
  if (!query) { stockOptions.value = []; return }
  try {
    const list = await searchStocks(query, 20)
    stockOptions.value = list.map((s: Stock) => ({ value: s.code, label: `${s.code} ${s.name}` }))
  } catch {
    stockOptions.value = []
  }
}

async function doAdd() {
  if (!selectedCode.value) { wMessage('warning', '请选择股票'); return }
  loadingAdd.value = true
  try {
    await portfolioStore.add(selectedCode.value, note.value)
    wMessage('success', '添加成功')
    showAdd.value = false
    selectedCode.value = ''
    note.value = ''
    loadQuotes(portfolioStore.items)
  } finally {
    loadingAdd.value = false
  }
}

function goDetail(row: Portfolio) {
  router.push(`/stocks/${row.code}`)
}

function removeItem(row: Portfolio) {
  portfolioStore.remove(row.code)
}

function fmtPrice(v: number): string {
  return v ? v.toFixed(2) : '-'
}

function pctClass(v: number): string {
  if (v > 0) return 'up'
  if (v < 0) return 'down'
  return ''
}

function fmtPct(v: number): string {
  if (!v) return '-'
  return (v > 0 ? '+' : '') + v.toFixed(2) + '%'
}
</script>
```

Then update the `<template>` block. Replace the `<el-table>` section to use `quoteMap` instead of `barMap`:

```vue
<template>
  <div>
    <div class="header-bar">
      <h2>{{ $t('stockList.title') }}</h2>
      <el-button type="primary" @click="showAdd = true">{{ $t('stockList.addStock') }}</el-button>
    </div>

    <el-table :data="portfolioStore.items" v-loading="loadingQuotes" style="margin-top: 16px">
      <el-table-column prop="code" :label="$t('stockList.code')" width="100">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">{{ row.code }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="name" :label="$t('stockList.name')" width="120" />
      <el-table-column label="开盘" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(quoteMap[row.code]?.open ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="最高" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(quoteMap[row.code]?.high ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="最低" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(quoteMap[row.code]?.low ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="现价" align="right" width="80">
        <template #default="{ row }">
          <span :class="pctClass(quoteMap[row.code]?.changePct ?? 0)">
            {{ fmtPrice(quoteMap[row.code]?.price ?? 0) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="涨跌幅" align="right" width="90">
        <template #default="{ row }">
          <span :class="pctClass(quoteMap[row.code]?.changePct ?? 0)">
            {{ fmtPct(quoteMap[row.code]?.changePct ?? 0) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="note" :label="$t('stockList.note')" />
      <el-table-column :label="$t('stockList.action')" width="120">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">{{ $t('stockList.detail') }}</el-button>
          <el-button link type="danger" @click="removeItem(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAdd" :title="$t('stockList.addStock')" width="400px">
      <el-form @submit.prevent="doAdd">
        <el-form-item :label="$t('stockList.code')">
          <el-select-v2
            v-model="selectedCode"
            :options="stockOptions"
            :placeholder="$t('stockList.selectStock')"
            clearable filterable remote
            :remote-method="searchStockOptions"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('stockList.note')">
          <el-input v-model="note" :placeholder="$t('common.empty')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loadingAdd" @click="doAdd">{{ $t('common.add') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.up   { color: var(--el-color-danger); }
.down { color: var(--el-color-success); }
</style>
```

- [ ] **Step 2: Run TypeScript type check**

```bash
cd web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 3: Run all Go tests**

```bash
cd /root/code/github/travelliu/stock && go test -race ./... 2>&1 | tail -20
```

Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add web/src/views/StockListView.vue
git commit -m "feat: show realtime quotes in portfolio list"
```

---

## Done

At this point:
- `GET /api/quotes/:code` returns a cached or on-demand realtime quote from Tencent Finance
- The cron job refreshes portfolio stocks every 30s during 09:15–15:00 weekdays
- `StockBasicCard` shows a "实时" badge and live open/high/low/vol/amount row when quote data is available
- `StockListView` (portfolio) shows realtime price and change% instead of historical bars
