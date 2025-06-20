package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all configuration values for the application
type Config struct {
	// Database configuration
	DatabaseName string
	DatabasePath string // Full path, empty means auto-detect

	// Output configuration
	DefaultFormat string // json, csv, markdown, oneline, or empty for standard
	ColorEnabled  bool
	PageSize      int // Default number of items to show in lists

	// Behavior configuration
	AutoReview      bool // Automatically show review after adding tasks
	ShowWarnings    bool // Show warnings about active tasks when reviewing
	ConfirmDone     bool // Require confirmation when marking parent tasks done
	DefaultPriority string

	// Git configuration
	GitRoot string // Detected git root, empty if not in git repo

	// Environment
	Editor string // Default editor for multi-line input
}

// NewConfig creates a new configuration with defaults
func NewConfig() *Config {
	return &Config{
		DatabaseName:    "claude-tasks.db",
		DefaultFormat:   "",
		ColorEnabled:    true,
		PageSize:        20,
		AutoReview:      false,
		ShowWarnings:    true,
		ConfirmDone:     false,
		DefaultPriority: "medium",
		Editor:          "vi",
	}
}

// Load loads configuration from environment variables
func (c *Config) Load() error {
	// Database configuration
	if dbName := os.Getenv("GTD_DATABASE_NAME"); dbName != "" {
		c.DatabaseName = dbName
	}
	if dbPath := os.Getenv("GTD_DATABASE_PATH"); dbPath != "" {
		c.DatabasePath = dbPath
	}

	// Output configuration
	if format := os.Getenv("GTD_DEFAULT_FORMAT"); format != "" {
		format = strings.ToLower(format)
		switch format {
		case "json", "csv", "markdown", "oneline", "standard", "":
			c.DefaultFormat = format
		default:
			return fmt.Errorf("invalid GTD_DEFAULT_FORMAT: %s", format)
		}
	}

	if colorStr := os.Getenv("GTD_COLOR"); colorStr != "" {
		color, err := strconv.ParseBool(colorStr)
		if err != nil {
			return fmt.Errorf("invalid GTD_COLOR value: %s", colorStr)
		}
		c.ColorEnabled = color
	} else if noColor := os.Getenv("NO_COLOR"); noColor != "" {
		// Support standard NO_COLOR env var
		c.ColorEnabled = false
	}

	if pageSizeStr := os.Getenv("GTD_PAGE_SIZE"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 {
			return fmt.Errorf("invalid GTD_PAGE_SIZE: %s", pageSizeStr)
		}
		c.PageSize = pageSize
	}

	// Behavior configuration
	if autoReview := os.Getenv("GTD_AUTO_REVIEW"); autoReview != "" {
		review, err := strconv.ParseBool(autoReview)
		if err != nil {
			return fmt.Errorf("invalid GTD_AUTO_REVIEW value: %s", autoReview)
		}
		c.AutoReview = review
	}

	if showWarnings := os.Getenv("GTD_SHOW_WARNINGS"); showWarnings != "" {
		warnings, err := strconv.ParseBool(showWarnings)
		if err != nil {
			return fmt.Errorf("invalid GTD_SHOW_WARNINGS value: %s", showWarnings)
		}
		c.ShowWarnings = warnings
	}

	if confirmDone := os.Getenv("GTD_CONFIRM_DONE"); confirmDone != "" {
		confirm, err := strconv.ParseBool(confirmDone)
		if err != nil {
			return fmt.Errorf("invalid GTD_CONFIRM_DONE value: %s", confirmDone)
		}
		c.ConfirmDone = confirm
	}

	if priority := os.Getenv("GTD_DEFAULT_PRIORITY"); priority != "" {
		priority = strings.ToLower(priority)
		switch priority {
		case "high", "medium", "low":
			c.DefaultPriority = priority
		default:
			return fmt.Errorf("invalid GTD_DEFAULT_PRIORITY: %s", priority)
		}
	}

	// Editor configuration
	if editor := os.Getenv("EDITOR"); editor != "" {
		c.Editor = editor
	}
	if visual := os.Getenv("VISUAL"); visual != "" {
		c.Editor = visual // VISUAL takes precedence over EDITOR
	}

	return nil
}

// LoadFromFile loads configuration from a file (future enhancement)
func (c *Config) LoadFromFile(path string) error {
	// TODO: Implement config file loading (YAML/TOML)
	// For now, we only support environment variables
	return nil
}

// GetDatabasePath returns the full path to the database
func (c *Config) GetDatabasePath() string {
	if c.DatabasePath != "" {
		return c.DatabasePath
	}
	if c.GitRoot != "" {
		return filepath.Join(c.GitRoot, c.DatabaseName)
	}
	// Fallback to current directory
	return c.DatabaseName
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate priority
	switch c.DefaultPriority {
	case "high", "medium", "low":
		// valid
	default:
		return fmt.Errorf("invalid default priority: %s", c.DefaultPriority)
	}

	// Validate format if set
	if c.DefaultFormat != "" {
		switch c.DefaultFormat {
		case "json", "csv", "markdown", "oneline":
			// valid
		default:
			return fmt.Errorf("invalid default format: %s", c.DefaultFormat)
		}
	}

	// Validate page size
	if c.PageSize < 1 {
		return fmt.Errorf("page size must be at least 1")
	}

	return nil
}

// String returns a string representation of the config for debugging
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("GTD Configuration:\n")
	sb.WriteString(fmt.Sprintf("  Database: %s\n", c.GetDatabasePath()))
	sb.WriteString(fmt.Sprintf("  Default Format: %s\n", c.DefaultFormat))
	sb.WriteString(fmt.Sprintf("  Color Enabled: %v\n", c.ColorEnabled))
	sb.WriteString(fmt.Sprintf("  Page Size: %d\n", c.PageSize))
	sb.WriteString(fmt.Sprintf("  Auto Review: %v\n", c.AutoReview))
	sb.WriteString(fmt.Sprintf("  Show Warnings: %v\n", c.ShowWarnings))
	sb.WriteString(fmt.Sprintf("  Confirm Done: %v\n", c.ConfirmDone))
	sb.WriteString(fmt.Sprintf("  Default Priority: %s\n", c.DefaultPriority))
	sb.WriteString(fmt.Sprintf("  Editor: %s\n", c.Editor))
	return sb.String()
}