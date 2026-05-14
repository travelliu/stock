# Go + Vue Rewrite — Design Spec

**Date:** 2026-05-14
**Status:** Approved (sections §1–§7)
**Author:** travelliu (with Claude)
**Reference architecture:** `/root/code/gitlab/mogdb_en/mtk/` (mtkd)

## 0. Goal

Rewrite the existing Python A-share spread-analysis CLI (`stock.py` / `analysis.py` /
`fetcher.py`) as a Go + Vue 3 application with:

- Multi-user web UI (Element Plus) with auth + HTTPS;
- Portfolio management (add / delete symbols, view per-stock analysis & prediction);
- Today's intraday draft (open / high / low / close) input, kept separate from the
  official daily history;
- A remote-only CLI (`stockctl`) that talks to the server, plus a Claude skill
  (`stock-analyst`) that wraps the CLI.

Behaviour of the analysis (6 spreads × 4 windows, the trading-plan table layout,
`to_tushare_code()` market suffix rules, etc.) must match the current Python
implementation bit-for-bit.

---

## 1. Architecture (Plan A — single repo, dual binaries, embedded frontend)

### 1.1 Repo layout

```
stock/
├── cmd/
│   ├── stockd/             # server binary (HTTP API + scheduler)
│   └── stockctl/           # remote CLI binary
├── pkg/
│   ├── shared/             # cross-cutting helpers
│   │   ├── stockcode/      # to_tushare_code + market detection
│   │   ├── spread/         # 6-spread computation
│   │   └── window/         # 历史/90/30/15 window slicing
│   ├── tushare/            # standalone Tushare HTTP SDK (token-injectable)
│   ├── analysis/           # ports analysis.py (means / model / reference table)
│   ├── stockd/             # server internals
│   │   ├── config/         # viper config loader + env var mapping
│   │   ├── db/             # GORM init + AutoMigrate (sqlite/mysql/postgres)
│   │   ├── middleware/     # requestID, panic recovery, auth
│   │   ├── models/         # GORM models
│   │   ├── services/       # business logic (portfolio, draft, analysis, sync…)
│   │   ├── scheduler/      # robfig/cron jobs
│   │   ├── http/           # gin handlers + router + swagger
│   │   └── utils/          # response envelope, error codes, i18n messages
│   └── stockctl/           # CLI command tree (cobra)
├── web/                    # Vue 3 + TS + Vite + Element Plus source
│   └── dist/               # build output, consumed by embed.go
├── embed.go                # package stock; //go:embed all:web/dist
├── docs/
├── go.mod / go.sum
└── README.md
```

### 1.2 Why this shape

- **Dual binaries** keep server / CLI deps independent (the CLI doesn't pull GORM
  drivers); both share `pkg/shared`, `pkg/tushare`, `pkg/analysis`.
- **`pkg/tushare`** is a thin SDK — anyone can import it independently of the
  server. Token is passed per-call so each user's override token works.
- **`pkg/stockd/services`** is a single package (not split per domain). A `StockD`
  struct aggregates `repo *db.Repo`, `logger`, `conf` and exposes all business
  methods. This mirrors the `mtkd` `services.MtkD` pattern and avoids premature
  package fragmentation.
- **`pkg/stockd/utils`** holds the HTTP response envelope, error-code constants,
  and i18n message maps (`zh` / `en`). This is infrastructure shared by HTTP and
  scheduler layers.
- **`pkg/stockd/middleware`** hosts `requestID`, `panicRecovery`, and `auth`
  gin handlers, mounted in a fixed order.
- **`web/` at repo root** matches the user's preference; `embed.go` lives at the
  repo root as `package stock` so it can `//go:embed all:web/dist` without `..`
  path issues. The server imports the root package to obtain the embedded FS.
- **AutoMigrate on startup** — no separate migrate command. `stockd` exposes
  exactly two subcommands: default (serve) and `version`. All admin / user /
  token management is done through the Web UI.

### 1.3 embed.go (root)

