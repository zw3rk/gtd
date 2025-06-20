# gtd Usage Guide

A SQLite-driven CLI task management tool following GTD (Getting Things Done) methodology.

## Quick Start

```bash
# Add a bug (goes to INBOX for review)
gtd add bug <<EOF
Fix login validation

Users can bypass login with empty password
EOF

# Add a feature (also goes to INBOX)
gtd add feature --priority high <<EOF
Add dark mode toggle

Implement user preference for dark/light theme
EOF

# Complete current active work first
gtd list  # Check your current commitments

# Review tasks in INBOX (only when active tasks are manageable)
gtd review

# Accept task from INBOX after review (moves to NEW)
gtd accept 1a2b3c4

# Start working on accepted tasks
gtd in-progress 1a2b3c4

# Complete tasks
gtd done 1a2b3c4
```

## Installation

```bash
# Build from source
make build

# Install to $GOPATH/bin
make install

# Or run directly
go run .
```

## Core Commands

### Task Creation

#### Add a Bug
```bash
# Basic bug (goes to INBOX for review)
gtd add bug <<EOF
Memory leak in parser

Parser doesn't free allocated memory after processing
EOF
gtd add bug <<EOF
Fix memory leak

Memory is not freed after processing large files
EOF

# With priority and tags
echo "Critical security issue
SQL injection in user search
EOF

# With source reference
gtd add bug --source "sentry:12345" <<EOF
Null pointer exception

Crash when user profile is missing data
EOF
```

#### Add a Feature
```bash
# Simple feature
gtd add feature <<EOF
Add user preferences

Allow users to customize their experience
EOF

# With full details
echo "Implement OAuth2 login
Support GitHub and Google providers
EOF
```

#### Add a Regression
```bash
# Regression from specific commit
gtd add regression --source "git:abc123" <<EOF
Search broken after refactor

Search returns no results after commit abc123
EOF
```

#### Add a Subtask
```bash
# Add subtask to existing task
echo "Write unit tests" | claude-gtd add-subtask 5 --kind bug --priority high
```

### INBOX and Review Workflow

```bash
# First, ensure current work is manageable
gtd list  # Check your active tasks

# Review all tasks in INBOX (tool warns if you have too many active tasks)
gtd review

# Accept task from INBOX (move to NEW state for future work)
gtd accept 1a2b3c4

# Reject task from INBOX (mark as INVALID)
gtd reject 1a2b3c4

# Export INBOX tasks for review
gtd review --output json
```

### State Management

```bash
# Start working on a task
gtd in-progress 3

# Complete a task
gtd done 3

# Cancel a task
gtd cancel 3

# Reopen a cancelled task (moves back to NEW)
gtd reopen 3
```

### Task Blocking

```bash
# Block task 5 by task 3 (task 5 can't proceed until task 3 is done)
claude-gtd block 5 --by 3

# Remove blocking relationship
claude-gtd unblock 5
```

### Listing and Filtering

#### Basic Listing
```bash
# List active tasks (default)
claude-gtd list

# List with one-line format
claude-gtd list --oneline

# List all tasks including done/cancelled
claude-gtd list --all

# List only completed tasks
claude-gtd list-done

# List only cancelled tasks
claude-gtd list-cancelled
```

#### Filtering
```bash
# Filter by state
claude-gtd list --state in_progress

# Filter by priority
claude-gtd list --priority high

# Filter by type
claude-gtd list --kind bug

# Filter by tag
claude-gtd list --tag security

# Combine filters
claude-gtd list --priority high --kind bug --tag critical
```

### Task Details

```bash
# Show full details of a task
claude-gtd show 7

# Shows:
# - Full description
# - Subtasks tree
# - Blocking relationships
# - All metadata
```

### Search

```bash
# Search in titles and descriptions
claude-gtd search "memory leak"

# Case-insensitive partial matching
claude-gtd search database

# Search with compact output
claude-gtd search --oneline connection
```

### Summary and Statistics

```bash
# Show full summary
claude-gtd summary

# Show only active tasks summary
claude-gtd summary --active
```

### Export

```bash
# Export to JSON (default)
claude-gtd export

# Export to CSV
claude-gtd export --format csv

# Export to Markdown
claude-gtd export --format markdown

# Export to file
claude-gtd export --format json --output tasks.json

# Export with filters
claude-gtd export --format csv --state done --kind bug

# Export only active tasks
claude-gtd export --active
```

## Input Format

Most commands that create tasks read from stdin. The tool supports Git-style commit message format:

