# Makefile for terraform-provider-danubedata

BINARY_NAME=terraform-provider-danubedata
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Go settings
GO=go
GOFLAGS=-trimpath
LDFLAGS=-s -w -X main.version=$(VERSION)

# Install paths
OS_ARCH=$(GOOS)_$(GOARCH)
INSTALL_PATH=~/.terraform.d/plugins/registry.terraform.io/AdrianSilaghi/danubedata/0.0.1/$(OS_ARCH)

.PHONY: all build install test testacc clean fmt lint docs help

all: build

## Build the provider
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)

## Install the provider locally for development
install: build
	mkdir -p $(INSTALL_PATH)
	cp $(BINARY_NAME) $(INSTALL_PATH)/

## Run unit tests
test:
	$(GO) test -v -race ./...

## Run unit tests with coverage
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## Run acceptance tests (requires DANUBEDATA_API_TOKEN)
testacc:
	TF_ACC=1 $(GO) test -v -timeout 60m ./internal/resources/... ./internal/datasources/...

## Run specific acceptance tests
testacc-%:
	TF_ACC=1 $(GO) test -v -timeout 30m -run "TestAcc$*" ./internal/resources/... ./internal/datasources/...

## Format code
fmt:
	$(GO) fmt ./...
	gofumpt -l -w .

## Lint code
lint:
	golangci-lint run

## Generate documentation
docs:
	$(GO) generate ./...
	tfplugindocs generate

## Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/

## Tidy dependencies
tidy:
	$(GO) mod tidy

## Download dependencies
deps:
	$(GO) mod download

## Verify dependencies
verify:
	$(GO) mod verify

## Run all checks before committing
check: fmt lint test build
	@echo "All checks passed!"

## Show help
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Examples:"
	@echo "  make build         Build the provider"
	@echo "  make install       Install locally for development"
	@echo "  make test          Run unit tests"
	@echo "  make testacc       Run acceptance tests"
	@echo "  make testacc-Vps   Run VPS acceptance tests only"
