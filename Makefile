# Stock — Go + Vue rewrite Makefile

GO         ?= go
PNPM       ?= pnpm
BIN        := $(CURDIR)/bin

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

.PHONY: all build test lint fmt vet clean info help web-build

all: build ## Build all binaries (default)

build: fmt bin/stockd bin/stockctl ## Build stockd and stockctl

bin/%: cmd/%/main.go ; $(info $(M) running build $*) @
	$(GO) build -trimpath $(GOLDFLAGS) -o $@ $<

test: ; $(info $(M) running tests) @ ## Run Go tests with race detector
	$(GO) test -race -cover ./...

lint: vet ; $(info $(M) running lint) @ ## Run go vet (and staticcheck if installed)
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || \
		echo "staticcheck not installed — skipping"

fmt: ; $(info $(M) running gofmt) @ ## Run go fmt on all source files
	$(GO) fmt ./...

vet: ; $(info $(M) running go vet) @ ## Run go vet
	$(GO) vet ./...

web-build: ; $(info $(M) running web build) @ ## Build frontend (skips if no package.json)
	@if [ -f web/package.json ]; then \
		cd web && $(PNPM) install --frozen-lockfile && $(PNPM) build; \
	else \
		echo "web/package.json not found — skipping frontend build"; \
	fi

info: ; $(info) @ ## Print build info
	@echo "Base Version:   \"$(BASE_VERSION)\""
	@echo "Binary Version: \"$(BINARY_VERSION)$(GIT_DIRTY)\""
	@echo "Git Commit:     \"$(GIT_COMMIT)\""
	@echo "Git Tag:        \"$(GIT_TAG)\""
	@echo "Build Time:     \"$(BUILD_TIME)\""

help: ; $(info) @ ## Print this help
	@grep -E '^[a-zA-Z1-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

clean: ; $(info $(M) cleaning) @ ## Remove build artifacts
	rm -rf $(BIN) coverage.out coverage.html
