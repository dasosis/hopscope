# hopscope Makefile.
#
# Recipes use POSIX sh, so run make from Linux, macOS, WSL or Git Bash.
# On plain Windows without make, the equivalent commands are:
#   go build -o bin/hopscope.exe ./cmd/hopscope
#   go test ./...

BINARY := hopscope
PKG    := github.com/dasosis/hopscope
MAIN   := ./cmd/hopscope

# Build metadata injected into internal/version. DATE is the commit time, not
# the wall-clock time, so building the same commit twice gives identical
# binaries (a reproducible build).
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null)
DATE    ?= $(shell git log -1 --format=%cI 2>/dev/null)

# -s -w strips the symbol table and DWARF debug info (smaller binary).
# The Go build info that internal/version reads is kept.
LDFLAGS := -s -w \
	-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.Date=$(DATE)

# CGO_ENABLED=0 produces static binaries and lets us cross-compile for every
# target from one machine with no C toolchain. See docs/decisions/0002.
GOBUILD := CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)"

PLATFORMS := linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64

GOEXE := $(shell go env GOEXE)

.DEFAULT_GOAL := help
.PHONY: help all build test test-race cover lint fmt tidy cross clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-10s %s\n", $$1, $$2}'

all: lint test build ## Lint, test and build

build: ## Build for the host platform into bin/
	$(GOBUILD) -o bin/$(BINARY)$(GOEXE) $(MAIN)

test: ## Run unit tests
	go test -count=1 ./...

test-race: ## Run unit tests with the race detector (needs cgo and a C compiler)
	CGO_ENABLED=1 go test -race -count=1 ./...

cover: ## Run tests and print per-function coverage
	go test -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code with the formatters configured in .golangci.yml
	golangci-lint fmt

tidy: ## Tidy go.mod and go.sum
	go mod tidy

cross: ## Cross-compile for all PLATFORMS into dist/
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "==> $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch $(GOBUILD) -o dist/$(BINARY)-$$os-$$arch$$ext $(MAIN) || exit 1; \
	done

clean: ## Remove build and test output
	rm -rf bin dist coverage.out
