.PHONY: all build test coverage coverage-summary bench lint clean help

all: test build

build: ## Build the deck binary
	go build -o deck .

test: ## Run all unit tests with statement coverage
	go test -v -cover ./...

coverage: ## Generate HTML statement coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage HTML report generated at coverage.html"

coverage-summary: ## Print function-level statement coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

bench: ## Run performance benchmarks
	go test -bench=. -benchmem ./...

lint: ## Check formatting and run go vet
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files found:" && gofmt -l . && exit 1)
	go vet ./...

fmt: ## Format all Go source files
	gofmt -w .

clean: ## Remove build artifacts and coverage reports
	rm -f deck coverage.out coverage.html

help: ## Display help options
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'
