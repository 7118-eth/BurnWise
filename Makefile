.PHONY: all build run test test-short test-cover clean install dev lint help

# Variables
APP_NAME := budget
BUILD_DIR := ./bin
GO_FILES := $(shell find . -name '*.go' -type f)
TEST_DATA_DIR := ./test/data
COVERAGE_DIR := ./coverage

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/budget

# Run the application
run: build
	@$(BUILD_DIR)/$(APP_NAME)

# Install the application
install: build
	@echo "Installing $(APP_NAME)..."
	@go install ./cmd/budget

# Run all tests
test:
	@echo "Running tests..."
	@mkdir -p $(TEST_DATA_DIR)
	@go test ./... -v

# Run tests without external API calls
test-short:
	@echo "Running tests (short mode)..."
	@mkdir -p $(TEST_DATA_DIR)
	@go test -short ./... -v

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@mkdir -p $(TEST_DATA_DIR) $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(TEST_DATA_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -rf data/*.db

# Development mode with live reload
dev:
	@echo "Starting development mode..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	@air

# Run linters
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go mod tidy
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Check for vulnerabilities
vuln:
	@echo "Checking for vulnerabilities..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Database operations
db-clean:
	@echo "Cleaning database..."
	@rm -f data/budget.db data/budget.db-*

db-backup:
	@echo "Backing up database..."
	@mkdir -p backups
	@cp data/budget.db backups/budget-$$(date +%Y%m%d-%H%M%S).db

# Help target
help:
	@echo "Budget Tracker Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build      - Build the application"
	@echo "  make run        - Build and run the application"
	@echo "  make install    - Install the application"
	@echo "  make test       - Run all tests"
	@echo "  make test-short - Run tests without external API calls"
	@echo "  make test-cover - Run tests with coverage report"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make dev        - Run in development mode with live reload"
	@echo "  make lint       - Run linters"
	@echo "  make fmt        - Format code"
	@echo "  make tidy       - Tidy dependencies"
	@echo "  make vuln       - Check for vulnerabilities"
	@echo "  make db-clean   - Clean database files"
	@echo "  make db-backup  - Backup database"
	@echo "  make help       - Show this help message"