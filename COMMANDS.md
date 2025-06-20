# GTD Command Reference

This document provides a comprehensive reference for all GTD commands.

## Task Creation Commands

### `gtd add`
Creates new tasks that start in INBOX state for review.

**Subcommands:**
- `gtd add bug` - Add a bug report
- `gtd add feature` - Add a feature request  
- `gtd add regression` - Add a regression report

**Usage:**
```bash
gtd add <type> [flags] <<EOF
Task Title

Task description (required, can be multiple lines)
EOF
```

**Flags:**
- `-p, --priority` - Task priority (high, medium, low) [default: medium]
- `-s, --source` - Source reference (e.g., file:line, issue#, version)
- `-t, --tags` - Comma-separated tags

**Examples:**
```bash
# Add a high-priority bug
gtd add bug --priority high --source "auth.go:42" <<EOF
Fix authentication bypass

Users can access admin panel without proper credentials.
This is a critical security vulnerability.
EOF

# Add a feature with tags
gtd add feature --tags "ui,enhancement" <<EOF
Implement dark mode

Add a toggle for dark/light theme switching.
Should persist user preference across sessions.
EOF
```

### `gtd add-subtask`
Adds a subtask to an existing task.

**Usage:**
```bash
gtd add-subtask <parent-id> --kind <type> [flags] <<EOF
Subtask Title

Subtask description
EOF
```

**Required Flags:**
- `--kind` - Task type (bug, feature, regression)

**Optional Flags:**
- `-p, --priority` - Task priority (high, medium, low) [default: medium]

## Task Review Commands

### `gtd review`
Shows all tasks in INBOX state that need to be triaged.

**Usage:**
```bash
gtd review [flags]
```

**Flags:**
- `-o, --output` - Output format (json, csv, markdown, oneline)

### `gtd accept`
Accepts a task from INBOX, moving it to NEW state.

**Usage:**
```bash
gtd accept <task-id>
```

### `gtd reject`
Rejects a task from INBOX, marking it as INVALID.

**Usage:**
```bash
gtd reject <task-id>
```

## State Management Commands

### `gtd in-progress`
Starts work on a task (NEW → IN_PROGRESS).

**Usage:**
```bash
gtd in-progress <task-id>
```

### `gtd done`
Marks a task as completed (IN_PROGRESS → DONE).

**Usage:**
```bash
gtd done <task-id>
```

**Note:** Parent tasks can only be marked done when all subtasks are DONE or CANCELLED.

### `gtd cancel`
Cancels a task (→ CANCELLED).

**Usage:**
```bash
gtd cancel <task-id>
```

### `gtd reopen`
Reopens a cancelled task (CANCELLED → NEW).

**Usage:**
```bash
gtd reopen <task-id>
```

## Task Organization Commands

### `gtd block`
Marks a task as blocked by another task.

**Usage:**
```bash
gtd block <task-id> --by <blocking-task-id>
```

**Required Flags:**
- `--by` - ID of the task that is blocking

### `gtd unblock`
Removes blocking status from a task.

**Usage:**
```bash
gtd unblock <task-id>
```

## Viewing Commands

### `gtd list`
Lists tasks with various filtering options.

**Usage:**
```bash
gtd list [flags]
```

**Flags:**
- `--oneline` - Show tasks in compact format
- `--all` - Show all tasks including DONE and CANCELLED
- `--state` - Filter by state (NEW, IN_PROGRESS, DONE, CANCELLED)
- `--priority` - Filter by priority (high, medium, low)
- `--kind` - Filter by kind (bug, feature, regression)
- `--tag` - Filter by tag
- `--blocked` - Show only blocked tasks
- `--limit` - Maximum number of tasks to show [default: 20]

**Examples:**
```bash
# List high-priority bugs
gtd list --priority high --kind bug

# Show all tasks in oneline format
gtd list --all --oneline

# Show blocked tasks
gtd list --blocked
```

### `gtd list-done`
Lists completed tasks.

**Usage:**
```bash
gtd list-done [flags]
```

**Flags:**
- `--oneline` - Show tasks in compact format

### `gtd list-cancelled`
Lists cancelled tasks.

**Usage:**
```bash
gtd list-cancelled [flags]
```

**Flags:**
- `--oneline` - Show tasks in compact format

### `gtd show`
Shows detailed information about a specific task.

**Usage:**
```bash
gtd show <task-id>
```

### `gtd summary`
Shows task statistics and summary.

**Usage:**
```bash
gtd summary [flags]
```

**Flags:**
- `--active` - Show only active task counts

## Search and Export Commands

### `gtd search`
Searches tasks by title and description.

**Usage:**
```bash
gtd search <query> [flags]
```

**Flags:**
- `-o, --output` - Output format (json, csv, markdown, oneline)

### `gtd export`
Exports tasks to different formats.

**Usage:**
```bash
gtd export [flags]
```

**Flags:**
- `-f, --format` - Output format (json, csv, markdown) [required]
- `--all` - Include all tasks (default excludes DONE/CANCELLED)
- `--state` - Filter by state
- `--priority` - Filter by priority
- `--kind` - Filter by kind

## Task ID Format

Task IDs are SHA-1 hashes (40 characters) that uniquely identify each task. You can use:
- Full hash: `abc123def456...` (40 chars)
- Short hash: `abc123d` (7+ chars, like git)
- Prefix: Any unique prefix of 4+ characters

## State Transitions

```
INBOX ─[accept]→ NEW ─[in-progress]→ IN_PROGRESS ─[done]→ DONE
  │                │                      └─[cancel]→ CANCELLED
  │                └─[cancel]→ CANCELLED ←─[reopen]┘
  └─[reject]→ INVALID
```

## Output Formats

- **Standard**: Git-style format with full details
- **Oneline**: Compact single-line format
- **JSON**: Machine-readable JSON format
- **CSV**: Comma-separated values for spreadsheets
- **Markdown**: Formatted for documentation

## Tips

1. **Task IDs**: Use tab completion or copy from list output
2. **Bulk Operations**: Use shell scripting for bulk updates
3. **Filtering**: Combine multiple filters for precise queries
4. **Workflow**: Follow INBOX → NEW → IN_PROGRESS → DONE flow