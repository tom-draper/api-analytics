# Monitor Service Makefile
.PHONY: build run clean install uninstall test deps watch daemon status stop version help

# Variables
APP_NAME := monitor
BINARY_NAME := monitor
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
SRC_DIR := ./cmd/monitor
BUILD_DIR := ./bin

# Default target
all: clean build

## build: Build the monitor binary
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## run: Build and run monitor
run: build
	@echo "Running $(APP_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

## daemon: Run monitor in background
daemon: build
	@echo "Starting $(APP_NAME) in background..."
	@nohup ./$(BUILD_DIR)/$(BINARY_NAME) > $(BUILD_DIR)/monitor.log 2>&1 &
	@echo $$! > $(BUILD_DIR)/monitor.pid
	@echo "Monitor started with PID $$(cat $(BUILD_DIR)/monitor.pid)"
	@echo "Logs: $(BUILD_DIR)/monitor.log"

## status: Check if monitor is running
status:
	@if [ -f $(BUILD_DIR)/monitor.pid ]; then \
		PID=$$(cat $(BUILD_DIR)/monitor.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Monitor is running (PID: $$PID)"; \
		else \
			echo "Monitor is not running (stale PID file found)"; \
			rm -f $(BUILD_DIR)/monitor.pid; \
		fi \
	else \
		echo "Monitor is not running"; \
	fi

## stop: Stop background monitor
stop:
	@if [ -f $(BUILD_DIR)/monitor.pid ]; then \
		PID=$$(cat $(BUILD_DIR)/monitor.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Stopping monitor (PID: $$PID)..."; \
			kill $$PID; \
			rm -f $(BUILD_DIR)/monitor.pid; \
			echo "Monitor stopped"; \
		else \
			echo "Monitor is not running (stale PID file removed)"; \
			rm -f $(BUILD_DIR)/monitor.pid; \
		fi \
	else \
		echo "Monitor is not running"; \
	fi

## restart: Restart background monitor
restart: stop daemon

## logs: Tail monitor logs
logs:
	@if [ -f $(BUILD_DIR)/monitor.log ]; then \
		tail -f $(BUILD_DIR)/monitor.log; \
	else \
		echo "No log file found at $(BUILD_DIR)/monitor.log"; \
	fi

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

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

## help: Show this help message
help:
	@echo "Monitor Service - Available commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""