```bash
# Single line (title only)
gtd add bug <<EOF
Task title

Task description
EOF

# Git-style format (title, blank line, body)
gtd add bug <<EOF
Fix critical security issue

Detailed description of the security vulnerability
and how to reproduce it.
EOF

# Multi-line with heredoc
gtd add feature <<EOF
Implement user preferences

Allow users to customize their experience with:
- Theme selection
- Notification settings
- Default filters
EOF

# Legacy format (title + immediate description, no blank line)
gtd add bug <<EOF
Quick fix

Just a small typo
EOF
```

## Task States

- **INBOX**: All new tasks start here and require review (default for add-* commands)
- **NEW**: Tasks that have been reviewed and accepted, ready to work on
- **IN_PROGRESS**: Currently being worked on
- **DONE**: Completed task
- **CANCELLED**: Cancelled task
- **INVALID**: Rejected task that should not be worked on

**Workflow:** INBOX → (review) → NEW/INVALID → IN_PROGRESS → DONE/CANCELLED

**Key Principle:** Everything starts in INBOX. Review decides what becomes actionable work (NEW) or gets rejected (INVALID).

## Priority Levels

- **high**: ! Critical tasks
- **medium**: = Normal priority (default)
- **low**: - Low priority

## Task Types

- **BUG**: Defects and issues
- **FEATURE**: New functionality
- **REGRESSION**: Previously working features that broke

## Output Formats

### Standard Format
```
[1] ! IN_PROGRESS Add user authentication
    Bug • Created: 2024-01-15 • Tags: security, auth
    
[2] = NEW Implement dark mode
    Feature • Created: 2024-01-16
```

### Oneline Format
```
[1] ! IN_PROGRESS Add user authentication
[2] = NEW Implement dark mode
```

### Export Formats
- **JSON**: Full task data with all fields
- **CSV**: Tabular format for spreadsheets
- **Markdown**: Human-readable with detailed descriptions

## Special Indicators

- ⊘ Task is blocked by another task
- ◇ INBOX state (requires review)
- ◆ NEW state
- ▶ IN_PROGRESS state
- ✓ DONE state
- ✗ CANCELLED state
- ✘ INVALID state

## Examples

### Complete Workflow
```bash
# 1. Add a feature with subtasks
gtd add feature --priority high --tags "ui,dashboard" <<EOF
Implement user dashboard

Create a customizable dashboard for users
EOF

# 2. Add subtasks
echo "Design dashboard layout" | claude-gtd add-subtask 1 --kind feature
echo "Implement API endpoints" | claude-gtd add-subtask 1 --kind feature
echo "Write tests" | claude-gtd add-subtask 1 --kind feature

# 3. Start working
claude-gtd in-progress 2  # Start with design

# 4. Check progress
claude-gtd show 1  # See parent with all subtasks

# 5. Complete subtasks
claude-gtd done 2
claude-gtd done 3
claude-gtd done 4

# 6. Complete parent
claude-gtd done 1
```

### INBOX Review Workflow
```bash
# Complete current work first
gtd list  # Ensure your active tasks are manageable

# Check what tasks need review (tool will warn if you have too many active tasks)
gtd review

# Review a specific task in detail
gtd show 1a2b3c4

# Accept task (move to NEW, ready for future work)
gtd accept 1a2b3c4

# Reject task (mark as INVALID)
gtd reject 5e6f7g8

# Export INBOX for team review
gtd review --output markdown
```

### Bug Triage Workflow
```bash
# List all high-priority bugs
gtd list --kind bug --priority high

# Search for specific issues
gtd search "memory leak"

# Block a feature by a bug
gtd block 10 --by 5  # Feature 10 blocked by bug 5

# Get summary of bug status
gtd list --kind bug --oneline
```

### Reporting
```bash
# Weekly summary
claude-gtd summary

# Export completed tasks for review
claude-gtd export --format markdown --state done --output completed-tasks.md

# Export all tasks for backup
claude-gtd export --format json --all --output backup.json
```

## Tips

1. **Use tags consistently**: Create a tagging convention (e.g., `security`, `performance`, `ui`)
2. **Set realistic priorities**: Reserve "high" for truly critical tasks
3. **Break down large tasks**: Use subtasks for better tracking
4. **Regular reviews**: Use `summary` and `list` commands to stay on top of tasks
5. **Document sources**: Use `--source` to track where tasks originated

## Database Location

The tool automatically finds your git repository root and stores the database there as `claude-tasks.db`. This ensures each project has its own task database.