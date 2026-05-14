# pkg/ Restructure — Design Spec

**Date:** 2026-05-14
**Status:** Approved (pending user review of written spec)
**Author:** travelliu (with Claude)
**Supersedes (partially):** §1.1 repo layout in `2026-05-14-go-vue-rewrite-design.md`

## 0. Goal

Flatten the package taxonomy under `pkg/` so each top-level directory has a single
clear responsibility, eliminate the duplicate definition between
`pkg/shared/spread.Bar` and `pkg/stockd/models.DailyBar`, and remove the
artificial `pkg/shared/` middle layer.

After this refactor:

- `pkg/tushare/` — Tushare API client and Tushare-wire data shapes only.
- `pkg/models/` — All project data models (GORM-mapped + JSON-mapped).
  Used by both the server binary (`stockd`) and the CLI binary (`stockctl` →
  renamed `cli`).
- `pkg/utils/` — Generic utility functions (stock-code formatting, spread math,
  distribution / range helpers).
- `pkg/analysis/` — Price-spread analysis pipeline (absorbs the analysis-specific
  windowing logic previously under `pkg/shared/window`).
- `pkg/cli/` — Renamed from `pkg/stockctl/` (CLI binary internals; `cmd/`
  subtree preserved as-is).
- `pkg/stockd/` — Server binary internals only; its `models/` subdirectory is
  removed (lifted to `pkg/models/`).
- `pkg/shared/` — **Deleted entirely.**

The refactor is **non-functional**: no DB schema change, no API contract change,
no CLI command change. Only file/import paths and one type unification.

## 1. Current State

```
pkg/
├── analysis/                    # Build / AnalysisResult / ModelTable / ReferenceTable
├── shared/                      # << to be dissolved
│   ├── spread/                  # OHLC, Spreads, Bar, Compute
│   ├── stockcode/               # ToTushareCode
│   └── window/                  # Names/Days/SpreadKeys/Window/MeansResult/Make/Means/
│                                # Composite/Distribution/RecommendedRange + private mean/roundTo
├── stockctl/                    # CLI binary internals (cmd/client/config/render)
├── stockd/
│   ├── models/                  # DailyBar/User/APIToken/JobRun/Portfolio/Stock/IntradayDraft
│   └── …                        # auth/bootstrap/config/db/http/middleware/services/utils
├── tushare/                     # client + Daily (returns []spread.Bar today)
└── version/
```

### 1.1 Duplicate definition (the main thing to fix)

Two structs describe the same concept:

```go
// pkg/shared/spread/spread.go
type Bar struct {
    TsCode, TradeDate                       string
    Open, High, Low, Close, Vol, Amount     float64
    Spreads                                 Spreads  // nested
}

// pkg/stockd/models/daily_bar.go
type DailyBar struct {
    TsCode, TradeDate                                       string  // GORM primary key
    Open, High, Low, Close, Vol, Amount                     float64
    SpreadOH, SpreadOL, SpreadHL, SpreadOC, SpreadHC, SpreadLC  float64  // flat
}
```

The first is the canonical "row" type used by `pkg/analysis.Build`,
`pkg/tushare.Daily()`, and the in-memory pipeline. The second is the
GORM-mapped table row. They have identical semantics but differ in field layout.
The Goroutine that bridges them is `pkg/stockd/services/analysis/analysis.go`,
which manually copies each field.

This is the duplicate to eliminate.

## 2. Target Structure

