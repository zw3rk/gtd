# GTD - Getting Things Done CLI Tool

A SQLite-driven command-line task management tool that follows the Getting Things Done (GTD) methodology. Built as a single static binary in Go with no external dependencies.

## Features

- **INBOX-first workflow**: All tasks start in INBOX for review before becoming actionable
- **Hierarchical tasks**: Support for parent tasks and subtasks with automatic dependency tracking
- **Task states**: NEW, IN_PROGRESS, DONE, CANCELLED, with proper state transitions
- **Task types**: BUG, FEATURE, REGRESSION for better categorization
- **Priority levels**: High, medium, low priority support
- **Blocking relationships**: Mark tasks as blocked by other tasks
- **Rich metadata**: Tags, source references (GitHub issues, file locations, etc.)
- **Multiple output formats**: Standard, oneline, JSON, CSV, Markdown
- **Git integration**: Automatically stores database at git repository root
- **Search functionality**: Full-text search across task titles and descriptions

## Quick Start

### Installation

#### Using Nix (Recommended)

```bash
# Install directly from GitHub
nix profile install github:zw3rk/gtd

# Or run without installing
nix run github:zw3rk/gtd -- --help
```

#### Pre-built Binaries

Download pre-built binaries from the [releases page](../../releases) or from our [Hydra CI](https://ci.zw3rk.com).

#### Building from Source

```bash
git clone https://github.com/zw3rk/gtd.git
cd gtd
go build -o gtd
```

### Basic Usage

```bash
# Add a new task (starts in INBOX)
gtd add bug <<EOF
Fix memory leak in authentication

Memory usage grows over time in the auth service
EOF

# Review inbox items
gtd review

# Accept a task (move from INBOX to NEW)
gtd accept <task-id>

# Start working on a task
gtd in-progress <task-id>

# List current tasks
gtd list

# Mark task as done
gtd done <task-id>

# Search tasks
gtd search "memory leak"
```

## GTD Workflow

This tool implements the core GTD principles:

### 1. Capture
All new tasks go into the **INBOX** state first:
```bash
gtd add bug <<EOF
Fix login bug

Users can't log in with special characters
EOF

gtd add feature <<EOF
Dark mode

Implement theme switching
EOF
```

### 2. Clarify & Organize
Review inbox items and decide what to do with them:
```bash
# See what needs review
gtd review

# Accept important tasks (INBOX → NEW)
gtd accept abc1234

# Reject irrelevant tasks (INBOX → INVALID)
gtd reject def5678
```

### 3. Engage
Work on your committed tasks:
```bash
# See current workload
gtd list

# Start a task (NEW → IN_PROGRESS)
gtd in-progress abc1234

# Complete it (IN_PROGRESS → DONE)
gtd done abc1234
```

## Command Reference

### Task Creation
- `gtd add bug` - Add a bug report
- `gtd add feature` - Add a feature request
- `gtd add regression` - Add a regression report
- `gtd add-subtask <parent-id>` - Add a subtask to existing task

### Task Management
- `gtd review` - Review tasks in INBOX
- `gtd accept <task-id>` - Accept task from INBOX (→ NEW)
- `gtd reject <task-id>` - Reject task from INBOX (→ INVALID)
- `gtd in-progress <task-id>` - Start working on task (→ IN_PROGRESS)
- `gtd done <task-id>` - Mark task as completed (→ DONE)
- `gtd cancel <task-id>` - Cancel task (→ CANCELLED)
- `gtd reopen <task-id>` - Reopen cancelled task (CANCELLED → NEW)

### Task Organization
- `gtd block <task-id> --by <blocking-task-id>` - Mark task as blocked
- `gtd unblock <task-id>` - Remove blocking status

### Viewing Tasks
- `gtd list` - List active tasks (NEW, IN_PROGRESS)
- `gtd list --all` - List all tasks including INBOX/INVALID
- `gtd list-done` - List completed tasks
- `gtd list-cancelled` - List cancelled tasks
- `gtd show <task-id>` - Show detailed task information
- `gtd summary` - Show task statistics

### Search & Export
- `gtd search <query>` - Search tasks by title/description
- `gtd export --format json` - Export tasks to JSON/CSV/Markdown

## Task States & Transitions

```
📥 INBOX ─[accept]→ 🆕 NEW ─[in-progress]→ 🔄 IN_PROGRESS ─[done]→ ✅ DONE
    │                   │                         └─[cancel]→ ❌ CANCELLED
    │                   └─[cancel]→ ❌ CANCELLED
    └─[reject]→ ⛔ INVALID
```

## Database Storage

- Database file: `claude-tasks.db` at git repository root
- Pure SQLite - no external dependencies
- Automatic schema creation and migration
- Per-project task isolation

## Development

### Prerequisites

- Go 1.21+
- SQLite3 (for development/testing)
- Nix (optional, for reproducible builds)

### Building

```bash
# Standard build
make build

# Static build (Linux)
CGO_ENABLED=1 go build -ldflags="-s -w" -o gtd

# Run tests
make test

# Run linter
make lint
```

### Nix Development

```bash
# Enter development shell
nix develop

# Build with nix
nix build

# Run tests in nix
nix develop -c make test
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Run the linter (`make lint`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## Architecture

- **Single binary**: No external runtime dependencies
- **SQLite backend**: Embedded database for reliability
- **Git-aware**: Automatically finds repository root
- **Cross-platform**: Linux, macOS, Windows support
- **Static linking**: Linux binaries are fully static

### Project Structure

```
gtd/
├── cmd/                    # CLI commands and UI logic
├── internal/
│   ├── database/          # SQLite database layer
│   ├── git/              # Git repository integration
│   ├── models/           # Core domain models and repository
│   ├── output/           # Output formatting and abstraction
│   └── services/         # Business logic and services
├── vendor/               # Go module dependencies
├── flake.nix            # Nix build configuration
├── Makefile             # Development tasks
└── README.md            # This file
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Configuration

GTD can be configured through environment variables. See [Configuration Guide](CONFIGURATION.md) for details.

Common configurations:
```bash
export GTD_COLOR="false"          # Disable colors
export GTD_PAGE_SIZE="50"         # Show more tasks
export GTD_DEFAULT_PRIORITY="high" # Default to high priority
```

## Support

- 📖 [User Guide](USAGE.md) - Detailed usage examples
- ⚙️ [Configuration Guide](CONFIGURATION.md) - Environment variables and settings
- 📝 [Command Reference](COMMANDS.md) - Complete command documentation
- 🤖 [AI Agent Guide](LLM_AGENT_USAGE.md) - Guide for AI assistants
- 🐛 [Issues](../../issues) - Bug reports and feature requests

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI interface
- Uses [go-sqlite3](https://github.com/mattn/go-sqlite3) for database access
- Inspired by the Getting Things Done methodology by David Allen
- Nix build system with [flake-parts](https://github.com/hercules-ci/flake-parts)
