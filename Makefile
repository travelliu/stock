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