```
pkg/
├── analysis/
│   ├── builder.go               # Build + buildModelTable + buildReferenceTable
│   ├── model.go                 # AnalysisResult / ModelTable / ReferenceTable / Input
│   ├── format.go                # unchanged
│   ├── window.go        [NEW]   # ← shared/window analysis-specific bits
│   ├── builder_test.go
│   ├── format_test.go
│   ├── parity_test.go
│   ├── window_test.go   [NEW]   # ← shared/window/window_test.go (analysis-specific portion)
│   └── testdata/
│
├── cli/                 [RENAMED from pkg/stockctl]
│   ├── cmd/                     # (root, login, logout, version, stock, portfolio, draft, admin)
│   ├── client/                  # HTTP client wrapper + tests
│   ├── config/                  # viper-based config + tests
│   └── render/                  # tabwriter rendering for analysis result
│
├── models/              [NEW]
│   ├── api_token.go             # ← pkg/stockd/models/api_token.go
│   ├── daily_bar.go             # ← pkg/stockd/models/daily_bar.go (rewritten: embeds Spreads)
│   ├── intraday_draft.go        # ← pkg/stockd/models/intraday_draft.go
│   ├── job_run.go               # ← pkg/stockd/models/job_run.go
│   ├── portfolio.go             # ← pkg/stockd/models/portfolio.go
│   ├── spreads.go       [NEW]   # Spreads{OH,OL,HL,OC,HC,LC} (← shared/spread.Spreads)
│   ├── stock.go                 # ← pkg/stockd/models/stock.go
│   ├── user.go                  # ← pkg/stockd/models/user.go
│   └── models_test.go           # ← pkg/stockd/models/models_test.go (updated package path)
│
├── stockd/                      # server binary internals
│   ├── auth/ · bootstrap/ · config/ · db/ · http/ · middleware/ · services/ · utils/
│   # (models/ subdirectory removed)
│
├── tushare/
│   ├── client.go                # unchanged
│   ├── client_test.go           # unchanged
│   ├── stock_basic.go           # unchanged
│   ├── stock_basic_test.go      # unchanged
│   ├── daily.go                 # Daily() now returns []models.DailyBar
│   └── daily_test.go            # adjusted to new return type
│
├── utils/               [NEW]
│   ├── stockcode.go             # ToTushareCode (← shared/stockcode)
│   ├── stockcode_test.go
│   ├── spreads.go               # ComputeSpreads(open,high,low,close) models.Spreads
│   ├── spreads_test.go          # ← shared/spread/spread_test.go (signature rewritten)
│   ├── distribution.go          # Distribution + Bin (← shared/window)
│   ├── range.go                 # RecommendedRange + Range (← shared/window)
│   └── math.go                  # package-private mean / roundTo
│
└── version/                     # unchanged
```

`pkg/shared/` (including `.gitkeep`) is deleted.

## 3. File-Level Moves

### 3.1 Moves (body unchanged except for package clause / import paths)

| From | To |
|---|---|
| `pkg/shared/stockcode/stockcode.go` | `pkg/utils/stockcode.go` |
| `pkg/shared/stockcode/stockcode_test.go` | `pkg/utils/stockcode_test.go` |
| `pkg/stockd/models/api_token.go` | `pkg/models/api_token.go` |
| `pkg/stockd/models/intraday_draft.go` | `pkg/models/intraday_draft.go` |
| `pkg/stockd/models/job_run.go` | `pkg/models/job_run.go` |
| `pkg/stockd/models/portfolio.go` | `pkg/models/portfolio.go` |
| `pkg/stockd/models/stock.go` | `pkg/models/stock.go` |
| `pkg/stockd/models/user.go` | `pkg/models/user.go` |
| `pkg/stockd/models/models_test.go` | `pkg/models/models_test.go` |
| `pkg/stockctl/**` (everything) | `pkg/cli/**` |

Package declaration in each moved file:

- `pkg/utils/*.go` (non-test) → `package utils` (was `package stockcode`).
- `pkg/utils/*_test.go` (external-test files) → `package utils_test` (was
  `package stockcode_test`); import `stock/pkg/shared/stockcode` → `stock/pkg/utils`;
  call sites `stockcode.ToTushareCode` → `utils.ToTushareCode`.
- `pkg/models/*` → stays `package models` (current `pkg/stockd/models` is already
  `package models`). Only the imports in `models_test.go` need updating
  (`stock/pkg/stockd/models` is replaced by relying on the same-package
  rename; `stock/pkg/stockd/db` import is unchanged).
- `pkg/cli/cmd/*` → stays `package cmd` (the rename only affects the import path,
  not the inner package names). Sibling imports
  `stock/pkg/stockctl/{client,render,config}` get rewritten to
  `stock/pkg/cli/{client,render,config}`.

In short: every moved `.go` file may need *some* small edit (import path or
external-package qualifier), but no logic changes happen in this category.

### 3.2 Split & rewrite

**`pkg/shared/spread/spread.go` is deconstructed:**

