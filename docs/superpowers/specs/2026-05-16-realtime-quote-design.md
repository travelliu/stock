# Realtime Quote Feature Design

**Date:** 2026-05-16  
**Status:** Approved

## Overview

Add intraday realtime price data to the stock analysis platform by polling Tencent Finance's public quote API (`qt.gtimg.cn`). The backend caches quotes in memory, refreshing every 30 seconds during trading hours. The frontend displays realtime price on the stock detail page and portfolio page.

---

## Architecture

```
Tencent Finance API (qt.gtimg.cn)
        ↑  every 30s, trading hours only, batch by portfolio stocks
        |
pkg/tencent/client.go   — HTTP client + response parser
        ↓
services.Service
  ├── realtimeCache  map[string]*models.RealtimeQuote  (RWMutex)
  ├── cron job: every 30s, checks trading hours before fetching
  └── GetRealtimeQuote(tsCode) → cache hit or on-demand fetch
        ↓
GET /api/quotes/:code   — no auth required (public market data)
        ↓
Frontend: StockDetailView + Portfolio page
```

**Key decisions:**

- Polling scope: all distinct `tsCode` values across all users' portfolios. Non-portfolio stocks are fetched on-demand (first request only per trading day).
- Trading hours: 09:15–15:00 Asia/Shanghai, Monday–Friday.
- Cache retention: no TTL eviction. Data persists until the next refresh. After market close, the final price of the day remains available.
- The cron job runs every 30 seconds using robfig/cron with second-level precision (`*/30 * * * * *`), and checks trading hours inside the job body.

---

## Data Model

### Go — `pkg/models/realtime_quote.go`

```go
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
    QuoteTime string    `json:"quoteTime"`  // [30] 行情时间 (原始字符串)
    UpdatedAt time.Time `json:"updatedAt"`  // 服务器拉取时间
}
```

### TypeScript — addition to `web/src/types/api.ts`

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

---

## Tencent Finance Client — `pkg/tencent/`

### Files

- `pkg/tencent/client.go` — HTTP client, fetch + parse
- `pkg/tencent/client_test.go` — unit tests with fixture response

### API

```
GET https://qt.gtimg.cn/q=sh600519,sz000858
```

Response (one line per stock):

```
v_sh600519="1~贵州茅台~600519~1780.00~1775.00~1780.00~12345~...~09:15:00~5.00~0.28~1785.00~1760.00~...";
```

### Parsing rules

1. Split response by `\n`, skip empty lines.
2. Each line: extract content between `"..."`.
3. Split by `~`, access by index constant.
4. Convert tsCode from Tencent format back to internal format:
   - `sh600519` → `600519.SH`
   - `sz000858` → `000858.SZ`

### Field index constants

```go
const (
    idxName      = 1
    idxPrice     = 3
    idxPrevClose = 4
    idxOpen      = 5
    idxVol       = 6
    idxChange    = 31
    idxChangePct = 32
    idxHigh      = 33
    idxLow       = 34
    idxAmount    = 37
    idxLimitUp   = 47
    idxLimitDown = 48
    idxQuoteTime = 30
    idxMinFields = 49   // minimum fields required for a valid record
)
```

### Code conversion

```go
// tsToCodes converts ["600519.SH", "000858.SZ"] → ["sh600519", "sz000858"]
func tsToCodes(tsCodes []string) []string

// tencentToTs converts "sh600519" → "600519.SH"
func tencentToTs(tencentCode string) string
```

The project already has `pkg/utils/stockcode.go`; check for reusable helpers before duplicating logic.

### HTTP behavior

