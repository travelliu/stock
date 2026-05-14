# P0 — Scaffolding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lay the Go module foundation: dependencies pinned, repo layout created, Makefile + CI skeleton in place. Exit criterion: `make build` succeeds (even with stub entrypoints), `make test` passes (zero tests), GitHub Actions runs `go test ./... && pnpm build` (allowed to no-op on missing `web/`).

**Architecture:** Single Go module rooted at `stock` (or current GitHub path — confirm before init). Dual-binary layout in `cmd/`; shared code in `pkg/`. The root package is `package stock` and will own `embed.go` (added in P4) so it can embed `web/dist` without `..` paths.

**Tech Stack:** Go 1.23+, GNU Make, GitHub Actions.

**Reference spec:** `docs/superpowers/specs/2026-05-14-go-vue-rewrite-design.md` §1.1, §7.2 (P0).

---

## File overview

| File | Responsibility |
|------|----------------|
| `go.mod` / `go.sum` | Module name + pinned deps |
| `cmd/stockd/main.go` | Server entrypoint stub (prints `stockd vX.Y.Z`, exits 0) |
| `cmd/stockctl/main.go` | CLI entrypoint stub (prints help, exits 0) |
| `pkg/shared/.gitkeep`, `pkg/stockd/.gitkeep`, `pkg/stockctl/.gitkeep` | Reserve subdirs |
| `web/.gitkeep` | Reserve frontend dir |
| `embed.go` | **Deferred to P4** — stubbed via `web/dist/.gitkeep` so embed compiles |
| `Makefile` | `build`, `test`, `web-build`, `lint`, `clean` targets |
| `.github/workflows/ci.yml` | Run Go test + Vue build (Vue allowed to skip when `web/package.json` absent) |
| `.gitignore` | Add `/bin/`, `/web/node_modules/`, `/web/dist/`, `*.test`, `coverage.out` |

The Python sources (`stock.py`, `analysis.py`, `fetcher.py`, `db.py`, `config.py`, `company.py`, `tests/`, `data/`, `reports/`, `scripts/`, `requirements.txt`, `.venv/`, `__pycache__/`) **stay in the repo for the duration of the rewrite** — they are the parity reference. Removal happens after P5 stabilises.

---

### Task 1: Init Go module + dependencies

**Files:**
- Create: `go.mod`
- Modify (will be created/updated by `go get`): `go.sum`
- Create: `.gitignore`

- [ ] **Step 1: Confirm module path**

Run: `git remote get-url origin`
Expected: something like `git@github.com:travelliu/stock.git` or `https://stock.git`. Use the resulting path as the Go module path (`stock`). If the remote differs, adjust below accordingly and note the divergence in the task notes.

- [ ] **Step 2: Initialise module**

Run: `go mod init stock`
Expected: creates `go.mod` with `module stock` and `go 1.23` (or current toolchain).

- [ ] **Step 3: Add server-side dependencies**

Run:
```bash
go get github.com/gin-gonic/gin@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/sqlite@latest
go get gorm.io/driver/mysql@latest
go get gorm.io/driver/postgres@latest
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/sirupsen/logrus@latest
go get github.com/robfig/cron/v3@latest
go get github.com/gin-contrib/sessions@latest
go get github.com/gin-contrib/static@latest
go get github.com/swaggo/gin-swagger@latest
go get github.com/swaggo/files@latest
go get golang.org/x/crypto/bcrypt@latest
go get github.com/stretchr/testify@latest
go get github.com/DATA-DOG/go-sqlmock@latest
go get golang.org/x/sync/singleflight@latest
go get github.com/gin-contrib/cors@latest
go get golang.org/x/term@latest
go get gopkg.in/yaml.v3@latest
```
Expected: `go.sum` populated; `go.mod` lists each as a `require`.

- [ ] **Step 4: Tidy**

Run: `go mod tidy`
Expected: no errors. `go.mod` and `go.sum` consistent.

- [ ] **Step 5: Write `.gitignore`**