| Old element | New location | Notes |
|---|---|---|
| `type Spreads struct { OH, OL, HL, OC, HC, LC float64 }` | `pkg/models/spreads.go` | Unchanged shape |
| `type OHLC struct { … }` | **Deleted** | Replaced by flat parameters on `ComputeSpreads` |
| `func Compute(b OHLC) Spreads` | `pkg/utils/spreads.go` as `func ComputeSpreads(open, high, low, close float64) models.Spreads` | Flat signature |
| `type Bar struct { … Spreads }` | **Deleted** | Replaced by unified `models.DailyBar` |

**`pkg/shared/window/window.go` is split:**

| Old element | New location | Notes |
|---|---|---|
| `var Names`, `var Days`, `var SpreadKeys` | `pkg/analysis/window.go` | Analysis-specific constants |
| `type Window struct { Name string; Rows []spread.Bar }` | `pkg/analysis/window.go` as `type Window struct { Name string; Rows []models.DailyBar }` | Row type updated |
| `type MeansResult map[string]map[string]*float64` | `pkg/analysis/window.go` | Unchanged |
| `func Make`, `func Means`, `func Composite`, `func extract` | `pkg/analysis/window.go` | `Rows []models.DailyBar` instead of `spread.Bar`; `extract` reads `r.Spreads.*` (works identically since embedded `Spreads` exposes the same field path) |
| `type Bin struct { … }`, `func Distribution` | `pkg/utils/distribution.go` | Generic histogram (no domain types) |
| `type Range struct { … }`, `func RecommendedRange` | `pkg/utils/range.go` | Generic tight-range search (no domain types) |
| `func mean`, `func roundTo` | `pkg/utils/math.go` (package-private) | Stay lowercase |
| `func ptr(v int) *int` | `pkg/analysis/window.go` (local to window.go) | Only used by `Days` initializer |

**`pkg/shared/window/window_test.go` is split** along the same axis: tests of
`Make / Means / Composite` go with `pkg/analysis/window_test.go`; tests of
`Distribution / RecommendedRange` go with the respective `pkg/utils/*_test.go`.

### 3.3 DailyBar unification

`pkg/models/daily_bar.go` becomes:

```go
package models

type DailyBar struct {
    TsCode    string `gorm:"primaryKey;size:16"`
    TradeDate string `gorm:"primaryKey;size:8"`
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Vol       float64
    Amount    float64
    Spreads   Spreads `gorm:"embedded;embeddedPrefix:spread_"`
}
```

`pkg/models/spreads.go`:

```go
package models

type Spreads struct {
    OH float64
    OL float64
    HL float64
    OC float64
    HC float64
    LC float64
}
```

Notes:

- **No `json` tags are added** to the new types — the current
  `pkg/stockd/models.DailyBar` carries no `json` tags either, so adding them
  here would be a behavioral change masked as a refactor.
- With `embedded;embeddedPrefix:spread_`, GORM generates the same column names
  (`spread_oh`, `spread_ol`, `spread_hl`, `spread_oc`, `spread_hc`, `spread_lc`)
  as the current schema. No DB migration is required.
- **JSON shape**: if any consumer marshals `models.DailyBar` directly (today,
  none do — see §3.3.1 below), the wire form changes from flat
  `SpreadOH/SpreadOL/...` fields to a nested `Spreads:{OH:..,OL:..,...}` object.

### 3.3.1 Check: no current handler exposes `models.DailyBar` JSON

The `/api/bars/:ts_code` handler returns a hand-rolled projection
(`TradeDate/Open/High/Low/Close`), not the full GORM struct. The analysis
endpoints return `pkganalysis.AnalysisResult` (formatted strings). Before
merging, re-grep the handler layer for `c.JSON(*, models.DailyBar` or
`c.JSON(*, dailyBar` to confirm. If any direct marshaling exists, either:

1. Add explicit `json` tags that flatten `Spreads.*` to top-level keys
   (`json:"spread_oh"` etc. inline-tagged via `inline:""` or a custom
   `MarshalJSON`), or
2. Rewire the handler to a projection DTO.

The non-functional guarantee of this refactor is contingent on this check
passing.

## 4. Callsite Updates

### 4.1 Import-path-only changes

