BINARY_NAME := worker
BUILD_DIR := bin
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
LDFLAGS := -ldflags "-s -w"
TAGS := lambda.norpc

.PHONY: build test lint vet fmt fmtcheck clean help

## build: Compile the Lambda binary
build:
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -tags $(TAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./src

## test: Run all tests with coverage
test:
	@echo "==> Running tests..."
	go test -v -race -cover ./...

## lint: Run golangci-lint
lint:
	@echo "==> Running linter..."
	golangci-lint run ./...

## vet: Run go vet
vet:
	@echo "==> Running vet..."
	go vet ./...

## fmt: Format code
fmt:
	@echo "==> Formatting code..."
	gofmt -s -l -w .

## fmtcheck: Check formatting
fmtcheck:
	@echo "==> Checking formatting..."
	@UNFORMATTED=$$(gofmt -s -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "Files not formatted:"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi

## clean: Remove build artifacts
clean:
	@echo "==> Cleaning..."
	@rm -rf $(BUILD_DIR)
	go clean

## help: Show this help message
help:
	@echo "Makefile commands"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
