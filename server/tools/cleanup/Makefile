# Cleanup Utility Makefile
.PHONY: build run clean install uninstall test deps watch version help

# Variables
APP_NAME := cleanup
BINARY_NAME := cleanup
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
SRC_DIR := ./cmd/cleanup
BUILD_DIR := ./bin

# Default target
all: clean build

## build: Build the cleanup binary
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## run: Build and run cleanup
run: build
	@echo "Running $(APP_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

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
	@echo "Cleaned build artifacts"

## watch: Watch for changes and rebuild (requires entr)
watch:
	@if command -v entr > /dev/null; then \
		echo "Watching for changes... (Ctrl+C to stop)"; \
		find . -name '*.go' | entr -r make run; \
	else \
		echo "entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Linux)"; \
	fi

## dry-run: Build and show what would be cleaned (if your app supports --dry-run)
dry-run: build
	@echo "Running cleanup in dry-run mode..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --dry-run || ./$(BUILD_DIR)/$(BINARY_NAME)

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

## help: Show this help message
help:
	@echo "Cleanup Utility - Available commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""