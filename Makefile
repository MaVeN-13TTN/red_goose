# Red-Goose Makefile

# Variables
BINARY_NAME=red-goose
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Determine the OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	PLATFORM=linux
endif
ifeq ($(UNAME_S),Darwin)
	PLATFORM=darwin
endif
ifeq ($(OS),Windows_NT)
	PLATFORM=windows
	BINARY_NAME=red-goose.exe
endif

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(PLATFORM)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/red-goose

# Cross-compile for multiple platforms
.PHONY: cross-compile
cross-compile:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 ./cmd/red-goose
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 ./cmd/red-goose
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe ./cmd/red-goose

# Install the application
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/red-goose

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Help
.PHONY: help
help:
	@echo "Red-Goose Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build           Build the application"
	@echo "  make cross-compile   Cross-compile for multiple platforms"
	@echo "  make install         Install the application"
	@echo "  make run             Run the application"
	@echo "  make clean           Clean build artifacts"
	@echo "  make test            Run tests"
	@echo "  make test-coverage   Run tests with coverage"
	@echo "  make fmt             Format code"
	@echo "  make lint            Lint code"
	@echo "  make help            Show this help message"