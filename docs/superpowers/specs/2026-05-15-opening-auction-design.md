# Opening Auction (集合竞价) Feature Design

**Goal:** Fetch daily opening auction data for all A-share stocks via Tushare `stk_auction`, store it, and expose a signal endpoint that surfaces stocks whose today's auction volume is ≥ 2× their 15-day average.

**Architecture:** New Tushare fetcher → new model/table → new service methods → new HTTP handlers → new scheduler job → new Web UI page.

**Tech Stack:** Go (GORM, Gin, robfig/cron), Vue 3 + Element Plus, Tushare `stk_auction` API.

---

## Data Model

File: `pkg/models/auction_bar.go`

```go
type AuctionBar struct {
    TsCode       string  `gorm:"primaryKey;size:16" json:"tsCode"`
    TradeDate    string  `gorm:"primaryKey;size:8"  json:"tradeDate"`
    Vol          int64   `json:"vol"`
    Price        float64 `json:"price"`
    Amount       float64 `json:"amount"`
    PreClose     float64 `json:"preClose"`
    TurnoverRate float64 `json:"turnoverRate"`
    VolumeRatio  float64 `json:"volumeRatio"`
    FloatShare   float64 `json:"floatShare"`
}
```

Primary key: `(ts_code, trade_date)` — upsert-safe.

Added to `db.AutoMigrate(...)`.

---

## Tushare Fetcher

File: `pkg/tushare/auction.go`

```go
type AuctionRequest struct {
    TsCode    string // optional: filter by stock
    TradeDate string // YYYYMMDD — fetch all stocks for this date
    StartDate string
    EndDate   string
}

func StkAuction(ctx context.Context, c *Client, token string, req AuctionRequest) ([]models.AuctionBar, error)
```

Fields requested: `ts_code,trade_date,vol,price,amount,pre_close,turnover_rate,volume_ratio,float_share`

Single call with `trade_date` returns all ~5500 A-share stocks (well within the 8000-row limit).

---

## Service Layer

File: `pkg/stockd/services/auction.go`

### `SyncAuction(ctx context.Context, date string) (int, error)`

- Calls `tushare.StkAuction` with `trade_date=date`
- Bulk upserts rows into `auction_bars` via `ON DUPLICATE KEY UPDATE` (MySQL) / `ON CONFLICT DO UPDATE` (SQLite/Postgres)
- Returns count of rows upserted

### `InitAuction(ctx context.Context) error`

- Computes the last 15 trading days (skip weekends; simple approximation: go back up to 25 calendar days and collect 15 weekdays)
- Calls `SyncAuction` for each date in ascending order
- Skips dates that already have data (count > 0)

### `ListAuctionSignals(ctx, date string, minRatio float64, page, limit int) (*PageResult[AuctionSignal], error)`

Signal query — finds stocks where today `vol ≥ avg(last 15 days) * minRatio`:

```sql
SELECT
    a.ts_code,
    s.name,
    a.vol          AS today_vol,
    AVG(h.vol)     AS avg_vol,
    a.vol / AVG(h.vol) AS ratio,
    a.price,
    a.amount
FROM auction_bars a
LEFT JOIN stocks s ON s.ts_code = a.ts_code
JOIN auction_bars h
    ON h.ts_code = a.ts_code
    AND h.trade_date < a.trade_date
    AND h.trade_date >= (15 prior trading days)
WHERE a.trade_date = :date
GROUP BY a.ts_code
HAVING avg_vol > 0 AND ratio >= :min_ratio
ORDER BY ratio DESC
```

Return type:

```go
type AuctionSignal struct {
    TsCode    string  `json:"tsCode"`
    Name      string  `json:"name"`
    TradeDate string  `json:"tradeDate"`
    TodayVol  int64   `json:"todayVol"`
    AvgVol    float64 `json:"avgVol"`
    Ratio     float64 `json:"ratio"`
    Price     float64 `json:"price"`
    Amount    float64 `json:"amount"`
}
```

---

## HTTP API

File: `pkg/stockd/http/auction.go`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/admin/auction/init` | admin | Backfill last 15 trading days |
| `POST` | `/api/admin/auction/sync` | admin | Sync a specific date (body: `{"date":"20260515"}`, defaults to today) |
| `GET`  | `/api/auction/signals` | user | Query surge signals |

`GET /api/auction/signals` query params:

| Param | Default | Description |
|-------|---------|-------------|
| `date` | today | YYYYMMDD |
| `min_ratio` | `2` | Minimum vol/avg multiplier |
| `page` | `1` | Page number |
| `limit` | `50` | Page size |

Routes added in `router.go`:
- Admin routes under existing `adm` group
- Signal route under authenticated `api` group

---

## Scheduler

`SchedulerConfig` gains a new field:

```go
AuctionFetchCron string `mapstructure:"auction_fetch_cron"`
```

Default: `"26 9 * * 1-5"` (9:26 on weekdays — data available from 9:25).

New job registered in `InitCron()`:

```go
s.RegisterCron("auction-fetch", s.cfg.Scheduler.AuctionFetchCron, func(ctx context.Context) error {
    today := time.Now().Format("20060102")
    _, err := s.SyncAuction(ctx, today)
    return err
})
```

`config.example.yaml` updated with:

```yaml
scheduler:
  auction_fetch_cron: "26 9 * * 1-5"
```

---

## Web UI

### Menu

`ConsoleMenu.vue`: add a "竞价" menu item (icon: `Histogram`) linking to `/auction`.

### Route

`router/index.ts`: add `{ path: '/auction', component: AuctionView, meta: { requiresAuth: true } }`.

### View: `AuctionView.vue`

- **Header controls:** date picker (default today), min-ratio input (default 2), search/refresh button
- **Table columns:** 代码 (link to stock detail), 名称, 今日竞价量, 15日均量, 倍数 (highlighted red if ≥ 2), 价格, 成交额
- **Pagination:** standard Element Plus pagination component
- Calls `GET /api/auction/signals`

### API type: `web/src/types/api.ts`

```typescript
export interface AuctionSignal {
  tsCode: string
  name: string
  tradeDate: string
  todayVol: number
  avgVol: number
  ratio: number
  price: number
  amount: number
}
```

### API client: `web/src/apis/auction.ts`

```typescript
export function listAuctionSignals(params: {
  date?: string
  minRatio?: number
  page?: number
  limit?: number
}): Promise<PageResult<AuctionSignal>>
```

---

## Error Handling

- `SyncAuction` with a non-trading day returns 0 rows → log and return `(0, nil)` (not an error)
- `InitAuction` skips dates that already have data (idempotent)
- Admin endpoints return `JobRun`-style response (status + message) consistent with existing `/api/admin/bars/sync`

---

## Out of Scope

- Alerts / push notifications for volume surges
- Intraday tick data
- `stk_auction_o` / `stk_auction_c` (closing auction) — separate feature if needed
