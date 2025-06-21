package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// Common flags for add commands
type addTaskFlags struct {
	priority string
	source   string
	tags     string
}

// newAddCommand creates the add command with subcommands
func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new task",
		Long:  `Add a new task to the system. Use subcommands to specify the task type.`,
		Example: `  # Add a bug task
  gtd add bug <<EOF
  Fix memory leak
  
  Memory usage grows unbounded when processing large files.
  EOF

  # Add a high-priority feature
  gtd add feature --priority high <<EOF
  Add dark mode
  
  Implement a toggle for dark/light theme switching.
  EOF`,
	}

	// Add subcommands for each task type
	cmd.AddCommand(
		newAddTaskCommand("bug", models.KindBug),
		newAddTaskCommand("feature", models.KindFeature),
		newAddTaskCommand("regression", models.KindRegression),
	)

	return cmd
}

// newAddTaskCommand creates a subcommand for adding a specific task type
func newAddTaskCommand(cmdName, taskKind string) *cobra.Command {
	var flags addTaskFlags

	// Build command metadata
	shortDesc := fmt.Sprintf("Add a new %s task", cmdName)
	longDesc := fmt.Sprintf(`Add a new %s task by providing a title and description.
Input is read from stdin in Git-style format:
  TITLE
  
  DESCRIPTION (required, can be multiple lines)`, cmdName)

	// Build examples based on task type
	var example string
	switch taskKind {
	case models.KindBug:
		example = `  gtd add bug <<EOF
Fix memory leak

Memory usage grows unbounded when processing large files.
Need to investigate the file processing loop.
EOF

  gtd add bug --priority high --source "app.go:42" <<EOF
Fix authentication bypass

Users can access admin panel without proper credentials.
This is a critical security vulnerability.
EOF`
	case models.KindFeature:
		example = `  gtd add feature <<EOF
Add dark mode

Implement a toggle for dark/light theme switching.
Should persist user preference across sessions.
EOF

  gtd add feature --priority medium --tags "ui,enhancement" <<EOF
Implement user preferences

Allow users to customize their dashboard layout.
Include options for widget placement and color schemes.
EOF`
	case models.KindRegression:
		example = `  gtd add regression <<EOF
Login broken after update

Authentication fails with valid credentials after v2.1.0 update.
Users report "Invalid credentials" error despite correct password.
EOF

  gtd add regression --priority high --source "v2.1.0" <<EOF
Search functionality regression

Search results are no longer sorted by relevance.
This worked correctly in v2.0.5 but broke in v2.1.0.
EOF`
	}

	cmd := &cobra.Command{
		Use:     cmdName,
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addTaskWithKind(cmd, taskKind, &flags)
		},
	}

	// Add common flags
	cmd.Flags().StringVarP(&flags.priority, "priority", "p", "medium",
		"Task priority (high, medium, low)")
	cmd.Flags().StringVarP(&flags.source, "source", "s", "",
		"Source reference (e.g., file:line, issue#, version)")
	cmd.Flags().StringVarP(&flags.tags, "tags", "t", "",
		"Comma-separated tags")

	return cmd
}

// addTaskWithKind handles the common logic for adding tasks
func addTaskWithKind(cmd *cobra.Command, kind string, flags *addTaskFlags) error {
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