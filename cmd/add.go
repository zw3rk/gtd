package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// Common flags for add commands
type addFlags struct {
	priority string
	source   string
	tags     string
}

// newAddBugCommand creates the add-bug command
func newAddBugCommand() *cobra.Command {
	var flags addFlags

	cmd := &cobra.Command{
		Use:   "add-bug",
		Short: "Add a new bug task",
		Long: `Add a new bug task by providing a title and description.
Input is read from stdin in Git-style format:
  TITLE
  
  DESCRIPTION (required, can be multiple lines)`,
		Example: `  claude-gtd add-bug <<EOF
Fix memory leak

Memory usage grows unbounded when processing large files.
Need to investigate the file processing loop.
EOF

  claude-gtd add-bug --priority high --source "app.go:42" <<EOF
Fix authentication bypass

Users can access admin panel without proper credentials.
This is a critical security vulnerability.
EOF`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addTask(cmd, models.KindBug, &flags)
		},
	}

	addCommonFlags(cmd, &flags)
	return cmd
}

// newAddFeatureCommand creates the add-feature command
func newAddFeatureCommand() *cobra.Command {
	var flags addFlags

	cmd := &cobra.Command{
		Use:   "add-feature",
		Short: "Add a new feature task",
		Long: `Add a new feature task by providing a title and description.
Input is read from stdin in Git-style format:
  TITLE
  
  DESCRIPTION (required, can be multiple lines)`,
		Example: `  claude-gtd add-feature <<EOF
Add dark mode

Implement a toggle for dark/light theme switching.
Should persist user preference across sessions.
EOF

  claude-gtd add-feature --priority medium --tags "ui,enhancement" <<EOF
Implement user preferences

Allow users to customize their dashboard layout.
Include options for widget placement and color schemes.
EOF`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addTask(cmd, models.KindFeature, &flags)
		},
	}

	addCommonFlags(cmd, &flags)
	return cmd
}

// newAddRegressionCommand creates the add-regression command
func newAddRegressionCommand() *cobra.Command {
	var flags addFlags

	cmd := &cobra.Command{
		Use:   "add-regression",
		Short: "Add a new regression task",
		Long: `Add a new regression task by providing a title and description.
Input is read from stdin in Git-style format:
  TITLE
  
  DESCRIPTION (required, can be multiple lines)`,
		Example: `  claude-gtd add-regression <<EOF
Login broken after update

Authentication fails with valid credentials after v2.1.0 update.
Users report "Invalid credentials" error despite correct password.
EOF

  claude-gtd add-regression --priority high --source "v2.1.0" <<EOF
Search functionality regression

Search results are no longer sorted by relevance.
This worked correctly in v2.0.5 but broke in v2.1.0.
EOF`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addTask(cmd, models.KindRegression, &flags)
		},
	}

	addCommonFlags(cmd, &flags)
	return cmd
}

// addCommonFlags adds common flags to add commands
func addCommonFlags(cmd *cobra.Command, flags *addFlags) {
	cmd.Flags().StringVarP(&flags.priority, "priority", "p", "medium",
		"Task priority (high, medium, low)")
	cmd.Flags().StringVarP(&flags.source, "source", "s", "",
		"Source reference (e.g., file:line, issue#, version)")
	cmd.Flags().StringVarP(&flags.tags, "tags", "t", "",
		"Comma-separated tags")
}

// addTask handles the common logic for adding tasks
func addTask(cmd *cobra.Command, kind string, flags *addFlags) error {
	// Read input
	title, description, err := readTaskInput(cmd.InOrStdin())
	if err != nil {
		return err
	}

	// Create task
	task := models.NewTask(kind, title, description)

	// Apply flags
	if flags.priority != "" {
		// Validate priority
		switch flags.priority {
		case models.PriorityHigh, models.PriorityMedium, models.PriorityLow:
			task.Priority = flags.priority
		default:
			return fmt.Errorf("invalid priority: %s (must be high, medium, or low)", flags.priority)
		}
	}

	task.Source = flags.source
	task.Tags = flags.tags

	// Save to database
	if err := repo.Create(task); err != nil {
		// Check if it's a validation error and provide helpful guidance
		if strings.Contains(err.Error(), "description is required") {
			return fmt.Errorf("failed to create task: %w\n\nTasks must include both a title and a description.\nUse Git-style format:\n  <title>\n  \n  <description>", err)
		}
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Output success message
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), formatTaskCreated(task.ID, kind)); err != nil {
		return err
	}

	return nil
}
