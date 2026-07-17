# API Service Makefile
.PHONY: build run test clean install dev docker-build docker-run help

# Variables
APP_NAME := "API service"
BINARY_NAME := api
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go related variables
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor
GOBIN := $(GOBASE)/bin
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"
BUILD_FLAGS := -trimpath $(LDFLAGS)

# Directories
SRC_DIR := ./cmd/api
BUILD_DIR := ./bin

# Default target
all: clean build

## build: Build the application binary
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## build-linux: Build for Linux (useful for Docker)
build-linux: deps
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(SRC_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)-linux"

## build-all: Build for multiple platforms
build-all: deps
	@echo "Building $(APP_NAME) for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SRC_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SRC_DIR)
	@echo "Built binaries for Linux, macOS, and Windows"

## run: Build and run the application
run: build
	@echo "Running $(APP_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

## dev: Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@if command -v air > /dev/null; then \
		echo "Starting development server with hot reload..."; \
		air -c .air.toml; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running normally..."; \
		$(MAKE) run; \
	fi

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-short: Run short tests only
test-short:
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## deps-update: Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

## lint: Run linters (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running linters..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2"; \
	fi

## format: Format Go code
format:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@if command -v goimports > /dev/null; then \
		echo "Running goimports..."; \
		goimports -w .; \
	fi

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned build artifacts"

## install: Install the binary to GOBIN
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed $(BINARY_NAME)"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "Built Docker image: $(APP_NAME):$(VERSION)"

## docker-run: Run Docker container
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 3000:3000 --env-file .env $(APP_NAME):latest

## docker-compose: Run with docker-compose
docker-compose:
	@echo "Starting with docker-compose..."
	@docker-compose up --build

## docker-compose-down: Stop docker-compose
docker-compose-down:
	@echo "Stopping docker-compose..."
	@docker-compose down

## setup: Initial project setup
setup: deps
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from example..."; \
		cp .env.example .env 2>/dev/null || echo "No .env.example found"; \
	fi
	@if [ ! -f .air.toml ]; then \
		echo "Creating air configuration..."; \
		echo 'root = "."' > .air.toml; \
		echo 'testdata_dir = "testdata"' >> .air.toml; \
		echo 'tmp_dir = "tmp"' >> .air.toml; \
		echo '' >> .air.toml; \
		echo '[build]' >> .air.toml; \
		echo '  args_bin = []' >> .air.toml; \
		echo '  bin = "./tmp/main"' >> .air.toml; \
		echo '  cmd = "go build -o ./tmp/main ./cmd/api"' >> .air.toml; \
		echo '  delay = 1000' >> .air.toml; \
		echo '  exclude_dir = ["assets", "tmp", "vendor", "testdata", "bin"]' >> .air.toml; \
		echo '  exclude_file = []' >> .air.toml; \
		echo '  exclude_regex = ["_test.go"]' >> .air.toml; \
		echo '  exclude_unchanged = false' >> .air.toml; \
		echo '  follow_symlink = false' >> .air.toml; \
		echo '  full_bin = ""' >> .air.toml; \
		echo '  include_dir = []' >> .air.toml; \
		echo '  include_ext = ["go", "tpl", "tmpl", "html"]' >> .air.toml; \
		echo '  include_file = []' >> .air.toml; \
		echo '  kill_delay = "0s"' >> .air.toml; \
		echo '  log = "build-errors.log"' >> .air.toml; \
		echo '  poll = false' >> .air.toml; \
		echo '  poll_interval = 0' >> .air.toml; \
		echo '  rerun = false' >> .air.toml; \
		echo '  rerun_delay = 500' >> .air.toml; \
		echo '  send_interrupt = false' >> .air.toml; \
		echo '  stop_on_root = false' >> .air.toml; \
	fi
	@echo "Development environment ready!"

## health: Check application health
health:
	@echo "Checking application health..."
	@curl -f http://localhost:3000/api/health || echo "Application not running or unhealthy"

## size: Show binary size
size: build
	@echo "Binary size:"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME) | awk '{print $$5 "\t" $$9}'

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

## help: Show this help message
help:
	@echo "API Analytics Service - Available commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""