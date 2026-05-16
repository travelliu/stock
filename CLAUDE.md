# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Chinese A-share intraday swing trading analysis platform. Go backend (stockd) + Vue 3 frontend, with a CLI client (stockctl). Pulls OHLCV data from Tushare Pro API, calculates price spread statistics (高-开, 开-低, 高-低, 开-收, 高-收, 低-收), and generates trading range predictions based on historical spread means.

- 股票代码 后台表存储全部为标准股票代码. 不带.SZ/SH. 只有在和其他外部系统对接时会增加

## Build Commands

```bash
make build          # Build both stockd and stockctl into bin/
make web-build      # Build Vue frontend into web/dist/
make test           # RunStockAnalysis Go tests with race detector
make lint           # RunStockAnalysis go vet + staticcheck
make fmt            # RunStockAnalysis gofmt
make info           # Print version/build info
make clean          # Remove bin/ and coverage files

# RunStockAnalysis a single test
go test -race -run TestFunctionName ./pkg/stockd/services/

# RunStockAnalysis backend locally (requires config.yaml or STOCKD_ env vars)
go run ./cmd/stockd

# RunStockAnalysis CLI client
go run ./cmd/stockctl --server http://localhost:8443 stocks list
```

## Architecture

### Backend (Go)

Two binaries share the `stock` Go module:

- **`cmd/stockd/`** — HTTP server (Gin). Entry point loads config → opens DB → bootstraps admin → starts scheduler → serves API + embedded frontend.
- **`cmd/stockctl/`** — CLI client (Cobra) that calls stockd's REST API remotely.

Core packages under `pkg/`:

| Package | Role |
|---------|------|
| `stockd/config/` | Viper-based YAML config with `STOCKD_` env var overrides |
| `stockd/db/` | GORM connection + AutoMigrate for all models |
| `stockd/http/` | Gin router, handlers, middleware (auth, CORS, gzip) |
| `stockd/services/` | Business logic layer — the `Service` struct holds DB, Tushare client, config, cron, stock cache |
| `stockd/services/analysis/` | Spread calculation and prediction engine |
| `stockd/auth/` | JWT session management, password hashing |
| `stockd/bootstrap/` | First-run admin user seeding |
| `models/` | GORM models: Stock, DailyBar (with embedded Spreads), Portfolio, User, APIToken, AnalysisPrediction, JobRun |
| `tushare/` | Tushare Pro API SDK with retry logic |
| `cli/` | stockctl commands, API client, terminal rendering |
| `version/` | Build-time version injection via ldflags |

### Frontend (Vue 3)

Located in `web/`. Built to `web/dist/` and embedded into the Go binary via `embed.go` (`//go:embed all:web/dist`). Uses Vue 3 Composition API, Element Plus, Pinia, Vue Router, Axios, Vitest + Playwright.

### Key Patterns

- **Service struct** (`services.Service`) is the central dependency container — all handlers receive it.
- **Stock cache**: in-memory `map[string]*Stock` (by code and by tsCode) loaded at startup, protected by `sync.RWMutex`.
- **Scheduler**: robfig/cron for daily bar fetch (weekdays 22:00) and stock list sync (Sundays 03:00), configurable via config.
- **Frontend embedding**: `embed.go` serves Vue build output from the Go binary; `NoRoute` falls back to `index.html` for SPA routing.
- **Spread calculation**: `Spreads` struct embedded in `DailyBar` with 6 spread types (OH, OL, HL, OC, HC, LC). Predictions use `(arithmetic_mean + median) / 2` of historical spreads.

### Configuration

YAML file (default `/etc/stockd/config.yaml`, overridable via `STOCKD_CONFIG` env). All keys overridable with `STOCKD_` prefix (e.g., `STOCKD_DATABASE_DSN`, `STOCKD_TUSHARE_DEFAULT_TOKEN`). Supports `.env` file via gotenv. Local dev uses SQLite; production typically uses MySQL.

### API Structure

REST API under `/api/`. Auth via session cookies (Gin sessions). Admin routes under `/api/admin/` require admin role. User-scoped routes (portfolio, tokens) require authentication. Health check at `/health`. Swagger UI at `/swagger/`.

### Version Injection

Build-time metadata injected via ldflags in Makefile: `pkg/version.version`, `gitCommit`, `gitTag`, `buildTimestamp`.

## Testing

31 test files across the Go codebase. RunStockAnalysis with `make test` (enables race detector). Tests use testify assertions. Key test areas: services, stock cache, analysis predictions, portfolio, auth, bootstrap.
