# External Data Integration Design

**Date:** 2026-05-16  
**Scope:** Integrate signal, valuation, research, and news data from external Chinese A-share APIs into the stockd backend.

---

## Background

The current project pulls OHLCV from Tushare Pro and real-time quotes from Tencent Finance. The goal is to add four additional data layers sourced from public HTTP APIs (no Python/akshare runtime dependency), implemented as pure Go HTTP clients following the existing `pkg/tencent/` pattern.

---

## Architecture

### New Packages

```
pkg/
├── ths/        # Tonghuashun (hot stocks, EPS forecast, industry comparison)
├── eastmoney/  # East Money (research reports, news, dragon-tiger board, lockup calendar)
├── baidu/      # Baidu Stock (concept blocks, fund flow)
└── hexin/      # hexin.cn (northbound fund minute flow)
```

Each package exposes a `Client` struct with an `Option`-based constructor (`NewClient`, `WithBaseURL`) matching the `pkg/tencent/` convention.

### Service Integration

New clients are added as fields to `services.Service` alongside the existing `tc *tencent.Client`:

```go
type Service struct {
    // existing fields ...
    tc   *tencent.Client
    ths  *ths.Client
    em   *eastmoney.Client
    bdu  *baidu.Client
    hxn  *hexin.Client
}
```

### HTTP Layer

New handler files in `pkg/stockd/http/`:

- `signals.go` — global signal endpoints
- `external.go` — per-stock external data endpoints (concepts, fund flow, dragon-tiger, lockup, EPS forecast, valuation, reports, news, announcements)

All new endpoints require `AuthRequired()` middleware.

### Caching

Signal endpoints (hot stocks, northbound, industry) use a 5-minute in-memory TTL cache, following the existing `realtimeCache` pattern in `Service`.

### No New Go Dependencies

All clients use standard library `net/http`. No new entries in `go.mod`.

---

## Batch 1: Signal Layer (Zero-auth direct APIs)

**New endpoints:**

| Method | Path | Source | Description |
|--------|------|--------|-------------|
| GET | `/api/signals/hot` | `zx.10jqka.com.cn` | Daily strong stocks + theme reason tags |
| GET | `/api/signals/northbound` | `data.hexin.cn` | Northbound fund minute flow (HGT/SGT) |
| GET | `/api/signals/industry` | Tonghuashun industry API | ~90 industries ranked by change%, net inflow, leader stock |
| GET | `/api/stocks/:code/concepts` | `finance.pae.baidu.com` | Industry/concept/region block membership |
| GET | `/api/stocks/:code/fund-flow` | `finance.pae.baidu.com` | Per-stock fund flow (`?type=realtime&date=YYYYMMDD` or `?type=history`) |

**Packages created:** `pkg/ths/`, `pkg/hexin/`, `pkg/baidu/`

---

## Batch 2: Dragon-Tiger Board + Lockup Calendar

**New endpoints:**

| Method | Path | Source | Description |
|--------|------|--------|-------------|
| GET | `/api/signals/dragon-tiger` | East Money datacenter API | Market-wide dragon-tiger board for a date (`?date=YYYY-MM-DD`) |
| GET | `/api/stocks/:code/dragon-tiger` | East Money datacenter API | Per-stock board history + top-5 buy/sell seats |
| GET | `/api/stocks/:code/lockup` | East Money datacenter API | Lockup expiry history + upcoming 90-day calendar |

**Package created:** `pkg/eastmoney/` (initial dragon-tiger + lockup endpoints)

---

## Batch 3: Valuation Layer

**New endpoints:**

| Method | Path | Source | Description |
|--------|------|--------|-------------|
| GET | `/api/stocks/:code/eps-forecast` | Tonghuashun THS API | Institutional consensus EPS by year (count, min, mean, max) |
| GET | `/api/stocks/:code/valuation` | Tencent quote (existing) + EPS forecast | Forward PE, PEG, PE-digestion years |

**Valuation formulas:**
- Forward PE = current price / consensus EPS (current year)
- CAGR = next year EPS / current year EPS − 1
- PEG = forward PE / (CAGR × 100); <1 cheap, 1–1.5 fair, >1.5 expensive
- PE digestion years = log(forward PE / 30) / log(1 + CAGR); anchor at 30× for A-share growth stocks

**Package extended:** `pkg/ths/` (add EPS forecast method)

---

## Batch 4: Research Reports + News + Announcements

**New endpoints:**

| Method | Path | Source | Description |
|--------|------|--------|-------------|
| GET | `/api/stocks/:code/reports` | `reportapi.eastmoney.com` | Research report list (title, org, date, rating, EPS forecasts) |
| GET | `/api/stocks/:code/reports/:id/pdf` | East Money PDF CDN | Server-side PDF proxy (avoids browser CORS) |
| GET | `/api/stocks/:code/news` | East Money news API | Per-stock news list |
| GET | `/api/signals/cls-news` | Cailian Press API | Real-time market flash news |
| GET | `/api/stocks/:code/announcements` | cninfo API | Exchange announcements list |

**Note:** Cailian Press and cninfo underlying API URLs are to be reverse-engineered from akshare source during implementation.

**Package extended:** `pkg/eastmoney/` (add reports, PDF proxy, news, announcements)

---

## Data Flow

```
Frontend → stockd HTTP handler → Service method → pkg/* client → External API
```

No data is persisted to the database. All responses are proxied in real time (with TTL cache where noted).

---

## Error Handling

- Return `502 Bad Gateway` with a descriptive message when the upstream API fails.
- Return `404` when the upstream returns empty data for a valid stock code.
- Never surface raw upstream error bodies to the client (log them server-side).

---

## Auth

All new endpoints require `AuthRequired()`. Signal endpoints (`/api/signals/*`) are also auth-gated since they are part of the authenticated app experience.
