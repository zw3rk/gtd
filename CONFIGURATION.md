# GTD Configuration Guide

GTD supports configuration through environment variables, allowing you to customize behavior without command-line flags.

## Environment Variables

### Database Configuration

- **`GTD_DATABASE_NAME`** - Name of the database file (default: `claude-tasks.db`)
  ```bash
  export GTD_DATABASE_NAME="my-tasks.db"
  ```

- **`GTD_DATABASE_PATH`** - Full path to database file (overrides auto-detection)
  ```bash
  export GTD_DATABASE_PATH="/home/user/tasks/project.db"
  ```

### Output Configuration

- **`GTD_DEFAULT_FORMAT`** - Default output format: `json`, `csv`, `markdown`, `oneline`, or empty for standard
  ```bash
  export GTD_DEFAULT_FORMAT="oneline"
  ```

- **`GTD_COLOR`** - Enable/disable colored output (default: `true`)
  ```bash
  export GTD_COLOR="false"
  ```

- **`NO_COLOR`** - Standard environment variable to disable colors (any non-empty value)
  ```bash
  export NO_COLOR=1
  ```

- **`GTD_PAGE_SIZE`** - Default number of items to show in lists (default: `20`)
  ```bash
  export GTD_PAGE_SIZE="50"
  ```

### Behavior Configuration

- **`GTD_AUTO_REVIEW`** - Automatically show review after adding tasks (default: `false`)
  ```bash
  export GTD_AUTO_REVIEW="true"
  ```

- **`GTD_SHOW_WARNINGS`** - Show warnings about active tasks when reviewing (default: `true`)
  ```bash
  export GTD_SHOW_WARNINGS="false"
  ```

- **`GTD_CONFIRM_DONE`** - Require confirmation when marking parent tasks done (default: `false`)
  ```bash
  export GTD_CONFIRM_DONE="true"
  ```

- **`GTD_DEFAULT_PRIORITY`** - Default priority for new tasks: `high`, `medium`, `low` (default: `medium`)
  ```bash
  export GTD_DEFAULT_PRIORITY="high"
  ```

### Editor Configuration

- **`EDITOR`** or **`VISUAL`** - Default editor for multi-line input (default: `vi`)
  ```bash
  export EDITOR="nano"
  export VISUAL="code --wait"  # VISUAL takes precedence
  ```

## Configuration Examples

### Minimal Setup
```bash
# Just change the database name
export GTD_DATABASE_NAME="work-tasks.db"
```

### Developer Setup
```bash
# Compact output, no colors, more items
export GTD_DEFAULT_FORMAT="oneline"
export GTD_COLOR="false"
export GTD_PAGE_SIZE="50"
export GTD_DEFAULT_PRIORITY="high"
```

### Team Setup
```bash
# Shared database, confirmations enabled
export GTD_DATABASE_PATH="/shared/team-tasks.db"
export GTD_CONFIRM_DONE="true"
export GTD_SHOW_WARNINGS="true"
```

### CI/CD Setup
```bash
# No colors, JSON output for parsing
export NO_COLOR=1
export GTD_DEFAULT_FORMAT="json"
export GTD_SHOW_WARNINGS="false"
```

## Shell Configuration

Add to your shell configuration file:

### Bash (~/.bashrc)
```bash
# GTD Configuration
export GTD_COLOR="true"
export GTD_PAGE_SIZE="30"
export GTD_DEFAULT_PRIORITY="medium"
```

### Zsh (~/.zshrc)
```bash
# GTD Configuration
export GTD_COLOR="true"
export GTD_PAGE_SIZE="30"
export GTD_DEFAULT_PRIORITY="medium"

# Fix heredoc issue in ZSH
alias gtd='command gtd'
```

### Fish (~/.config/fish/config.fish)
```fish
# GTD Configuration
set -x GTD_COLOR true
set -x GTD_PAGE_SIZE 30
set -x GTD_DEFAULT_PRIORITY medium
```

## Per-Project Configuration

Use direnv to set project-specific configuration:

### .envrc
```bash
# Project-specific GTD settings
export GTD_DATABASE_NAME="project-tasks.db"
export GTD_DEFAULT_PRIORITY="high"
export GTD_AUTO_REVIEW="true"
```

## Configuration Precedence

1. Command-line flags (when available)
2. Environment variables
3. Default values

## Viewing Current Configuration

To see what configuration values are being used:

```bash
# Show all GTD environment variables
env | grep GTD_

# Show color-related variables
env | grep -E "(GTD_COLOR|NO_COLOR)"
```

## Tips

1. **Database Path**: If `GTD_DATABASE_PATH` is not set, GTD will look for the database at the git repository root
2. **Colors**: GTD respects the standard `NO_COLOR` environment variable
3. **Page Size**: Set a higher page size if you have many tasks and want to see more at once
4. **Auto Review**: Enable `GTD_AUTO_REVIEW` if you follow strict GTD methodology
5. **Warnings**: Disable `GTD_SHOW_WARNINGS` in scripts to avoid interactive prompts

## Future Enhancements

Configuration file support (`.gtdrc` or `gtd.toml`) is planned for future releases.