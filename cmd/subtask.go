package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// newAddSubtaskCommand creates the add-subtask command
func newAddSubtaskCommand() *cobra.Command {
	var flags struct {
		kind     string
		priority string
	}

	cmd := &cobra.Command{
		Use:   "add-subtask PARENT_ID --kind bug|feature|regression [flags]",
		Short: "Add a subtask to an existing task",
		Long: `Add a subtask to an existing task by providing the parent task ID and kind.
Input is read from stdin in Git-style format:
  TITLE
  
  DESCRIPTION (required, can be multiple lines)`,
		Example: `  claude-gtd add-subtask abc123 --kind bug <<EOF
Fix auth module leak

Memory leak in the authentication module causing OOM errors.
Need to properly dispose of session objects.
EOF

  claude-gtd add-subtask 1a2b --kind feature --priority high <<EOF
Implement dark mode toggle

Add a switch in settings to toggle between light and dark themes.
Should update all UI components to respect the theme setting.
EOF`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get parent ID (hash or hash prefix)
			parentID := args[0]

			// Validate kind is provided
			if flags.kind == "" {
				return fmt.Errorf("kind is required (use --kind flag)")
			}

			// Validate and normalize kind value
			var normalizedKind string
			switch flags.kind {
			case "bug", "BUG":
				normalizedKind = models.KindBug
			case "feature", "FEATURE":
				normalizedKind = models.KindFeature
			case "regression", "REGRESSION":
				normalizedKind = models.KindRegression
			default:
				return fmt.Errorf("invalid kind: %s (must be bug, feature, or regression)", flags.kind)
			}

			// Check parent exists
			parent, err := repo.GetByID(parentID)
			if err != nil {
				return fmt.Errorf("parent task not found: %w", err)
			}

			// Read input
			title, description, err := readTaskInput(cmd.InOrStdin())
			if err != nil {
				return err
			}

			// Create subtask
			task := models.NewTask(normalizedKind, title, description)
			task.Parent = &parent.ID

			// Apply priority if specified
			if flags.priority != "" {
				switch flags.priority {
				case models.PriorityHigh, models.PriorityMedium, models.PriorityLow:
					task.Priority = flags.priority
				default:
					return fmt.Errorf("invalid priority: %s (must be high, medium, or low)", flags.priority)
				}
			}

			// Save to database
			if err := repo.Create(task); err != nil {
				return fmt.Errorf("failed to create subtask: %w", err)
			}

			// Output success message
			fmt.Fprintf(cmd.OutOrStdout(),
				"Created %s subtask %s for task %s (%s)\n",
				flags.kind, task.ShortHash(), parent.ShortHash(), parent.Title)

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.kind, "kind", "",
		"Task kind (bug, feature, regression) [required]")
	cmd.MarkFlagRequired("kind")

	cmd.Flags().StringVarP(&flags.priority, "priority", "p", "medium",
		"Task priority (high, medium, low)")

	return cmd
}