Create `.gitignore` (overwriting if any exists for Python — merge entries, don't drop existing Python rules):
```
# Go
/bin/
*.test
coverage.out
coverage.html

# Frontend
/web/node_modules/
/web/dist/

# Existing Python (preserve)
__pycache__/
*.pyc
.venv/
.pytest_cache/
.env
data/*.db
```

- [ ] **Step 6: Verify**

Run: `go build ./... 2>&1 | head`
Expected: no errors (zero packages to build is fine).

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum .gitignore
git commit -m "chore: init Go module and pin dependencies"
```

---

### Task 2: Create repo layout + entrypoint stubs

**Files:**
- Create: `cmd/stockd/main.go`
- Create: `cmd/stockctl/main.go`
- Create: `pkg/shared/.gitkeep`, `pkg/stockd/.gitkeep`, `pkg/stockctl/.gitkeep`, `pkg/tushare/.gitkeep`, `pkg/analysis/.gitkeep`
- Create: `pkg/stockd/middleware/.gitkeep`, `pkg/stockd/utils/.gitkeep`, `pkg/stockd/models/.gitkeep`, `pkg/stockd/services/.gitkeep`
- Create: `web/.gitkeep`, `web/dist/.gitkeep`
- Create: `embed.go` (root) — declared in P4, **here we add a stub** so the root package compiles without `//go:embed`.

- [ ] **Step 1: Write `cmd/stockd/main.go`**

```go
package main

import "fmt"

const Version = "0.0.0-dev"

func main() {
	fmt.Printf("stockd %s\n", Version)
}
```

- [ ] **Step 2: Write `cmd/stockctl/main.go`**

```go
package main

import "fmt"

const Version = "0.0.0-dev"

func main() {
	fmt.Printf("stockctl %s\n", Version)
}
```

- [ ] **Step 3: Write `embed.go` stub at repo root**

```go
package stock

// Embed-related code lives here; see P4 task 33 for the //go:embed declaration.
// Keeping the package non-empty so other packages can import "stock".

const PackageName = "stock"
```

- [ ] **Step 4: Create directory placeholders**

Run:
```bash
mkdir -p pkg/shared pkg/stockd/{middleware,utils,models,services,config,db} pkg/stockctl pkg/tushare pkg/analysis web/dist
touch pkg/shared/.gitkeep \
      pkg/stockd/middleware/.gitkeep pkg/stockd/utils/.gitkeep \
      pkg/stockd/models/.gitkeep pkg/stockd/services/.gitkeep \
      pkg/stockd/config/.gitkeep pkg/stockd/db/.gitkeep \
      pkg/stockctl/.gitkeep pkg/tushare/.gitkeep pkg/analysis/.gitkeep \
      web/.gitkeep web/dist/.gitkeep
```

- [ ] **Step 5: Build all packages**

Run: `go build ./...`
Expected: no errors. Three packages compile: `stock`, `.../cmd/stockd`, `.../cmd/stockctl`.

- [ ] **Step 6: Run the binaries**

Run: `go run ./cmd/stockd && go run ./cmd/stockctl`
Expected:
```
stockd 0.0.0-dev
stockctl 0.0.0-dev
```

- [ ] **Step 7: Commit**

```bash
git add cmd/ pkg/ web/ embed.go
git commit -m "chore: scaffold cmd/, pkg/, web/ directories with stub mains"
```

---

### Task 3: Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Write the Makefile**

```makefile
# Stock — Go + Vue rewrite Makefile

GO         ?= go
PNPM       ?= pnpm
BIN_DIR    := bin
LDFLAGS    := -s -w -X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)

.PHONY: all build build-stockd build-stockctl web-build test lint clean fmt vet

all: build

build: build-stockd build-stockctl

build-stockd:
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/stockd ./cmd/stockd

build-stockctl:
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/stockctl ./cmd/stockctl

web-build:
	@if [ -f web/package.json ]; then \
		cd web && $(PNPM) install --frozen-lockfile && $(PNPM) build; \
	else \
		echo "web/package.json not found — skipping frontend build"; \
	fi

test:
	$(GO) test -race -cover ./...

lint:
	$(GO) vet ./...
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || \
		echo "staticcheck not installed — skipping"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html
```

- [ ] **Step 2: Verify each target**

Run: `make build`
Expected: `bin/stockd` and `bin/stockctl` produced. `./bin/stockd` prints `stockd <version>`.

Run: `make test`
Expected: `ok` / `no test files` lines, exit 0.

Run: `make lint`
Expected: `go vet` clean; staticcheck message tolerated.

Run: `make web-build`
Expected: prints `web/package.json not found — skipping frontend build`.

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "build: add Makefile with build/test/lint/web-build targets"
```

---

### Task 4: CI skeleton (GitHub Actions)

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write the workflow**

```yaml
name: ci

on:
  push:
    branches: [master, main]
  pull_request:

jobs:
  go:
    name: Go build + test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true
      - name: Go vet
        run: go vet ./...
      - name: Go test
        run: go test -race -cover ./...
      - name: Go build
        run: |
          mkdir -p bin
          go build -o bin/stockd  ./cmd/stockd
          go build -o bin/stockctl ./cmd/stockctl

  web:
    name: Frontend build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: pnpm
          cache-dependency-path: web/pnpm-lock.yaml
        continue-on-error: true
      - name: Frontend build
        run: |
          if [ -f web/package.json ]; then
            cd web && pnpm install --frozen-lockfile && pnpm build
          else
            echo "web/package.json not present yet — skipping"
          fi
```

- [ ] **Step 2: Local sanity check**

Run: `make build && make test && make lint`
Expected: all green.

If `act` is installed locally, optionally run: `act -j go` to simulate the workflow. Otherwise, push will validate.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add GitHub Actions for Go test/build and frontend build"
```

---

## Exit criterion

- [ ] `go build ./...` clean
- [ ] `make test` clean
- [ ] `make build` produces `bin/stockd` and `bin/stockctl`
- [ ] Both binaries run and print their version stubs
- [ ] CI workflow committed (will turn green on the next push)

## Hand-off

Next: [P1 — Shared libraries](./2026-05-14-p1-shared-libs.md). P1 starts populating `pkg/shared/*`, `pkg/tushare`, and `pkg/analysis` with TDD-driven Go ports of the Python logic.
