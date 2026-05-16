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
        prod-build prod-up prod-down prod-logs prod-restart

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

##@ Cleanup

clean: ; $(info $(M) cleaning) @ ## Remove build artifacts
	rm -rf $(BIN) coverage.out coverage.html

##@ Help

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*## "; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5); next} \
		/^[a-zA-Z0-9_.-]+:.*## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

selfhost-build: ; $(info) @ ## Build backend/web from the current checkout and start the self-hosted stack
	@if [ ! -f .env ]; then \
		echo "==> Creating .env from .env.example..."; \
		cp .env.example .env; \
		JWT=$$(openssl rand -hex 32); \
		if [ "$$(uname)" = "Darwin" ]; then \
			sed -i '' "s/^STOCKD_SERVER_SESSION_SECRET=.*/JWT_SECRET=$$JWT/" .env; \
		else \
			sed -i "s/^STOCKD_SERVER_SESSION_SECRET=.*/JWT_SECRET=$$JWT/" .env; \
		fi; \
		echo "==> Generated random JWT_SECRET"; \
	fi
	@echo "==> Building Stockd from the current checkout..."
	docker compose -f docker-compose.yml -f docker-compose.self.yml up -d --build
	@echo "==> Waiting for backend to be ready..."
	@for i in $$(seq 1 30); do \
		if curl -sf http://localhost:$${STOCKD_PORT:-8443}/health > /dev/null 2>&1; then \
			break; \
		fi; \
		sleep 2; \
	done
	@if curl -sf http://localhost:$${STOCKD_PORT:-8443}/health > /dev/null 2>&1; then \
		echo ""; \
		echo "✓ stockd is running!"; \
		echo "  Frontend: http://localhost:$${STOCKD_PORT:-8443}"; \
		echo ""; \
	else \
		echo ""; \
		echo "Services are still starting. Check logs:"; \
		echo "  docker compose -f docker-compose.self.yml logs"; \
	fi
	
selfhost-stop: ## Stop the self-hosted Docker Compose stack
	@echo "==> Stopping Stockd services..."
	docker compose -f docker-compose.yml down
	@echo "✓ All services stopped."