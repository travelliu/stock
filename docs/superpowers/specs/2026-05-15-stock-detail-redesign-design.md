# Stock Detail Page Redesign & Bug Fixes

Date: 2026-05-15

## Overview

Redesign the stock detail page, fix navigation bugs, change code format from tsCode to short code, and add backend caching.

## Layer 1: Backend Infrastructure

### 1.1 Short Code Support

All backend APIs that accept `tsCode` (e.g., `300476.SZ`) must also accept short codes (e.g., `300476`).

**Mapping function** `resolveTsCode(code string) (string, error)`:
- Input contains `.` → use as-is (backward compatible)
- Input is digits only → look up in stock cache → return `ts_code`
- Not found → return 404

**Affected endpoints** (path params change from `:tsCode` to `:code`):
- `GET /api/stocks/:code`
- `GET /api/bars/:code`
- `GET /api/analysis/:code`
- `GET /api/analysis/predictions/:code`
- `POST /api/portfolio` (body field)
- `DELETE /api/portfolio/:code`
- `PATCH /api/portfolio/:code`

Portfolio request body changes from `{ "ts_code": "300476.SZ" }` to `{ "code": "300476" }`.

### 1.2 Stock Info Cache

In-memory cache using `sync.RWMutex` + `map[string]StockInfo`.

```
StockInfo {
    TsCode   string  // "300476.SZ"
    Name     string  // "千方科技"
    Industry string
}
```

- Loaded from DB on service startup
- Key: short code (e.g., "300476")
- Refreshed after admin sync/import operations
- No TTL needed — data rarely changes

### 1.3 Session Persistence

Change Gin session cookie from session-only to persistent:
- `MaxAge: 86400 * 7` (7 days)
- `HttpOnly: true`
- `SameSite: Lax`

## Layer 2: Frontend Routing & Navigation

### 2.1 Route Format Change

Route params: `:tsCode` → `:code`.

```
Before: /stocks/300476.SZ
After:  /stocks/300476
```

All `router.push` calls and `<router-link>` components use short codes.

API layer: all functions accept short code, send short code to backend.

### 2.2 Direct URL Access Fix

**Root cause**: auth guard redirects before `fetchMe()` completes on page load.

**Fix**:
- In App.vue, await `auth.fetchMe()` before mounting router
- In auth guard, if auth check is still loading, wait instead of redirecting
- Ensure backend serves `index.html` for all `/stocks/*` paths (SPA fallback)

### 2.3 Back Button on Stock Detail

Add a back arrow (`← 返回列表`) at the top-left of StockDetailView, above StockBasicCard.
Uses `router.push('/stocks')` or `router.back()`.

### 2.4 Stock List Shows Names

Portfolio table adds a "名称" (Name) column.
Backend `GET /api/portfolio` must include stock name in response (join with stocks table if needed).

## Layer 3: Stock Detail UI Restructuring

### 3.1 Tab Structure

Remove: BasicTab, StatisticsTab.
Add: PredictionTab, DailyBarsTab, PredictionRecordsTab.

Routes:
```
/stocks/:code                    → StockDetailView (layout: back button + StockBasicCard + tabs)
/stocks/:code                    → PredictionTab (default)
/stocks/:code/bars               → DailyBarsTab
/stocks/:code/predictions        → PredictionRecordsTab
```

### 3.2 PredictionTab

Three sections stacked vertically:

**Section 1 — Live Price & Prediction**
- Real open price: auto-filled from today's bar, editable by user
- Button: "计算预测" triggers `POST /api/analysis/recalc?code=xxx`
- Real HLC: from latest bar data (close may be empty during trading hours)
- Predicted HLC: from `GET /api/analysis/predictions/:code`

**Section 2 — Spread Model Table**
- Reuse existing `SpreadModelTable` component
- Shows mean/median spread stats across time windows

**Section 3 — Spread Analysis & Distribution**
- Reuse existing `SpreadHistogram` component
- Shows spread distribution histogram

### 3.3 DailyBarsTab

- Reuse `DailyBarTable` with pagination
- Backend `GET /api/bars/:code` adds `page` and `limit` query params
- Default: 20 records per page

### 3.4 PredictionRecordsTab (Backtest)

- Table columns: date, open, predicted high/low/close, actual high/low/close
- Uses existing `GET /api/analysis/predictions/:code` with pagination
- Default: 20 records per page

## Summary of File Changes

### Backend (Go)
- `pkg/stockd/services/stock/stock.go` — add cache, add `resolveTsCode`
- `pkg/stockd/http/router.go` — update route params
- `pkg/stockd/http/handler.go` — update handler param parsing
- `pkg/stockd/http/prediction.go` — update param parsing
- `pkg/stockd/services/stock/stock.go` — add cache init and refresh

### Frontend (Vue)
- `web/src/router/index.ts` — change `:tsCode` to `:code`, add bars/predictions child routes
- `web/src/views/stock/StockDetailView.vue` — add back button, 3-tab layout
- `web/src/views/stock/PredictionTab.vue` — new or refactored component
- `web/src/views/stock/DailyBarsTab.vue` — new component (paginated daily bars)
- `web/src/views/stock/PredictionRecordsTab.vue` — new component (backtest table)
- `web/src/apis/*.ts` — update API calls to use short code
- `web/src/views/StockListView.vue` — add name column
- `web/src/stores/auth.ts` — fix auth initialization timing
- `web/src/App.vue` — await fetchMe before router mount
