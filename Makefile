# Claude GTD Makefile
# Provides standard build, test, and lint targets

# Variables
BINARY_NAME := claude-gtd
GO := go
GOFLAGS := 
CGO_ENABLED := 1
LDFLAGS := -ldflags="-s -w"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .

# Build static binary
.PHONY: build-static
build-static:
	@echo "Building static $(BINARY_NAME)..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test ./... -v

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -cover ./... -v

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, using go vet instead" && $(GO) vet ./...)
	@which golangci-lint > /dev/null && golangci-lint run || true

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	$(GO) clean

# Install binary to GOPATH/bin
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install

# Run the program
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run .

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build         - Build the binary"
	@echo "  make build-static  - Build static binary"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install       - Install binary to GOPATH/bin"
	@echo "  make run           - Run the program"
	@echo "  make help          - Show this help message"