# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `gtd`, a SQLite-driven CLI task management tool implemented as a single static binary in Go. The project follows GTD (Getting Things Done) methodology and stores tasks per-project in a `claude-tasks.db` file at the git repository root.

## Development Environment

### Nix and Flakes
- Use `flake.nix` in the git root for environment provisioning
- Use `direnv` with `.envrc` configured to `use flake` for automatic environment loading

## Development Commands

### Build
```bash
# Build the binary
go build -o gtd

# Build static binary
CGO_ENABLED=1 go build -ldflags="-s -w" -o gtd
```

### Test
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestName ./...
```

### Lint
```bash
# Install golangci-lint if needed
nix run nixpkgs#golangci-lint -- run

# Or use go vet
go vet ./...
```

## Architecture

### Key Components

1. **Database Layer**: SQLite embedded database with tasks table storing:
   - Task hierarchy (parent/subtask relationships)
   - Task states: INBOX, NEW, IN_PROGRESS, DONE, CANCELLED, INVALID
   - Task types: BUG, FEATURE, REGRESSION
   - Priority levels: high, medium, low
   - Blocking relationships between tasks

2. **Service Layer**: Business logic separated from data access:
   - `TaskService` interface for all task operations
   - Centralized state transition validation
   - Clean separation between commands and business logic

3. **CLI Interface**: Commands organized by function:
   - Task creation: `add bug`, `add feature`, `add regression`, `add-subtask`
   - Review workflow: `review`, `accept`, `reject`
   - State management: `in-progress`, `done`, `cancel`, `reopen`, `block`, `unblock`
   - Querying: `list`, `list-done`, `list-cancelled`, `show`, `search`
   - Reporting: `summary`, `export`

4. **Output Layer**: Consistent formatting across commands:
   - `Formatter` abstraction for different output formats
   - Shared formatting functions in `internal/output`
   - Support for colored and plain text output

5. **Git Integration**: Automatically finds git repository root and stores database there

### Database Schema

The core `tasks` table includes:
- Hierarchical structure via `parent` field
- State tracking with validation constraints
- Blocking relationships via `blocked_by` field
- Metadata: tags, source references, timestamps

### Design Principles

- Single static binary with no external dependencies
- Database auto-created on first run
- All operations relative to git repository root
- Support for multiple output formats (standard, oneline, JSON, CSV, Markdown)
- Parent tasks cannot be marked DONE if subtasks are incomplete

## Important Implementation Details

1. **Git Root Detection**: The tool must search upward from current directory to find `.git` and use that as the base for `claude-tasks.db`

2. **Task State Rules**: 
   - All new tasks start in INBOX state
   - Tasks must be accepted (INBOX → NEW) before work can begin
   - Parent tasks can only be marked DONE when all subtasks are DONE or CANCELLED
   - Blocked tasks are visually indicated with [BLOCKED] or ⊘ symbol
   - State transitions provide helpful error messages guiding users to the correct command

3. **Input Handling**: 
   - Multi-line input (title + body) is read from stdin using heredoc syntax
   - Automatic detection and removal of ZSH heredoc artifacts ("EOF < /dev/null")
   - Git-style format: title, blank line, then description

4. **Source Format**: Flexible format supporting various references like `file:line:col@githash`, `GitHub:issue/123`, `Slack:C123/p456`

## Recent Architecture Improvements

1. **Dependency Injection**: Removed global state in favor of App struct
2. **Service Layer**: Added TaskService interface for business logic
3. **Command Consolidation**: Combined add-bug/feature/regression into `add` with subcommands
4. **Output Abstraction**: Created formatter package for consistent output
5. **Better Error Messages**: State transition errors now include helpful guidance

## Code Quality Guidelines

- Create frequent meaningful, high quality commit messages
- Use dependency injection instead of global variables
- Separate business logic from presentation logic
- Write tests for new functionality
- Keep formatting logic in the output package

## Memory Guidelines

- Always mark all completed tasks as completed in the relevant .md files.