```go
package stock

import (
    "embed"
    "io/fs"
    "net/http"

    "github.com/gin-contrib/static"
)

//go:embed all:web/dist
var StaticDir embed.FS

type embedFS struct{ http.FileSystem }

func (e embedFS) Exists(prefix, filepath string) bool {
    if _, err := e.Open(filepath); err != nil {
        return false
    }
    return true
}

func EmbedFolder() static.ServeFileSystem {
    sub, _ := fs.Sub(StaticDir, "web/dist")
    return embedFS{http.FS(sub)}
}
```

`pkg/stockd/http/router.go` calls `stock.EmbedFolder()` and mounts it via
`static.Serve("/", ...)` with SPA fallback to `index.html`.

---

## 2. Data Model (GORM, AutoMigrate)

All tables created at startup via `db.AutoMigrate(...)`. SQLite default; MySQL /
Postgres switchable via config.

```go
// pkg/stockd/models/user.go
type User struct {
    // 用户ID
    ID           uint      `json:"id" gorm:"primaryKey;comment:用户ID"`
    // 用户名
    Username     string    `json:"username" gorm:"uniqueIndex;size:64;not null;comment:用户名"`
    // 密码哈希 (bcrypt)
    PasswordHash string    `json:"-" gorm:"not null;comment:密码哈希"`
    // 角色 user | admin
    Role         string    `json:"role" gorm:"size:16;not null;comment:角色"`
    // Tushare Token 覆盖
    TushareToken string    `json:"tushareToken,omitempty" gorm:"size:128;comment:Tushare Token 覆盖"`
    // 是否禁用
    Disabled     bool      `json:"disabled" gorm:"not null;default:false;comment:是否禁用"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}

// pkg/stockd/models/api_token.go
type APIToken struct {
    // ID
    ID         uint   `json:"id" gorm:"primaryKey;comment:ID"`
    // 用户ID
    UserID     uint   `json:"userID" gorm:"index;not null;comment:用户ID"`
    // Token 名称
    Name       string `json:"name" gorm:"size:64;not null;comment:Token 名称"`
    // Token 哈希 (sha256 of plain stk_xxx)
    TokenHash  string `json:"-" gorm:"uniqueIndex;size:64;not null;comment:Token 哈希"`
    // 最后使用时间
    LastUsedAt *time.Time `json:"lastUsedAt,omitempty" gorm:"comment:最后使用时间"`
    // 过期时间
    ExpiresAt  *time.Time `json:"expiresAt,omitempty" gorm:"comment:过期时间"`
    CreatedAt  time.Time  `json:"createdAt"`
}

