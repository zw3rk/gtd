# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `claude-gtd`, a SQLite-driven CLI task management tool implemented as a single static binary in Go. The project follows GTD (Getting Things Done) methodology and stores tasks per-project in a `claude-tasks.db` file at the git repository root.

## Development Environment

### Nix and Flakes
- Use `flake.nix` in the git root for environment provisioning
- Use `direnv` with `.envrc` configured to `use flake` for automatic environment loading

## Development Commands

### Build
```bash
# Build the binary
go build -o claude-gtd

# Build static binary
CGO_ENABLED=1 go build -ldflags="-s -w" -o claude-gtd
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
   - Task states: NEW, IN_PROGRESS, DONE, CANCELLED
   - Task types: BUG, FEATURE, REGRESSION
   - Priority levels: high, medium, low
   - Blocking relationships between tasks

2. **CLI Interface**: Commands organized by function:
   - Task creation: `add-bug`, `add-feature`, `add-regression`, `add-subtask`
   - State management: `in-progress`, `done`, `cancel`, `block`, `unblock`
   - Querying: `list`, `list-done`, `list-cancelled`, `show`, `search`
   - Reporting: `summary`, `export`

3. **Git Integration**: Automatically finds git repository root and stores database there

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
   - Parent tasks can only be marked DONE when all subtasks are DONE or CANCELLED
   - Blocked tasks should be visually indicated in output

3. **Input Handling**: Multi-line input (title + body) is read from stdin using heredoc syntax

4. **Source Format**: Flexible format supporting various references like `file:line:col@githash`, `GitHub:issue/123`, `Slack:C123/p456`

## Code Quality Guidelines

- Create frequent meaningful, high quality commit messsages.