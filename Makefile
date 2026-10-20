.PHONY: build format lint run test test-race

SOURCE_CODE ?= ./cmd/... ./internal/... ./test/unit/...
REV := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_OUTPUT_DIR := ./dist

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "%-24s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build simple-admin binary from ./cmd/simple-admin
	mkdir -p $(BUILD_OUTPUT_DIR)
	go build -trimpath -ldflags "-s -w -X github.com/Olian04/simple-admin/cmd/simple-admin/version.Revision=$(REV) -X github.com/Olian04/simple-admin/cmd/simple-admin/version.BuildTime=$(BUILD_TIME)" -o $(BUILD_OUTPUT_DIR)/simple-admin ./cmd/simple-admin

run: ## Run simple-admin from ./cmd/simple-admin
	go run ./cmd/simple-admin

format: ## Run go fmt and gofmt
	go fmt ./...
	gofmt -w .

lint: ## Run go vet, module verify, vuln scan (report-only), golangci (staticcheck, errcheck, gosec, revive)
	go vet ./...
	go mod verify
	# govulncheck is report-only: remaining findings are Go stdlib advisories tied
	# to the toolchain release (not project code). Keep the toolchain patched.
	-go tool govulncheck $(SOURCE_CODE)
	go tool golangci-lint run $(SOURCE_CODE)

test: ## Run unit tests
	go test -shuffle=on -timeout 120s $(SOURCE_CODE)

test-race: ## Run unit tests with race detector
	go test -race -shuffle=on -timeout 180s $(SOURCE_CODE)