// pkg/stockd/models/stock.go — shared catalog, not user-scoped
type Stock struct {
    // TS代码 e.g. 600519.SH
    TsCode    string `json:"tsCode" gorm:"primaryKey;size:16;comment:TS代码"`
    // 股票代码
    Code      string `json:"code" gorm:"index;size:8;not null;comment:股票代码"`
    // 股票名称
    Name      string `json:"name" gorm:"size:32;not null;comment:股票名称"`
    // 地域
    Area      string `json:"area" gorm:"size:16;comment:地域"`
    // 行业
    Industry  string `json:"industry" gorm:"size:32;comment:行业"`
    // 市场 主板/创业板/...
    Market    string `json:"market" gorm:"size:16;comment:市场"`
    // 交易所 SSE/SZSE
    Exchange  string `json:"exchange" gorm:"size:8;comment:交易所"`
    // 上市日期 YYYYMMDD
    ListDate  string `json:"listDate" gorm:"size:8;comment:上市日期"`
    // 是否退市
    Delisted  bool   `json:"delisted" gorm:"not null;default:false;comment:是否退市"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// pkg/stockd/models/daily_bar.go — shared official history
type DailyBar struct {
    // TS代码
    TsCode    string  `json:"tsCode" gorm:"primaryKey;size:16;comment:TS代码"`
    // 交易日期 YYYYMMDD
    TradeDate string  `json:"tradeDate" gorm:"primaryKey;size:8;comment:交易日期"`
    Open      float64 `json:"open" gorm:"comment:开盘价"`
    High      float64 `json:"high" gorm:"comment:最高价"`
    Low       float64 `json:"low" gorm:"comment:最低价"`
    Close     float64 `json:"close" gorm:"comment:收盘价"`
    Vol       float64 `json:"vol" gorm:"comment:成交量"`
    Amount    float64 `json:"amount" gorm:"comment:成交额"`
    SpreadOH  float64 `json:"spreadOH" gorm:"comment:最高-开盘价差"`
    SpreadOL  float64 `json:"spreadOL" gorm:"comment:开盘-最低价差"`
    SpreadHL  float64 `json:"spreadHL" gorm:"comment:最高-最低价差"`
    SpreadOC  float64 `json:"spreadOC" gorm:"comment:开盘-收盘价差"`
    SpreadHC  float64 `json:"spreadHC" gorm:"comment:最高-收盘价差"`
    SpreadLC  float64 `json:"spreadLC" gorm:"comment:最低-收盘价差"`
}

// pkg/stockd/models/portfolio.go — per user, symbols only (v1)
type Portfolio struct {
    // ID
    ID      uint      `json:"id" gorm:"primaryKey;comment:ID"`
    // 用户ID
    UserID  uint      `json:"userID" gorm:"uniqueIndex:idx_user_code;not null;comment:用户ID"`
    // TS代码
    TsCode  string    `json:"tsCode" gorm:"uniqueIndex:idx_user_code;size:16;not null;comment:TS代码"`
    // 备注
    Note    string    `json:"note" gorm:"size:255;comment:备注"`
    AddedAt time.Time `json:"addedAt"`
}

// pkg/stockd/models/intraday_draft.go — today's user-entered OHLC
type IntradayDraft struct {
    // ID
    ID        uint      `json:"id" gorm:"primaryKey;comment:ID"`
    // 用户ID
    UserID    uint      `json:"userID" gorm:"uniqueIndex:idx_user_code_date;not null;comment:用户ID"`
    // TS代码
    TsCode    string    `json:"tsCode" gorm:"uniqueIndex:idx_user_code_date;size:16;not null;comment:TS代码"`
    // 交易日期 YYYYMMDD
    TradeDate string    `json:"tradeDate" gorm:"uniqueIndex:idx_user_code_date;size:8;not null;comment:交易日期"`
    Open      *float64  `json:"open,omitempty" gorm:"comment:开盘价"`
    High      *float64  `json:"high,omitempty" gorm:"comment:最高价"`
    Low       *float64  `json:"low,omitempty" gorm:"comment:最低价"`
    Close     *float64  `json:"close,omitempty" gorm:"comment:收盘价"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// pkg/stockd/models/job_run.go — scheduler execution record
type JobRun struct {
    // ID
    ID        uint      `json:"id" gorm:"primaryKey;comment:ID"`
    // 任务名称
    JobName   string    `json:"jobName" gorm:"size:64;not null;comment:任务名称"`
    // 开始时间
    StartedAt time.Time `json:"startedAt" gorm:"comment:开始时间"`
    // 结束时间
    FinishedAt *time.Time `json:"finishedAt,omitempty" gorm:"comment:结束时间"`
    // 状态 running / finish / failed
    Status    string    `json:"status" gorm:"size:16;not null;comment:状态"`
    // 消息
    Message   string    `json:"message" gorm:"size:512;comment:消息"`
}
```

Notes:

- `daily_bars` is global / shared so all users share one ingest pipeline.
- `intraday_drafts` keeps `trade_date` in the unique key so a user can keep
  several days of drafts if needed (UI shows today by default).
- Foreign keys are intentionally omitted (sqlite + cross-driver simplicity).
  Service layer enforces ownership.

---

## 3. HTTP API + Auth

### 3.1 Dual auth

| Caller         | Mechanism                                       | Header                          |
|----------------|-------------------------------------------------|---------------------------------|
| Browser (Vue)  | Cookie session (`gin-contrib/sessions`, cookie store, 7 d) | `Cookie: session=…`             |
| CLI / Skill    | Bearer API token (`stk_xxx…`, sha256 in DB)     | `Authorization: Bearer stk_…`   |

### 3.2 Middleware stack (fixed order)

All middleware lives in `pkg/stockd/middleware/` and is mounted in this order:

1. **RequestID** — injects `X-Request-ID` (or extracts from incoming header) and
   stores it in `gin.Context` and `context.Context`.
2. **PanicRecovery** — `defer recover()`; logs stack trace via `logger.Errorf`,
   returns `code: 500` envelope, aborts request.
3. **CORS** — `gin-contrib/cors`, configured for the SPA origin.
4. **Auth** — tries `Authorization: Bearer` first, falls back to session cookie.
   On success attaches `*models.User` and effective Tushare token to context.
   On failure returns `401` (not aborting, so public routes can skip auth).
5. **Logger** — records method, path, status, latency, requestID.

### 3.3 Routes (all under `/api`)

```
POST   /api/auth/login           {username,password} → set cookie
POST   /api/auth/logout
GET    /api/auth/me

# admin-only
POST   /api/admin/users          create user (admin issues accounts)
GET    /api/admin/users
PATCH  /api/admin/users/:id      reset password / role / disable / tushare override
DELETE /api/admin/users/:id

# user self-service
GET    /api/me/tokens
POST   /api/me/tokens            {name,expires_at?} → returns plain stk_xxx once
DELETE /api/me/tokens/:id
PATCH  /api/me/tushare_token     {token}            (per-user override)
POST   /api/me/password          {old,new}

# stocks (shared catalog)
GET    /api/stocks?q=...&limit=  search by code / name
GET    /api/stocks/:tsCode
POST   /api/admin/stocks/sync                       trigger stocklist sync
POST   /api/admin/stocks/import-csv                 multipart CSV upload

# portfolio (user-scoped)
GET    /api/portfolio
POST   /api/portfolio            {ts_code,note?}
DELETE /api/portfolio/:tsCode
PATCH  /api/portfolio/:tsCode    {note}

# daily bars + drafts
GET    /api/bars/:tsCode?from=&to=          official history
POST   /api/admin/bars/sync                  trigger daily-fetch for tracked stocks
GET    /api/drafts/today?ts_code=
PUT    /api/drafts                {ts_code, trade_date, open?, high?, low?, close?}
DELETE /api/drafts/:id

# analysis (read-only, derives spread tables + reference table + trading plan)
GET    /api/analysis/:tsCode
       ?actual_open=&actual_high=&actual_low=&actual_close=
       ?yesterday_close=         (auto-resolved from latest daily_bar if absent)
       ?with_draft=true|false    default true; when true, any actual_* not
                                 explicitly passed is filled from today's
                                 intraday_draft (if a draft row exists)
```

### 3.4 Response envelope

All JSON responses follow the same envelope (mirrors `mtkd` `utils.HTTPResponse`):

```json
{ "requestID": "uuid", "code": 200, "message": "Succeed", "data": {...} }
{ "requestID": "uuid", "code": 40002, "message": "未登录或登录已过期", "data": null }
```

- `requestID` — generated at request entry, carried through logs.
- `code` — business code; `200` means success. HTTP status is always `200 OK`.
- `message` — looked up from i18n maps (`zh` / `en`) by `code` + `Accept-Language`.
- `data` — payload on success; `null` on error.

Paginated lists wrap pagination inside `data`:
```json
{
  "requestID": "uuid",
  "code": 200,
  "message": "Succeed",
  "data": {
    "list": [...],
    "meta": { "total": 100, "page": 1, "limit": 20 }
  }
}
```

### 3.5 Error code system

Business errors are defined in `pkg/stockd/utils/errors.go` and `pkg/stockd/utils/msg.go`, following the `mtkd` pattern.

`pkg/stockd/utils/errors.go`:
```go
const (
    SUCCESS            = 200
    ERROR              = 500
    ErrInvalidParam    = 40001
    ErrUnauthorized    = 40002
    ErrForbidden       = 40003
    ErrUserNotFound    = 40004
    ErrInvalidPassword = 40005
    ErrUserDisabled    = 40006
    ErrStockNotFound   = 40007
    ErrDraftInvalid    = 40008
    ErrTokenInvalid    = 40009
    ErrTokenExpired    = 40010
    ErrDuplicateUser   = 40011
    ErrInvalidCode     = 40012
    ErrTaskRun         = 40013
    ErrTaskNoRunReport = 40014
)
```

`pkg/stockd/utils/msg.go`:
```go
var defaultZhMsg = map[int]string{
    SUCCESS:            "Succeed",
    ERROR:              "系统异常，请联系管理员",
    ErrInvalidParam:    "参数校验失败: %s",
    ErrUnauthorized:    "未登录或登录已过期",
    ErrForbidden:       "权限不足",
    ErrUserNotFound:    "用户{%v}不存在",
    ErrInvalidPassword: "密码错误",
    ErrUserDisabled:    "账号已被禁用",
    ErrStockNotFound:   "股票{%v}不存在",
    ErrDraftInvalid:    "草稿不存在或已失效",
    ErrTokenInvalid:    "无效的 Token",
    ErrTokenExpired:    "Token 已过期",
    ErrDuplicateUser:   "用户名{%v}已存在",
    ErrInvalidCode:     "股票代码格式错误",
    ErrTaskRun:         "任务正在运行中",
    ErrTaskNoRunReport: "任务不存在",
}
```

Error construction:
```go
return utils.New(utils.ErrUserNotFound, username)
// or
return utils.Wrap(utils.ErrUserNotFound, err, "get user failed")
```

### 3.6 Analysis response shape

Mirrors the Python `_build_spread_model_table` + `_build_reference_table` output
so the CLI can render the existing trading-plan layout unchanged:

```jsonc
{
  "ts_code": "600519.SH",
  "stock_name": "贵州茅台",
  "yesterday_close": 1620.00,
  "windows": ["历史", "近3月", "近1月", "近2周"],
  "model_table": {
    "headers": ["时段", "开盘与最高价", "开盘与最低价", "最高与最低价", "开盘与收盘价", "最高与收盘价", "最低与收盘价"],
    "rows": [
      ["历史",  "1.23", "0.89", "2.10", "0.75", "1.05", "0.62"],
      // ... one row per window (4 windows + 1 composite)
    ]
  },
  "reference_table": {
    "headers": [...],
    "rows":    [...]    // formatted cells, each row is []string
  }
}
```

### 3.7 Swagger

`swaggo/gin-swagger` mounted at `/swagger/index.html` for developer reference.

---

## 4. Frontend (Vue 3 + TS + Element Plus)

### 4.1 Stack

- Vite + Vue 3 + TypeScript
- Element Plus (tables only — `<el-table>`, `<el-form>`, `<el-dialog>`)
- Pinia (auth store, portfolio store)
- Vue Router
- axios with interceptor → on 401 redirect to `/login`

### 4.2 Pages

| Route                    | Page                  | Notes                                                |
|--------------------------|-----------------------|------------------------------------------------------|
| `/login`                 | Login                 | Username + password; calls `/api/auth/login`         |
| `/portfolio`             | My Portfolio          | Table of tracked stocks; add (with stock search), delete, edit note |
| `/stock/:tsCode`         | Stock Detail          | Tabs: Analysis · History · Today's Draft             |
| `/stock/:tsCode` → Analysis | Spread model + reference tables; "use draft / manual override" form |
| `/stock/:tsCode` → History  | Paginated daily bars table                       |
| `/stock/:tsCode` → Draft    | OHLC input form (today, pre-filled if exists)    |
| `/settings`              | Settings              | Change password · Tushare token override · API tokens |
| `/admin/users`           | (admin only) User mgmt| Create / disable / reset-password                    |
| `/admin/sync`            | (admin only) Manual sync triggers + last-run status |

### 4.3 Auth flow

1. App boot: `GET /api/auth/me` — if 401, redirect `/login`.
2. After login, Pinia stores user; router guard checks `role === 'admin'` for
   `/admin/*`.
3. CSRF is unnecessary because session cookie is `SameSite=Lax` and all
   mutating endpoints require either same-site cookie or Bearer token.

### 4.4 Build & embed

```
cd web && pnpm install && pnpm build
# produces web/dist; Go embed picks it up at compile time
```

Dev mode: Vite proxy `/api → http://localhost:8080` and Go server runs with
`--dev` to skip embed (serve from filesystem if `web/dist` present, otherwise
return a "frontend not built" stub on `/`).

---

## 5. CLI + Skill

### 5.1 `stockctl` (cobra)

Pure remote client. Reads server URL + token from (in order) flags →
`STOCKCTL_*` env → `~/.config/stockctl/config.yaml`.

```
stockctl login                         # interactive, stores Bearer token in config
stockctl logout

stockctl portfolio list
stockctl portfolio add 600519
stockctl portfolio rm 600519

stockctl stock search "茅台"
stockctl stock analysis 600519 \
        [--actual-open 1620] [--high 1650] [--low 1600] [--close 1630] \
        [--use-draft] [--format table|json]
stockctl stock history 600519 --from 20250101 --to 20250514

stockctl draft set 600519 --open 1620 --high 1650 --low 1600 --close 1630
stockctl draft get 600519
stockctl draft clear 600519

stockctl admin user create alice --role user
stockctl admin sync bars
stockctl admin sync stocklist --csv ./stock_basic.csv

stockctl version
```

`--format json` returns the raw `/api/analysis/...` payload so the skill can
re-render.

### 5.2 Skill (`.claude/skills/stock-analyst/`)

The existing SKILL.md is rewritten so that:

- Trigger keywords (买点 / 卖点 / 价差 / 区间 / 持仓 / 预测 …) stay identical.
- Data acquisition swaps `python stock.py …` → `stockctl stock analysis CODE
  --format json`.
- Output template (model table, reference table, "建议买入区间", "建议止盈位")
  is preserved verbatim.
- Skill assumes `stockctl` is on PATH and already logged in; if not, instructs
  the user to run `stockctl login`.

### 5.3 Skill ↔ server contract

The skill only consumes JSON via the CLI; it never hits the HTTP API directly.
This keeps the skill platform-agnostic and avoids embedding API tokens in the
skill prompt.

---

## 6. Scheduling, Config, Deployment

### 6.1 Internal scheduler (`robfig/cron/v3`)

Jobs registered at startup, each guarded by a singleflight key so manual
triggers and cron runs can't pile up:

| Job             | Cron (server TZ Asia/Shanghai) | Action                                            |
|-----------------|--------------------------------|---------------------------------------------------|
| `daily-fetch`   | `0 22 * * 1-5`                 | For every distinct `ts_code` referenced in any portfolio: fetch missing trading days from Tushare `daily`, compute 6 spreads, upsert into `daily_bars`. |
| `stocklist-sync`| `0 3 * * 0`                    | Tushare `stock_basic`; upsert into `stocks`.      |

Manual triggers:

- `POST /api/admin/bars/sync` (or `stockctl admin sync bars`) → enqueues
  `daily-fetch` immediately (singleflight-deduped).
- `POST /api/admin/stocks/sync` → `stocklist-sync`.
- `POST /api/admin/stocks/import-csv` → bulk upsert from uploaded CSV; same
  columns as Tushare `stock_basic`.

A `job_runs` table records each run: `id, job_name, started_at, finished_at,
status, message`. Admin UI / CLI surfaces the last run per job.

### 6.2 Tushare token selection

Per request / job:

```
effective_token = user.TushareToken (if set)
                  else config.tushare.default_token
```

Scheduler runs use the server default token.

### 6.3 Config (viper)

`/etc/stockd/config.yaml` (or `--config` flag):

```yaml
server:
  listen: ":8443"
  base_url: "https://stock.example.com"
  tls:
    enabled: true
    cert_file: /etc/stockd/tls/fullchain.pem
    key_file:  /etc/stockd/tls/privkey.pem
  session_secret: "change-me"        # required, >= 32 bytes

database:
  driver: sqlite                     # sqlite | mysql | postgres
  dsn: "/var/lib/stockd/stock.db"    # driver-specific

tushare:
  default_token: ""                  # required for sync to work
  base_url: "http://api.tushare.pro"
  timeout: 30s

scheduler:
  daily_fetch_cron:    "0 22 * * 1-5"
  stocklist_sync_cron: "0 3 * * 0"
  enabled: true

logging:
  level: info
  format: json
```

Environment variable mapping (`pkg/stockd/config/env.go`), following the `mtkd`
`Envs` pattern so every config key can be overridden by an env var:

```go
var Envs = map[string]string{
    "STOCKD_HTTP_ADDR":        "server.listen",
    "STOCKD_TLS_CERT":         "server.tls.certFile",
    "STOCKD_TLS_KEY":          "server.tls.keyFile",
    "STOCKD_SESSION_SECRET":   "server.sessionSecret",
    "STOCKD_DB_DRIVER":        "database.driver",
    "STOCKD_DB_DSN":           "database.dsn",
    "STOCKD_TUSHARE_TOKEN":    "tushare.defaultToken",
    "STOCKD_TUSHARE_BASE_URL": "tushare.baseUrl",
    "STOCKD_LOG_LEVEL":        "logging.level",
    "STOCKD_LOG_FORMAT":       "logging.format",
}
```

Viper loads in this priority (high → low):
`--flag` → `STOCKD_*` env → `~/.config/stockd/config.yaml` → `/etc/stockd/config.yaml` → defaults.

`stockd` (no subcommand) reads config, opens DB, AutoMigrates, starts scheduler
+ HTTP server. `stockd version` prints version / commit / build date.

### 6.4 First-run bootstrap

If `users` table is empty at startup, server prints (and logs) a randomly
generated admin password and seeds `admin / <random>`. Operator then logs in
and changes it.

### 6.5 Deployment

- Single static binary (`stockd`) ships with embedded frontend.
- Reverse proxy optional — `stockd` can terminate TLS itself.
- SQLite is the default; pointing `database.dsn` at MySQL / Postgres is a pure
  config change.
- systemd unit lives in `deploy/stockd.service` (template).

---

## 7. Testing + Task Breakdown

### 7.1 Testing strategy

| Layer        | Tooling                              | Coverage target |
|--------------|--------------------------------------|-----------------|
| Go unit      | `testing` + `testify`                | ≥ 80 %          |
| Go DB layer  | `go-sqlmock` + real sqlite in-memory | -               |
| Go HTTP      | `httptest` + gin                     | golden JSON     |
| Parity tests | replay fixtures captured from current Python (`stock.py show 600519 --json`) and assert byte-identical `model_table` / `reference_table` | 100 % match required |
| Frontend     | Vitest (unit) + Playwright (smoke: login → add stock → analysis → draft) | smoke green     |

Parity fixtures live in `pkg/analysis/testdata/` and are produced once via a
small `tools/dump_python_fixture.py` helper run against the legacy code.

### 7.2 Task breakdown (50 tasks, P0 → P6)

P0 = foundation; later phases parallelisable where noted.

#### P0 — Scaffolding (sequential)

1. Init Go module (`go 1.23`), add deps (gin, gorm, cobra, viper, logrus,
   robfig/cron, gin-contrib/static, gin-contrib/sessions, swaggo, bcrypt,
   testify, go-sqlmock).
2. Create repo layout (`cmd/`, `pkg/`, `web/`, `embed.go` stub).
3. Add `Makefile` targets: `build`, `test`, `web-build`, `lint`.
4. CI skeleton (GitHub Actions): Go test + Vue build.

#### P1 — Shared libraries (parallelisable after P0)

5. `pkg/shared/stockcode` — port `to_tushare_code()` + table-driven tests.
6. `pkg/shared/spread` — port `compute_spreads()` + tests.
7. `pkg/shared/window` — `_compute_window_means` helper + tests.
8. `pkg/tushare` — HTTP client, `Daily`, `StockBasic` calls, token per-call,
   retry/backoff, `httptest`-driven tests.
9. `pkg/analysis` — port `_build_spread_model_table` +
   `_build_reference_table` + trading-plan; load Python fixtures and assert
   equality.

#### P2 — Server core (after P1)

10. `pkg/stockd/config` — viper loader + validation + env var mapping (`env.go`).
11. `pkg/stockd/db` — multi-driver init (`db.Repo` struct), AutoMigrate,
    connection pool, custom `GormLogger`, sqlite-in-memory + `sqlmock` test
    harness (`InitMockDB`).
12. `pkg/stockd/models` — all GORM models from §2 (including `JobRun`).
13. `pkg/stockd/utils` — `HTTPResponse` / `HTTPResponseFailed` envelope,
    error-code constants (`errors.go`), i18n message maps (`msg.go`, `zh`+
    `en`).
14. `pkg/stockd/middleware` — `requestID`, `panicRecovery`, `auth`
    (session + Bearer), `logger` middlewares.
15. Log initialisation (`logrus` + GORM logger bridge).
16. First-run bootstrap (seed admin if `users` empty).

#### P3 — Services (parallelisable, all after P2)

Single `pkg/stockd/services` package with a `StockD` struct (mirrors `mtkd`
`services.MtkD`). Constructed via functional `Option` pattern.

17. `services` skeleton — `StockD` struct, `repo *db.Repo`, `logger`,
    `conf`, `Option` constructor.
18. `services/user` methods — CRUD, role checks, password change.
19. `services/token` methods — issue / revoke API tokens, return plain text
    once.
20. `services/stock` methods — search, get, CSV import (streaming parser),
    Tushare sync.
21. `services/portfolio` methods — list / add / remove / note.
22. `services/draft` methods — get today / upsert / clear, validation.
23. `services/bars` methods — official history query, sync (incremental,
    compute spreads, upsert).
24. `services/analysis` methods — orchestrate bars + draft + override params
    → `pkg/analysis`.
25. `services/scheduler` methods — register cron jobs, `singleflight`, record
    `job_runs`.

#### P4 — HTTP layer (after P3)

26. Mount middleware stack (`requestID` → `panicRecovery` → `CORS` → `auth`
    → `logger`) in `http/router.go`.
27. Response envelope + error mapper (`utils.HTTPRequestSuccess` /
    `utils.HTTPRequestFailedV4`).
28. Auth routes (`/api/auth/*`).
29. Admin user routes.
30. Self-service routes (`/api/me/*`).
31. Stock + portfolio routes.
32. Bars + draft routes.
33. Analysis route (with query-param overrides).
34. Admin sync routes (bars / stocklist / CSV upload).
35. Swagger annotations + mount `/swagger`.
36. Static SPA mount with `index.html` fallback (from `embed.go`).
37. TLS termination + graceful shutdown.

#### P5 — CLI + Skill (parallelisable, after P4 envelope + core routes stable)

38. `stockctl` cobra skeleton + config loader (flags / env / file).
39. `stockctl login` / `logout` / `version`.
40. `stockctl portfolio` group.
41. `stockctl stock search|analysis|history` (table + JSON renderers; reuse
    Python output template for `table`).
42. `stockctl draft` group.
43. `stockctl admin user|sync` group.
44. Rewrite `.claude/skills/stock-analyst/SKILL.md` to call
    `stockctl … --format json`.

#### P6 — Frontend (parallelisable with P5, after P4 envelope + core routes stable)

45. Vite scaffold (Vue 3 + TS + Element Plus + Pinia + Router); axios client
    with 401 interceptor.
46. Login page + auth store.
47. Portfolio page (table + add dialog with stock search + delete).
48. Stock detail page (Analysis / History / Draft tabs).
49. Settings page (password + Tushare token + API tokens).
50. Admin pages (users + manual sync + last-run status); Playwright smoke test.

### 7.3 Parallelism summary

- After P0, P1 tasks (5–9) run in parallel.
- After P1, P2 (10–16) runs sequentially (config → db → models → utils →
  middleware → logger → bootstrap).
- After P2, P3 services (17–25) can split across 2 streams:
  - Stream A: user / token / portfolio / draft
  - Stream B: stock / bars / analysis / scheduler
- P4 handlers (26–37) can be cut by route group; finalise envelope + middleware
  first (26–27).
- P5 (CLI/skill) and P6 (frontend) run in parallel once P4 envelope + ~5 core
  routes are stable (no need to wait for all 37).
- P4 HTTP handlers can be cut by route group; finalise envelope first (23).
- P5 (CLI/skill) and P6 (frontend) run in parallel once P4 stabilises.

---

## 8. Out of scope (deferred)

- Full trading ledger (shares + cost basis + PnL) — future v2.
- Charts / candlesticks — UI is tables only for now.
- Backups / restores.
- Multi-tenant org separation (single-tenant, admin-managed user pool).
- SSO / OAuth.
