.PHONY: build run clean test deps fmt lint

# Variables
BINARY_NAME=docker-image-checker
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/checker

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

# Run in daemon mode
daemon: build
	@echo "Running $(BINARY_NAME) in daemon mode..."
	@$(BINARY_PATH) --daemon

# Run once
once: build
	@echo "Running $(BINARY_NAME) once..."
	@$(BINARY_PATH) --once

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Install the application
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_PATH) /usr/local/bin/

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"

# Docker build (for containerized deployment)
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME) .

# Help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  daemon        - Run in daemon mode"
	@echo "  once          - Run once and exit"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  install       - Install binary to /usr/local/bin"
	@echo "  dev-setup     - Setup development environment"
	@echo "  docker-build  - Build Docker image"
	@echo "  help          - Show this help"