| File | Change |
|---|---|
| `cmd/stockctl/main.go` | `stock/pkg/stockctl/cmd` → `stock/pkg/cli/cmd` |
| 20 files importing `stock/pkg/stockd/models` (see list below) | → `stock/pkg/models` |
| Files inside `pkg/stockctl/**` referencing sibling `stock/pkg/stockctl/*` | → `stock/pkg/cli/*` |

The 20 files importing `pkg/stockd/models`:

- `pkg/stockd/auth/auth_test.go`
- `pkg/stockd/auth/middleware.go`
- `pkg/stockd/bootstrap/bootstrap.go`
- `pkg/stockd/bootstrap/bootstrap_test.go`
- `pkg/stockd/db/db.go`
- `pkg/stockd/http/handler/auth_test.go`
- `pkg/stockd/services/analysis/analysis.go`
- `pkg/stockd/services/analysis/analysis_test.go`
- `pkg/stockd/services/bars/bars.go`
- `pkg/stockd/services/bars/bars_test.go`
- `pkg/stockd/services/draft/draft.go`
- `pkg/stockd/services/portfolio/portfolio.go`
- `pkg/stockd/services/scheduler/scheduler.go`
- `pkg/stockd/services/scheduler/scheduler_test.go`
- `pkg/stockd/services/stock/csv.go`
- `pkg/stockd/services/stock/stock.go`
- `pkg/stockd/services/stock/stock_test.go`
- `pkg/stockd/services/token/token.go`
- `pkg/stockd/services/user/user.go`
- (plus `pkg/stockd/models/models_test.go` which is itself moved)

### 4.2 Substantive updates

**`pkg/tushare/daily.go`** (and `daily_test.go`):

```go
// before
func Daily(ctx context.Context, c *Client, token string, req DailyRequest) ([]spread.Bar, error)

// after
func Daily(ctx context.Context, c *Client, token string, req DailyRequest) ([]models.DailyBar, error)
```

The inner construction:

```go
// before
bar := spread.Bar{TsCode: ..., …}
bar.Spreads = spread.Compute(spread.OHLC{Open: open, High: high, Low: low, Close: close})

// after
bar := models.DailyBar{TsCode: ..., …}
bar.Spreads = utils.ComputeSpreads(open, high, low, close)
```

**`pkg/analysis/builder.go` + `model.go` + `*_test.go`**:

```go
// before
import (
    "stock/pkg/shared/spread"
    "stock/pkg/shared/window"
)
type Input struct { ... Rows []spread.Bar ... }
WindowMeans window.MeansResult

// after
import "stock/pkg/models"
type Input struct { ... Rows []models.DailyBar ... }
WindowMeans MeansResult     // same package now
```

Drop the `window.` prefix on calls (`window.Make` → `Make`, `window.Names` →
`Names`, etc.) since `window.go` is now a file in `package analysis`.

**`pkg/stockd/services/analysis/analysis.go`**:

```go
// before — manual Bar construction
rows := make([]spread.Bar, 0, len(bars))
for _, b := range bars {
    rows = append(rows, spread.Bar{
        TsCode: b.TsCode, ..., 
        Spreads: spread.Spreads{OH: b.SpreadOH, OL: b.SpreadOL, ...},
    })
}
res := pkganalysis.Build(pkganalysis.Input{ ..., Rows: rows, ... })

// after — direct pass-through
res := pkganalysis.Build(pkganalysis.Input{ ..., Rows: bars, ... })
```

The whole `rows` construction loop disappears (this is the concrete payoff of
unifying `Bar` and `DailyBar`).

**`pkg/stockd/services/analysis/analysis_test.go`**:

Change existing inline-construction `models.DailyBar{SpreadOH: 2, SpreadOL: 2, …}`
to `models.DailyBar{Spreads: models.Spreads{OH: 2, OL: 2, …}}`.

## 5. Migration Sequence

**Single atomic commit.** Order within the commit (for self-review readability):

1. Create `pkg/models/` by moving `pkg/stockd/models/*.go` and editing
   `daily_bar.go` to embed `Spreads` (also add `pkg/models/spreads.go`).
2. Create `pkg/utils/` by moving `pkg/shared/stockcode/*` and writing
   `spreads.go` / `distribution.go` / `range.go` / `math.go`.
3. Create `pkg/analysis/window.go` from the analysis-specific subset of
   `pkg/shared/window/window.go` (with `[]spread.Bar` → `[]models.DailyBar`).
