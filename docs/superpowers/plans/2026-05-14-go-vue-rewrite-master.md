# Go + Vue Rewrite — Master Plan (Index)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement the phase plans task-by-task. Each phase lives in its own plan file; start with P0 and work through phases in dependency order.

**Goal:** Rewrite the Python A-share spread-analysis CLI as a Go + Vue 3 multi-user web app (server `stockd`, remote CLI `stockctl`, Vue SPA, Claude skill) with **bit-for-bit parity** with the existing Python analysis output.

**Architecture:** Single repo, dual Go binaries, shared `pkg/`, Vue frontend embedded via `//go:embed`. SQLite by default, switchable to MySQL/Postgres. Dual auth (session cookie for browser, Bearer token for CLI/skill). Internal scheduler (`robfig/cron`) replaces all cron jobs.

**Tech Stack:** Go 1.23+ (gin, gorm, cobra, viper, logrus, robfig/cron, gin-contrib/sessions/static, swaggo, bcrypt, testify, go-sqlmock), Vue 3 + TypeScript + Vite + Element Plus + Pinia + Vue Router + axios.

**Reference spec:** [docs/superpowers/specs/2026-05-14-go-vue-rewrite-design.md](../specs/2026-05-14-go-vue-rewrite-design.md)

---

## Phase Index

| Phase | Plan file | Tasks | Dependencies | Output |
|-------|-----------|-------|--------------|--------|
| P0 — Scaffolding | [P0](./2026-05-14-p0-scaffolding.md) | 1–4 | — | Go module + repo layout + Makefile + CI skeleton |
| P1 — Shared libraries | [P1](./2026-05-14-p1-shared-libs.md) | 5–9 | P0 | `pkg/shared/{stockcode,spread,window}`, `pkg/tushare`, `pkg/analysis` with **parity tests** vs Python |
| P2 — Server core | [P2](./2026-05-14-p2-server-core.md) | 10–14 | P1 | viper config + env mapping, `db.Repo` multi-driver + AutoMigrate, models, auth, first-run bootstrap |
| P3 — Services | [P3](./2026-05-14-p3-services.md) | 15–22 | P2 | `StockD` service struct with user/token/stock/portfolio/draft/bars/analysis/scheduler methods |
| P4 — HTTP layer | [P4](./2026-05-14-p4-http-layer.md) | 23–34 | P3 | middleware stack, response envelope, gin routes, Swagger, TLS, embedded SPA mount, graceful shutdown |
| P5 — CLI + Skill | [P5](./2026-05-14-p5-cli-skill.md) | 35–41 | P4 envelope + core routes stable | `stockctl` cobra commands + rewritten Claude skill |
| P6 — Frontend | [P6](./2026-05-14-p6-frontend.md) | 42–47 | P4 envelope + core routes stable (parallel with P5) | Vue 3 SPA with login/portfolio/detail/settings/admin pages + Playwright smoke |

## Parallelism

- After **P0** completes, all P1 tasks (5–9) can run in parallel.
- **P2** is sequential (config→db→models→auth→bootstrap) but quick.
- After P2, **P3** splits into 2 streams: (16,17,18,19) user/token/stock/portfolio and (20,21,22) draft/bars/analysis/scheduler.
- **P4** mounts middleware stack (task 23) first, finalises envelope (task 24), then route groups (25–33) are parallelisable.
- **P5** and **P6** run in parallel once P4 envelope + ~5 core routes are stable.

## TaskCreate dependency chain (suggested)

When loading the 47 tasks into the task list, use this dependency wiring:

```
P0: 1 → 2 → 3 → 4
P1: 4 → {5,6,7,8} → 9
P2: 9 → 10 → 11 → 12 → 13 → 14
P3: 14 → 15 → {16,17,18,19} and {20,21,22}
P4: {15..22} → 23 → 24 → {25,26,27,28,29,30,31,32,33} → 34
P5: 34 → 35 → {36,37,38,39,40} → 41
P6: 34 → 42 → {43,44,45,46} → 47
```

`finishing-a-development-branch` runs once all phases are green.

## Commit policy

Each task ends with a commit (per the per-task plan). Use conventional-commits prefixes: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`. Co-author trailers are disabled globally — do not add them.

## Acceptance gate (per phase)

A phase is "done" when:

- All its tasks are checked off in the plan file.
- `make test` is green in the relevant scope.
- The phase's exit criterion (listed in its plan header) is verifiable from the CLI / browser.

## Out of scope (deferred)

Per spec §8: full trading ledger, charts/candlesticks, backups, multi-tenant org separation, SSO. Do not implement these.