- Single request per batch (all codes comma-separated in one URL).
- 5-second context timeout per request.
- On HTTP error or parse failure, log and return partial results (don't fail the whole batch).

---

## Service Layer Changes

### `services.Service` — new fields

```go
realtimeCache map[string]*models.RealtimeQuote  // key: tsCode (e.g. "600519.SH")
realtimeMu   sync.RWMutex
tc           *tencent.Client
```

### `NewService` signature update

```go
func NewService(db *gorm.DB, ts *tushare.Client, tc *tencent.Client,
    cfg *config.Config, logger *logrus.Logger) *Service
```

The Tencent client is constructed in `cmd/stockd/main.go` (or equivalent entry point) and injected.

### New methods — `services/realtime.go`

```go
// GetRealtimeQuote returns cached quote or fetches on-demand if cache miss.
func (s *Service) GetRealtimeQuote(ctx context.Context, tsCode string) (*models.RealtimeQuote, error)

// refreshRealtimeQuotes fetches all portfolio stocks and updates the cache.
func (s *Service) refreshRealtimeQuotes(ctx context.Context)

// isTradingHours returns true if now is 09:15–15:00 Asia/Shanghai on a weekday.
func isTradingHours() bool

// portfolioTsCodes returns distinct tsCode values from all portfolios.
func (s *Service) portfolioTsCodes(ctx context.Context) ([]string, error)
```

### Scheduler — `services/scheduler.go`

Register cron job with second-level precision:

```go
s.cron.AddFunc("*/30 * * * * *", func() {
    if !isTradingHours() {
        return
    }
    ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
    defer cancel()
    s.refreshRealtimeQuotes(ctx)
})
```

---

## HTTP Layer

### New route — `pkg/stockd/http/router.go`

```go
api.GET("/quotes/"+codeUrl, h.GetQuote)  // no auth, consistent with /api/stocks/:code
```

### New handler — `pkg/stockd/http/quote.go`

```go
func (h *handler) GetQuote(c *gin.Context) {
    code := c.Param(codeValue)
    tsCode := utils.ToTushareCode(code)  // normalize to tsCode format, e.g. "600519" → "600519.SH"
    q, err := h.svc.GetRealtimeQuote(c.Request.Context(), tsCode)
    if err != nil {
        utils.HTTPRequestFailedV5(c, err)
        return
    }
    utils.HTTPRequestSuccess(c, 200, q)
}
```

---

## Frontend Changes

### `web/src/apis/stocks.ts`

```typescript
export const getQuote = (code: string): Promise<RealtimeQuote> =>
  $http.get(`/quotes/${code}`) as any
```

### `web/src/views/stock/StockDetailView.vue`

- Import `getQuote` and `RealtimeQuote`.
- Add `const quote = ref<RealtimeQuote | null>(null)`.
- In `onMounted`, fetch quote in parallel with existing calls.
- Pass `quote` to `StockBasicCard` as new optional prop.

### `web/src/components/StockBasicCard.vue`

- Add optional prop `quote?: RealtimeQuote`.
- When `quote` is present: show `quote.price`, `quote.changePct`, `quote.open`, `quote.high`, `quote.low`, `quote.vol`, `quote.amount`.
- When `quote` is absent: fall back to `lastBar.close` (current behavior, unchanged).
- Add a small "实时" badge when displaying realtime data.

### Portfolio page

- The portfolio page currently shows a list of holdings (`StockListView` or a portfolio-specific view).
- For each holding, call `getQuote(item.code)` (can be batched with `Promise.all`).
- Display current price and change% inline.

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Tencent API unreachable | Log warning, return cache if available; 404 if no cache |
| Market closed, no cache | Return `null` quote; frontend falls back to last bar |
| Partial parse failure | Skip that stock, log field count mismatch |
| Batch too large | Tencent accepts ~100 codes per request; chunk if needed (not expected for typical portfolio size) |

---

## Testing

### Backend

- `pkg/tencent/client_test.go`: parse fixture response string, verify all field extractions.
- `services/realtime_test.go`: mock tencent client, test cache hit/miss/refresh logic and `isTradingHours`.

### Frontend

- Unit test `StockBasicCard` renders realtime badge when quote prop is present.
- Unit test fallback to `lastBar` when quote is absent.

---

## Out of Scope

- Bid/ask orderbook display (fields [9]–[28]).
- 52-week high/low (fields [41], [42]).
- WebSocket / SSE push (not needed given polling approach).
- Persisting realtime quotes to DB.
