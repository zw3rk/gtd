# Claude GTD Specification

## Overview
`claude-gtd` is a SQLite-driven CLI task management tool implemented as a single static binary in Go. It stores tasks per-project in a `claude-tasks.db` file at the git repository root.

## Database Location
- The tool searches upward from the current directory to find the nearest `.git` directory
- Creates/uses `claude-tasks.db` in the git root directory
- Errors if not inside a git repository

## Database Schema

### tasks table
```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent INTEGER REFERENCES tasks(id),
    priority TEXT CHECK(priority IN ('high', 'medium', 'low')) DEFAULT 'medium',
    state TEXT CHECK(state IN ('NEW', 'IN_PROGRESS', 'DONE', 'CANCELLED')) DEFAULT 'NEW',
    kind TEXT CHECK(kind IN ('BUG', 'FEATURE', 'REGRESSION')) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    source TEXT,
    blocked_by INTEGER REFERENCES tasks(id),
    tags TEXT
);

CREATE INDEX idx_state_priority ON tasks(state, priority);
CREATE INDEX idx_parent ON tasks(parent);
```

## Commands

### Task Creation
```bash
claude-gtd add-bug [--priority|-p high/medium/low] [--source|-s "source-string"] [--tags|-t "tag1,tag2"] <<EOF
TITLE
BODY
EOF

claude-gtd add-feature [--priority|-p high/medium/low] [--source|-s "source-string"] [--tags|-t "tag1,tag2"] <<EOF
TITLE
BODY
EOF

claude-gtd add-regression [--priority|-p high/medium/low] [--source|-s "source-string"] [--tags|-t "tag1,tag2"] <<EOF
TITLE
BODY
EOF

# Add subtask
claude-gtd add-subtask PARENT_ID --kind bug/feature/regression [--priority|-p high/medium/low] <<EOF
TITLE
BODY
EOF
```

### State Management
```bash
claude-gtd in-progress ID
claude-gtd done ID         # errors if subtasks are not DONE/CANCELLED
claude-gtd cancel ID       # sets state to CANCELLED
claude-gtd block ID --by BLOCKING_ID
claude-gtd unblock ID
```

### Listing & Querying
```bash
# Default: shows top 20 tasks (IN_PROGRESS first, then NEW), sorted by priority
claude-gtd list [--oneline] [--all]

# List completed tasks, most recent first
claude-gtd list-done [--oneline]

# List cancelled tasks
claude-gtd list-cancelled [--oneline]

# Filter options
claude-gtd list --state NEW --priority high
claude-gtd list --kind bug --tag backend
claude-gtd list --blocked  # shows tasks with blocked_by set

# Show task details with subtasks
claude-gtd show ID

# Search in title and description
claude-gtd search "database migration"
```

### Reporting
```bash
# Summary statistics
claude-gtd summary

# Export data
claude-gtd export --format json > tasks.json
claude-gtd export --format csv > tasks.csv
claude-gtd export --format markdown > tasks.md
```

## Output Formats

### Standard list output
```
[42] HIGH 🔴 NEW       Fix database connection leak
     Source: db.go:127:15@abc123
     Tags: backend, urgent
     
[38] MED  🟡 IN_PROG  Implement user authentication
     Blocked by: #42
```

### Oneline format
```
[42] 🔴 NEW Fix database connection leak
[38] 🟡 IN_PROG Implement user authentication
```

## Implementation Notes
- Single binary with embedded SQLite
- No external dependencies beyond Go standard library + SQLite driver
- Database auto-created on first run at `{git_root}/claude-tasks.db`
- All timestamps in UTC
- Source format examples: `file:line:col@githash`, `GitHub:issue/123`, `Slack:C123/p456`
- Add `.gitignore` entry for `claude-tasks.db` unless user wants to track tasks in version control
