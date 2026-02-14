build: ## Build binary for Lambda
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -tags lambda.norpc -o bootstrap ./src

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofmt -s -l -w .

fmtcheck: ## Check formatting
	@UNFORMATTED=$$(gofmt -s -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "Files not formatted:"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi

check: test lint vet fmtcheck ## Run checks

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

.PHONY: test lint vet fmt fmtcheck check help