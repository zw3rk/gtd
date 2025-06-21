# Tasks Export

Total tasks: 8

| ID | Type | State | Priority | Title | Tags | Source | Parent | Blocked By |
|---|---|---|---|---|---|---|---|---|
| 1 | FEATURE | NEW | medium | Remove global state completely | - | - | - | - |
| 2 | BUG | NEW | medium | Standardize output formatting | - | - | - | - |
| 3 | FEATURE | NEW | medium | Add full-text search support | - | - | - | - |
| 4 | FEATURE | NEW | medium | Implement config file loading | - | - | - | - |
| 5 | FEATURE | NEW | medium | Add test coverage for missing components | - | - | - | - |
| 6 | FEATURE | DONE | medium | Add database performance indexes | - | - | - | - |
| 7 | FEATURE | DONE | medium | Add helpful error messages with suggestions | - | - | - | - |
| 8 | BUG | DONE | medium | ZSH heredoc artifacts in task descriptions | - | - | - | - |

## Task Details

### #ec1c011c25eaf187bfd9ed7854e6db559c162ad9: Remove global state completely

Complete the refactoring to remove global db and repo variables from
cmd/root.go. All commands should use the App struct for dependency
injection. This is marked as a TODO in the code.

- **Type:** Feature
- **State:** NEW ◆
- **Priority:** medium =
- **Created:** 2025-06-21 02:38:58
- **Updated:** 2025-06-21 02:39:09

### #d8c1769abfc813b3b5cb312e2ddf472af9c40239: Standardize output formatting

Some commands bypass the output formatter abstraction and use direct
fmt.Fprintf calls (e.g., block.go, export.go). All output should go
through the configured formatter for consistency.

- **Type:** Bug
- **State:** NEW ◆
- **Priority:** medium =
- **Created:** 2025-06-21 02:37:39
- **Updated:** 2025-06-21 02:39:09

### #c46fe5f4e6823b1ab83d651510be01f3e80204ba: Add full-text search support

Implement full-text search functionality for better performance on large
datasets. The current LIKE-based search with wildcards cannot use indexes
effectively. Consider using SQLite's FTS5 extension.

- **Type:** Feature
- **State:** NEW ◆
- **Priority:** medium =
- **Created:** 2025-06-21 02:37:23
- **Updated:** 2025-06-21 02:39:09

### #9f3ede7d59f780a3821db5aefcfaf81adf3c5d2d: Implement config file loading

Implement YAML/TOML configuration file support as noted in the TODO comment
in internal/config/config.go. This would allow users to set default
configurations without relying solely on environment variables.

- **Type:** Feature
- **State:** NEW ◆
- **Priority:** medium =
- **Created:** 2025-06-21 02:36:50
- **Updated:** 2025-06-21 02:39:09

### #4c5550f9970bf9fddce3a1e44886e085928871c6: Add test coverage for missing components

Add test files for commands and utilities that currently lack test coverage,
including: terminal.go, review.go, format.go, reopen.go, app.go, utils.go,
add_new.go, and errors/suggestions.go. This will improve code confidence
and help catch regressions.

- **Type:** Feature
- **State:** NEW ◆
- **Priority:** medium =
- **Created:** 2025-06-21 02:36:32
- **Updated:** 2025-06-21 02:39:09

### #3c47fa5ec55aec5f9658f6a4d2b19dc34fba4fb7: Add database performance indexes

Add indexes for commonly queried fields to improve performance:
composite index on (state, priority), indexes on parent, ID prefix,
kind/state, blocked_by, created, updated, and tags.

- **Type:** Feature
- **State:** DONE ✓
- **Priority:** medium =
- **Created:** 2025-06-21 02:41:27
- **Updated:** 2025-06-21 02:41:32

### #f836cfb4b1cea4bc951a4164700b75dab7d314e0: Add helpful error messages with suggestions

When tasks are not found, provide "did you mean?" suggestions using
Levenshtein distance to help users identify the correct task ID.

- **Type:** Feature
- **State:** DONE ✓
- **Priority:** medium =
- **Created:** 2025-06-21 02:39:53
- **Updated:** 2025-06-21 02:39:58

### #81b4006e62d54eafb9a8d8e3222ae1c12f45606a: ZSH heredoc artifacts in task descriptions

Tasks created using ZSH heredoc syntax had "EOF < /dev/null" appended
to their descriptions. This was due to ZSH's unique heredoc behavior.

- **Type:** Bug
- **State:** DONE ✓
- **Priority:** medium =
- **Created:** 2025-06-21 02:39:40
- **Updated:** 2025-06-21 02:39:45

