# Stock — Go + Vue rewrite Makefile

GO         ?= go
PNPM       ?= pnpm
BIN        := $(CURDIR)/bin
COMPOSE    := docker compose

V  = 0
Q  = $(if $(filter 1,$V),,@)
M  = $(shell printf "\033[34;1m▶\033[0m")

BASE_VERSION = $(shell grep 'version = ' pkg/version/version.go | sed -E 's/.*"(.+)".*/\1/')
GIT_COMMIT   = $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
GIT_TAG      = $(shell git describe --tags --exact-match 2>/dev/null || echo "")
GIT_DIRTY    = $(shell test -n "`git status --porcelain`" && echo "-dirty" || echo "")
BUILD_TIME   = $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Final version: git tag if present, otherwise base version from code
ifdef VERSION
	BINARY_VERSION = $(VERSION)
else ifneq ($(GIT_TAG),)
	BINARY_VERSION = $(GIT_TAG)
else
	BINARY_VERSION = $(BASE_VERSION)
endif

LDFLAGS := -s -w \
	-X stock/pkg/version.version=$(BINARY_VERSION)$(GIT_DIRTY) \
	-X stock/pkg/version.gitCommit=$(GIT_COMMIT) \
	-X stock/pkg/version.gitTag=$(GIT_TAG) \
	-X stock/pkg/version.buildTimestamp=$(BUILD_TIME)

GOLDFLAGS = -ldflags '$(LDFLAGS)'

.DEFAULT_GOAL := help

.PHONY: all build test lint fmt vet clean info help web-build \
        dev-up dev-down dev-logs \
        selfhost selfhost-build selfhost-stop selfhost-logs

##@ Build

all: build ## Build all binaries (default)

build: fmt bin/stockd bin/stockctl ## Build stockd and stockctl

bin/%: cmd/%/main.go ; $(info $(M) running build $*) @
	$(GO) build -trimpath $(GOLDFLAGS) -o $@ $<

web-build: ; $(info $(M) building web) @ ## Build frontend into web/dist
	@if [ -f web/package.json ]; then \
		cd web && $(PNPM) install --frozen-lockfile && $(PNPM) build; \
	else \
		echo "web/package.json not found — skipping frontend build"; \
	fi

##@ Quality

test: ; $(info $(M) running tests) @ ## Run Go tests with race detector
	$(GO) test -race -cover ./...

lint: vet ; $(info $(M) running lint) @ ## Run go vet (and staticcheck if installed)
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || \
		echo "staticcheck not installed — skipping"

fmt: ; $(info $(M) running gofmt) @ ## Run go fmt on all source files
	$(GO) fmt ./...

vet: ; $(info $(M) running go vet) @ ## Run go vet
	$(GO) vet ./...

##@ Info

info: ; $(info) @ ## Print build info
	@echo "Base Version:   \"$(BASE_VERSION)\""
	@echo "Binary Version: \"$(BINARY_VERSION)$(GIT_DIRTY)\""
	@echo "Git Commit:     \"$(GIT_COMMIT)\""
	@echo "Git Tag:        \"$(GIT_TAG)\""
	@echo "Build Time:     \"$(BUILD_TIME)\""

##@ Dev (local)

dev-up: ; $(info $(M) starting dev MySQL) @ ## Start local MySQL (for go run ./cmd/stockd)
	$(COMPOSE) up -d mysql

dev-down: ; $(info $(M) stopping dev MySQL) @ ## Stop local MySQL
	$(COMPOSE) down

dev-logs: ; $(info $(M) tailing MySQL logs) @ ## Tail MySQL logs
	$(COMPOSE) logs -f mysql

##@ Self-host

selfhost: ; $(info $(M) starting self-hosted stack) @ ## Start stack with pre-built image
	$(COMPOSE) up -d

selfhost-build: web-build ; $(info $(M) building and starting self-hosted stack) @ ## Build web+image from source and start
	@if [ ! -f .env ]; then cp .env.example .env; fi
	VERSION=$(BINARY_VERSION)$(GIT_DIRTY) COMMIT=$(GIT_COMMIT) \
		$(COMPOSE) -f docker-compose.yml -f docker-compose.self.yml up -d --build
	@echo "==> Waiting for /health..."
	@for i in $$(seq 1 30); do \
		if curl -sf http://localhost:$${STOCKD_PORT:-8443}/health > /dev/null 2>&1; then \
			echo "✓ stockd is ready at http://localhost:$${STOCKD_PORT:-8443}"; \
			break; \
		fi; \
		sleep 2; \
	done

selfhost-stop: ; $(info $(M) stopping self-hosted stack) @ ## Stop the self-hosted stack
	$(COMPOSE) -f docker-compose.yml -f docker-compose.self.yml down

selfhost-logs: ; $(info $(M) tailing stockd logs) @ ## Tail stockd logs
	$(COMPOSE) -f docker-compose.yml -f docker-compose.self.yml logs -f stockd

##@ Cleanup

clean: ; $(info $(M) cleaning) @ ## Remove build artifacts
	rm -rf $(BIN) coverage.out coverage.html

##@ Help

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*## "; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5); next} \
		/^[a-zA-Z0-9_.-]+:.*## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)