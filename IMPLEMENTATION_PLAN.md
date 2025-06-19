# Claude GTD Implementation Plan

## Overview
This document provides a concrete step-by-step plan for implementing the claude-gtd project following Test-Driven Development (TDD) principles. Each step includes testing requirements and clear deliverables.

## Phase 1: Project Setup and Infrastructure

### Step 1: Initialize Nix Development Environment
- [x] Create `flake.nix` with Go development environment
- [x] Add SQLite and Go build tools
- [x] Create `.envrc` with `use flake`
- [x] Add `.gitignore` with common Go patterns and `claude-tasks.db`
- [x] Test: Verify `direnv allow` loads environment correctly

### Step 2: Initialize Go Module
- [x] Run `go mod init github.com/zw3rk/claude-gtd`
- [x] Create basic `main.go` with hello world
- [x] Test: `go run .` prints hello world
- [x] Commit: "Initialize Go module with basic main function"

### Step 3: Setup Project Structure
- [x] Create directory structure:
  - `cmd/` - CLI command implementations
  - `internal/database/` - Database layer
  - `internal/models/` - Data models
  - `internal/git/` - Git repository detection
- [x] Create `Makefile` with build, test, lint targets
- [x] Test: `make build` creates binary
- [x] Commit: "Add project directory structure and Makefile"

## Phase 2: Core Infrastructure (TDD)

### Step 4: Git Repository Detection
- [x] Write tests for git root detection in `internal/git/git_test.go`
  - Test finding .git in current directory
  - Test finding .git in parent directories
  - Test error when no .git found
- [x] Implement `FindGitRoot()` function
- [x] Test: All git detection tests pass
- [x] Commit: "Add git repository root detection"

### Step 5: Database Connection Layer
- [x] Write tests for database initialization in `internal/database/db_test.go`
  - Test database creation
  - Test connection pooling
  - Test schema creation
- [x] Implement database connection with SQLite driver
- [x] Implement schema migration/creation
- [x] Test: Database tests pass, schema is created correctly
- [x] Commit: "Add database connection layer with schema creation"

### Step 6: Task Model and Repository
- [x] Write tests for Task model in `internal/models/task_test.go`
  - Test task creation with validation
  - Test state transitions
  - Test parent/child relationships
- [x] Implement Task struct with validation methods
- [x] Implement TaskRepository with CRUD operations
- [x] Test: Model validation and repository tests pass
- [x] Commit: "Add Task model and repository with validation"

## Phase 3: CLI Command Implementation (TDD)

### Step 7: CLI Framework Setup
- [x] Write tests for CLI command structure
- [x] Implement basic CLI with cobra or similar
- [x] Add command routing and help system
- [x] Test: CLI help and version commands work
- [x] Commit: "Add CLI framework with command routing"

### Step 8: Task Creation Commands
- [x] Write integration tests for add-bug, add-feature, add-regression
  - Test reading from stdin
  - Test flag parsing (priority, source, tags)
  - Test database persistence
- [x] Implement add-bug command
- [x] Implement add-feature command
- [x] Implement add-regression command
- [x] Test: All creation commands work with various inputs
- [x] Commit: "Add task creation commands (bug, feature, regression)"

### Step 9: Subtask Management
- [ ] Write tests for add-subtask command
  - Test parent validation
  - Test subtask creation
  - Test hierarchy queries
- [ ] Implement add-subtask command
- [ ] Add parent task validation
- [ ] Test: Subtask creation and hierarchy work correctly
- [ ] Commit: "Add subtask creation with parent validation"

### Step 10: State Management Commands
- [ ] Write tests for state transitions
  - Test in-progress command
  - Test done command with subtask validation
  - Test cancel command
- [ ] Implement in-progress command
- [ ] Implement done command with subtask checking
- [ ] Implement cancel command
- [ ] Test: State transitions work with proper validation
- [ ] Commit: "Add state management commands (in-progress, done, cancel)"

### Step 11: Blocking Relationships
- [ ] Write tests for blocking functionality
  - Test block command
  - Test unblock command
  - Test blocked task queries
