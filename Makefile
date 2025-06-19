# gtd Makefile
# Self-documenting Makefile for the gtd project

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME := gtd

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
	@printf "$(COLOR_BOLD)gtd - SQLite-driven CLI task management tool$(COLOR_RESET)\n"
	@printf "\n"
	@printf "$(COLOR_BOLD)Usage:$(COLOR_RESET)\n"
	@printf "  make $(COLOR_GREEN)<target>$(COLOR_RESET)\n"
	@printf "\n"
	@printf "$(COLOR_BOLD)Targets:$(COLOR_RESET)\n"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: ## Build the binary
	@printf "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)\n"
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@printf "$(COLOR_GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)\n"

.PHONY: build-static
build-static: ## Build a static binary
	@printf "$(COLOR_BLUE)Building static binary...$(COLOR_RESET)\n"
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@printf "$(COLOR_GREEN)✓ Static binary built: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)\n"

.PHONY: run
run: ## Run the application
	$(GOCMD) run .

.PHONY: install
install: build ## Install the binary to $GOPATH/bin
	@printf "$(COLOR_BLUE)Installing $(BINARY_NAME)...$(COLOR_RESET)\n"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@printf "$(COLOR_GREEN)✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)\n"

##@ Testing

.PHONY: test
test: ## Run all tests
	@printf "$(COLOR_BLUE)Running tests...$(COLOR_RESET)\n"
	$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run tests in short mode
	$(GOTEST) -short ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@printf "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)\n"
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	@printf "$(COLOR_GREEN)✓ Coverage report saved to coverage.out$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)Run 'make coverage-html' to view HTML report$(COLOR_RESET)\n"

.PHONY: coverage-html
coverage-html: test-coverage ## Generate HTML coverage report
	@printf "$(COLOR_BLUE)Generating HTML coverage report...$(COLOR_RESET)\n"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@printf "$(COLOR_GREEN)✓ HTML report saved to coverage.html$(COLOR_RESET)\n"

.PHONY: test-race
test-race: ## Run tests with race detector
	@printf "$(COLOR_BLUE)Running tests with race detector...$(COLOR_RESET)\n"
	$(GOTEST) -race ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

##@ Code Quality

.PHONY: fmt
fmt: ## Format code
	@printf "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)\n"
	$(GOFMT) ./...
	@printf "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)\n"

.PHONY: vet
vet: ## Run go vet
	@printf "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)\n"
	$(GOVET) ./...
	@printf "$(COLOR_GREEN)✓ No issues found$(COLOR_RESET)\n"

.PHONY: lint
lint: ## Run golangci-lint
	@printf "$(COLOR_BLUE)Running linter...$(COLOR_RESET)\n"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
		printf "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_YELLOW)⚠ golangci-lint not installed. Install with:$(COLOR_RESET)\n"; \
		printf "  nix run nixpkgs#golangci-lint\n"; \
	fi

.PHONY: check
check: fmt vet test ## Run all checks (format, vet, test)
	@printf "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)\n"

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	@printf "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)\n"
	$(GOMOD) download
	@printf "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)\n"

.PHONY: deps-update
deps-update: ## Update dependencies
	@printf "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)\n"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@printf "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)\n"

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	$(GOMOD) verify

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	@printf "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)\n"
	$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@printf "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)\n"

.PHONY: clean-db
clean-db: ## Remove the database file
	@printf "$(COLOR_YELLOW)⚠ Removing claude-tasks.db...$(COLOR_RESET)\n"
	@rm -f claude-tasks.db
	@printf "$(COLOR_GREEN)✓ Database removed$(COLOR_RESET)\n"

##@ Development Helpers

.PHONY: todo
todo: ## Show TODO/FIXME comments in code
	@printf "$(COLOR_BOLD)TODO/FIXME items:$(COLOR_RESET)\n"
	@grep -rn "TODO\|FIXME" --include="*.go" . || printf "$(COLOR_GREEN)✓ No TODO/FIXME items found$(COLOR_RESET)\n"

.PHONY: size
size: build ## Show binary size
	@printf "$(COLOR_BOLD)Binary size:$(COLOR_RESET)\n"
	@ls -lh $(BINARY_NAME) | awk '{print $$5 " " $$9}'

.PHONY: loc
loc: ## Count lines of code
	@printf "$(COLOR_BOLD)Lines of code:$(COLOR_RESET)\n"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -n 1

##@ Git Hooks

.PHONY: pre-commit
pre-commit: check ## Run pre-commit checks
	@printf "$(COLOR_GREEN)✓ Pre-commit checks passed$(COLOR_RESET)\n"

.PHONY: setup-hooks
setup-hooks: ## Setup git hooks
	@printf "$(COLOR_BLUE)Setting up git hooks...$(COLOR_RESET)\n"
	@echo "#!/bin/sh\nmake pre-commit" > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@printf "$(COLOR_GREEN)✓ Git hooks installed$(COLOR_RESET)\n"