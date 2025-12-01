# Checkup Usage Checker Makefile
.PHONY: build run run-email run-users run-monitors run-database run-all run-help clean install deps help

# Variables
APP_NAME := checkup
BINARY_NAME := checkup
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"
BUILD_FLAGS := -trimpath $(LDFLAGS)

# Directories
SRC_DIR := ./cmd/checkup
BUILD_DIR := ./bin

# Default target
all: clean build

## build: Build the checkup binary
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## run: Run checkup with default settings
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

## run-email: Check email usage
run-email: build
	@echo "Checking email usage..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --email

## run-users: Check user statistics
run-users: build
	@echo "Checking user statistics..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --users

## run-monitors: Check monitor status
run-monitors: build
	@echo "Checking monitors..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --monitors

## run-database: Check database usage
run-database: build
	@echo "Checking database..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --database

## run-all: Run all checks
run-all: build
	@echo "Running all checks..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --email
	@echo ""
	@./$(BUILD_DIR)/$(BINARY_NAME) --users
	@echo ""
	@./$(BUILD_DIR)/$(BINARY_NAME) --monitors
	@echo ""
	@./$(BUILD_DIR)/$(BINARY_NAME) --database

## run-help: Show help message
run-help: build
	@./$(BUILD_DIR)/$(BINARY_NAME) --help

## install: Install binary to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed! You can now run '$(BINARY_NAME)' from anywhere"

## uninstall: Remove binary from /usr/local/bin
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalled"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Cleaned build artifacts"

## watch: Watch for changes and rebuild (requires entr)
watch:
	@if command -v entr > /dev/null; then \
		echo "Watching for changes... (Ctrl+C to stop)"; \
		find . -name '*.go' | entr -r make run; \
	else \
		echo "entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Linux)"; \
	fi

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

## help: Show this help message
help:
	@echo "Checkup Usage Checker - Available commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""