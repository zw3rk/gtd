# claude-gtd Makefile
# Self-documenting Makefile for the claude-gtd project

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME := claude-gtd

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet
GOFMT := $(GOCMD) fmt

# Build parameters
BUILD_DIR := .
LDFLAGS := -ldflags="-s -w"

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

.PHONY: help
help: ## Show this help message
	@echo "$(COLOR_BOLD)claude-gtd - SQLite-driven CLI task management tool$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET)"
	@echo "  make $(COLOR_GREEN)<target>$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Targets:$(COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: ## Build the binary
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-static
build-static: ## Build a static binary
	@echo "$(COLOR_BLUE)Building static binary...$(COLOR_RESET)"
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)✓ Static binary built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: run
run: ## Run the application
	$(GOCMD) run .

.PHONY: install
install: build ## Install the binary to $GOPATH/bin
	@echo "$(COLOR_BLUE)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "$(COLOR_GREEN)✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

##@ Testing

.PHONY: test
test: ## Run all tests
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run tests in short mode
	$(GOTEST) -short ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	@echo "$(COLOR_GREEN)✓ Coverage report saved to coverage.out$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Run 'make coverage-html' to view HTML report$(COLOR_RESET)"

.PHONY: coverage-html
coverage-html: test-coverage ## Generate HTML coverage report
	@echo "$(COLOR_BLUE)Generating HTML coverage report...$(COLOR_RESET)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ HTML report saved to coverage.html$(COLOR_RESET)"

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(COLOR_BLUE)Running tests with race detector...$(COLOR_RESET)"
	$(GOTEST) -race ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

##@ Code Quality

.PHONY: fmt
fmt: ## Format code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	$(GOFMT) ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	$(GOVET) ./...
	@echo "$(COLOR_GREEN)✓ No issues found$(COLOR_RESET)"

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
		echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Install with:$(COLOR_RESET)"; \
		echo "  nix run nixpkgs#golangci-lint"; \
	fi

.PHONY: check
check: fmt vet test ## Run all checks (format, vet, test)
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	$(GOMOD) download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	$(GOMOD) verify

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)"
	$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

.PHONY: clean-db
clean-db: ## Remove the database file
	@echo "$(COLOR_YELLOW)⚠ Removing claude-tasks.db...$(COLOR_RESET)"
	@rm -f claude-tasks.db
	@echo "$(COLOR_GREEN)✓ Database removed$(COLOR_RESET)"

##@ Development Helpers

.PHONY: todo
todo: ## Show TODO/FIXME comments in code
	@echo "$(COLOR_BOLD)TODO/FIXME items:$(COLOR_RESET)"
	@grep -rn "TODO\|FIXME" --include="*.go" . || echo "$(COLOR_GREEN)✓ No TODO/FIXME items found$(COLOR_RESET)"

.PHONY: size
size: build ## Show binary size
	@echo "$(COLOR_BOLD)Binary size:$(COLOR_RESET)"
	@ls -lh $(BINARY_NAME) | awk '{print $$5 " " $$9}'

.PHONY: loc
loc: ## Count lines of code
	@echo "$(COLOR_BOLD)Lines of code:$(COLOR_RESET)"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -n 1

##@ Git Hooks

.PHONY: pre-commit
pre-commit: check ## Run pre-commit checks
	@echo "$(COLOR_GREEN)✓ Pre-commit checks passed$(COLOR_RESET)"

.PHONY: setup-hooks
setup-hooks: ## Setup git hooks
	@echo "$(COLOR_BLUE)Setting up git hooks...$(COLOR_RESET)"
	@echo "#!/bin/sh\nmake pre-commit" > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(COLOR_GREEN)✓ Git hooks installed$(COLOR_RESET)"