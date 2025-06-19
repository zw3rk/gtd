package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zw3rk/claude-gtd/internal/models"
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
Input is read from stdin in the format:
  TITLE
  DESCRIPTION (optional, can be multiple lines)`,
		Example: `  echo "Fix auth module leak" | claude-gtd add-subtask 42 --kind bug
  claude-gtd add-subtask 10 --kind feature --priority high <<EOF
Implement dark mode toggle
Add a switch in settings to toggle between light and dark themes
EOF`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse parent ID
			parentID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid parent ID: %s", args[0])
			}
			
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
			task.Parent = &parentID
			
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
				"Created %s subtask #%d for task #%d (%s)\n", 
				flags.kind, task.ID, parent.ID, parent.Title)
			
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