- [ ] Implement block command
- [ ] Implement unblock command
- [ ] Add blocking validation to state transitions
- [ ] Test: Blocking relationships work correctly
- [ ] Commit: "Add task blocking and unblocking functionality"

## Phase 4: Query and Display Features (TDD)

### Step 12: Basic List Command
- [ ] Write tests for list command
  - Test default sorting (IN_PROGRESS first, then NEW)
  - Test priority sorting
  - Test pagination (top 20)
- [ ] Implement list command with formatting
- [ ] Add --oneline flag support
- [ ] Add --all flag for no pagination
- [ ] Test: List command displays tasks correctly
- [ ] Commit: "Add basic list command with sorting and formatting"

### Step 13: Filtered Listing
- [ ] Write tests for filtered queries
  - Test state filtering
  - Test priority filtering
  - Test kind filtering
  - Test tag filtering
  - Test blocked filtering
- [ ] Implement query builder with filters
- [ ] Add filter flags to list command
- [ ] Implement list-done and list-cancelled commands
- [ ] Test: All filtering options work correctly
- [ ] Commit: "Add filtered listing with multiple query options"

### Step 14: Task Details and Search
- [ ] Write tests for show command
  - Test task detail display
  - Test subtask tree display
  - Test error on invalid ID
- [ ] Write tests for search command
  - Test title search
  - Test description search
  - Test case-insensitive search
- [ ] Implement show command with subtask tree
- [ ] Implement search command
- [ ] Test: Show and search commands work correctly
- [ ] Commit: "Add show and search commands"

## Phase 5: Reporting and Export (TDD)

### Step 15: Summary Statistics
- [ ] Write tests for summary command
  - Test count by state
  - Test count by priority
  - Test count by kind
- [ ] Implement summary statistics calculation
- [ ] Create formatted summary output
- [ ] Test: Summary shows accurate statistics
- [ ] Commit: "Add summary statistics command"

### Step 16: Export Functionality
- [ ] Write tests for export formats
  - Test JSON export
  - Test CSV export
  - Test Markdown export
- [ ] Implement JSON exporter
- [ ] Implement CSV exporter
- [ ] Implement Markdown exporter
- [ ] Test: All export formats produce valid output
- [ ] Commit: "Add export functionality (JSON, CSV, Markdown)"

## Phase 6: Polish and Performance

### Step 17: Output Formatting
- [ ] Write tests for output formatting
  - Test emoji indicators
  - Test color output (if terminal supports)
  - Test proper alignment
- [ ] Implement rich output formatting
- [ ] Add source and tag display
- [ ] Test: Output matches specification examples
- [ ] Commit: "Add rich output formatting with emojis and alignment"

### Step 18: Performance Optimization
- [ ] Write performance benchmarks
- [ ] Add database indexes (already in schema)
- [ ] Optimize query performance
- [ ] Implement connection pooling
- [ ] Test: Benchmarks show acceptable performance
- [ ] Commit: "Optimize database queries and connection handling"

### Step 19: Error Handling and Validation
- [ ] Write tests for edge cases
  - Test circular parent relationships
  - Test invalid state transitions
  - Test database errors
- [ ] Implement comprehensive error handling
- [ ] Add helpful error messages
- [ ] Test: All edge cases handled gracefully
- [ ] Commit: "Add comprehensive error handling and validation"

### Step 20: Final Integration and Documentation
- [ ] Write end-to-end integration tests
- [ ] Create USAGE.md with example invocations
- [ ] Update CLAUDE.md with final architecture
- [ ] Build static binary and test deployment
- [ ] Test: Full workflow tests pass
- [ ] Commit: "Add integration tests and usage documentation"

## Testing Strategy

1. **Unit Tests**: Test individual functions and methods in isolation
2. **Integration Tests**: Test command execution with real database
3. **End-to-End Tests**: Test complete workflows from CLI to database
4. **Benchmarks**: Test performance of critical operations

## Success Criteria

- [ ] All tests pass with >80% coverage
- [ ] Static binary builds successfully
- [ ] All commands work as specified in PROJECT.md
- [ ] Performance is acceptable for 1000+ tasks
- [ ] Error messages are helpful and clear