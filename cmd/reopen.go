package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// newReopenCommand creates the reopen command to move tasks from CANCELLED to NEW
func newReopenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reopen <task-id>",
		Short: "Reopen a cancelled task (move to NEW state)",
		Long: `Reopen a cancelled task by moving it back to NEW state.
		
This command allows you to resume work on tasks that were previously cancelled.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Find the task
			task, err := repo.GetByID(taskID)
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}

			// Check current state
			if task.State != models.StateCancelled {
				return fmt.Errorf("task %s is not in CANCELLED state (current: %s)", task.ShortHash(), task.State)
			}

			// Update to NEW state
			if err := repo.UpdateState(task.ID, models.StateNew); err != nil {
				return fmt.Errorf("failed to update task state: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Task %s reopened (moved from CANCELLED to NEW)\n", task.ShortHash())
			return nil
		},
	}

	return cmd
}