4. Update `pkg/tushare/daily.go` to return `[]models.DailyBar` and call
   `utils.ComputeSpreads`.
5. Update `pkg/analysis/builder.go` + `model.go` + `*_test.go` to use
   `models.DailyBar` and drop the `window.` prefix.
6. Update `pkg/stockd/services/analysis/analysis.go` to remove manual `Bar`
   construction.
7. Rewrite imports `stock/pkg/stockd/models` → `stock/pkg/models` across 20
   files (sed-able).
8. `git mv pkg/stockctl pkg/cli` and rewrite imports `stock/pkg/stockctl/...` →
   `stock/pkg/cli/...` across 11 files (including `cmd/stockctl/main.go`).
9. `git rm -r pkg/shared/`.
10. `go build ./...` + `go test -race ./...` + `go vet ./...` + (if installed)
    `gosec ./...` must all pass.

## 6. Verification

Mandatory before declaring done:

- `go build ./...` — clean.
- `go vet ./...` — clean.
- `go test -race ./...` — all green, with particular attention to:
  - `pkg/analysis/parity_test.go` (Python parity — the canary for unchanged
    analysis semantics).
  - `pkg/models/models_test.go` (a new test, see below).
  - `pkg/stockd/services/analysis/analysis_test.go` (verifies the now-direct
    `models.DailyBar → analysis.Input` flow).
- `pkg/models/models_test.go` already contains `TestModelsRoundTrip` which
  writes a `models.DailyBar` (the old flat-field version). After the refactor,
  augment this test with a dedicated `TestDailyBarHasSpreadColumns` that
  `AutoMigrate`s a `DailyBar` into a fresh in-memory SQLite DB and asserts
  `Migrator().HasColumn(&DailyBar{}, "spread_oh")` (and the other five) are
  true. This is the structural test that `embeddedPrefix:spread_` keeps the
  column names identical to the pre-refactor schema.
- Optional manual check: `sqlite3 <db> ".schema daily_bars"` before and after,
  diff = empty.

## 7. Risks

- **GORM column-name regression.** If `embeddedPrefix:spread_` is mis-spelled
  (e.g., missing trailing underscore) or omitted, GORM will rename columns to
  `o_h, o_l, …` and break the existing table. Mitigation: the GORM round-trip
  test in §6.
- **Parity drift.** `pkg/analysis/parity_test.go` is the contract. Any signature
  change that subtly reorders operations (e.g., the now-direct
  `models.DailyBar` slice passed in pre-sorted/unsorted order) could shift the
  pipeline output. Mitigation: parity test must remain green.
- **Hidden JSON consumers of `models.DailyBar`.** If any HTTP handler currently
  marshals `models.DailyBar` directly, the wire shape will change from flat
  `spread_oh` fields to a nested `spreads` object. Mitigation: grep the HTTP
  layer before merging; the spec already records §3.3 that no production
  handler does this today, but recheck.
- **`pkg/cli` import-path churn.** A handful of CLI cobra files cross-import
  each other (`pkg/stockctl/client`, `pkg/stockctl/render`); all must be
  rewritten consistently. Mitigation: rely on `go build` to surface any miss.
- **Test file fragmentation.** `pkg/shared/window/window_test.go` is split into
  three files in three packages. Take care to preserve every assertion (use
  a diff to verify no cases drop).

## 8. Non-Goals

This refactor explicitly does **not**:

- Modify DB schema (column names preserved via `embeddedPrefix`).
- Modify API JSON contracts (HTTP handler projections unchanged).
- Modify CLI command surface or `cmd/` internal structure.
- Replace CLI inline anonymous response structs with `models.Stock /
  Portfolio / IntradayDraft` (a separate cleanup, deferred).
- Add `json:"-"` to `User.PasswordHash` / `APIToken.TokenHash` (a separate
  security pass, deferred — these types are not currently marshaled by any
  handler).
- Deduplicate `pkg/cli/render.AnalysisResult` against `pkg/analysis.AnalysisResult`
  (a separate cleanup, deferred).
- Touch `pkg/version/` or `pkg/stockd/utils/` (envelope, errors).

---

**Implementation handoff:** Once approved, this spec feeds into the
`writing-plans` skill to produce a step-by-step implementation plan.
