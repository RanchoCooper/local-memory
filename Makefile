# LocalMemory Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Binary names
CLI_BINARY=localmemory
SERVER_BINARY=localmemory-server
MCP_BINARY=localmemory-mcp

# Build directories
BUILD_DIR=./build
DATA_DIR=./data

# Python
PYTHON=python3
PIP=pip3
PYTHON_DIR=./python

# Docker
DOCKER_IMAGE=localmemory/localmemory
DOCKER_TAG=latest

.PHONY: all build build-cli build-server build-mcp test test-unit test-integration test-coverage clean fmt lint install run run-server run-mcp docker-build docker-up docker-down python-deps help

# Default target
all: fmt test build

## build: Build all binaries
build: build-cli build-server build-mcp

## build-cli: Build CLI binary
build-cli:
	$(GOBUILD) -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/cli

## build-server: Build HTTP server binary
build-server:
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_BINARY) ./cmd/server

## build-mcp: Build MCP server binary
build-mcp:
	$(GOBUILD) -o $(BUILD_DIR)/$(MCP_BINARY) ./cmd/mcp

## test: Run all tests
test: test-unit test-integration

## test-unit: Run unit tests
test-unit:
	$(GOTEST) -v -short ./tests/unit/...

## test-integration: Run integration tests
test-integration:
	$(GOTEST) -v ./tests/integration/...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./tests/unit/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## fmt: Format code
fmt:
	$(GOFMT) ./...

## lint: Run linter (requires golangci-lint)
lint:
	golangci-lint run

## install: Install dependencies
install:
	$(GOMOD) download
	$(GOMOD) tidy

## run: Run CLI
run: build-cli
	$(BUILD_DIR)/$(CLI_BINARY) $(ARGS)

## run-server: Run HTTP server
run-server: build-server
	$(BUILD_DIR)/$(SERVER_BINARY)

## run-mcp: Run MCP server (for Claude Code integration)
run-mcp: build-mcp
	$(BUILD_DIR)/$(MCP_BINARY)

## python-deps: Install Python dependencies
python-deps:
	$(PIP) install -r $(PYTHON_DIR)/requirements.txt

## docker-build: Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## docker-up: Start services with docker-compose
docker-up:
	docker-compose up -d

## docker-down: Stop services with docker-compose
docker-down:
	docker-compose down

## help: Show this help message
help:
	@echo "LocalMemory Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
