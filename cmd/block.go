package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newBlockCommand creates the block command
func newBlockCommand() *cobra.Command {
	var blockingTaskID string

	cmd := &cobra.Command{
		Use:   "block TASK_ID --by BLOCKING_TASK_ID",
		Short: "Mark a task as blocked by another task",
		Long: `Mark a task as blocked by another task.
This indicates that the task cannot proceed until the blocking task is completed.`,
		Example: `  claude-gtd block abc123 --by def456
  claude-gtd block 1a2b --by 3c4d`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get task ID (hash or hash prefix)
			taskID := args[0]

			// Validate blocking task ID was provided
			if blockingTaskID == "" {
				return fmt.Errorf("blocking task ID is required (use --by flag)")
			}

			// Get both tasks to show info
			task, err := repo.GetByID(taskID)
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}

			blockingTask, err := repo.GetByID(blockingTaskID)
			if err != nil {
				return fmt.Errorf("blocking task not found: %w", err)
			}

			// Validate not blocking by itself
			if task.ID == blockingTask.ID {
				return fmt.Errorf("cannot block a task by itself")
			}

			// Block the task
			if err := repo.Block(task.ID, blockingTask.ID); err != nil {
				return fmt.Errorf("failed to block task: %w", err)
			}

			// Output success message
			if _, err := fmt.Fprintf(cmd.OutOrStdout(),
				"Task %s is now blocked by task %s\n  %s\n  blocked by: %s\n",
				task.ShortHash(), blockingTask.ShortHash(), task.Title, blockingTask.Title); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&blockingTaskID, "by", "", "ID/hash of the task that is blocking this task")
	// MarkFlagRequired panics on error, so we can safely ignore the return value
	_ = cmd.MarkFlagRequired("by")

	return cmd
}

// newUnblockCommand creates the unblock command
func newUnblockCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unblock TASK_ID",
		Short: "Remove blocking status from a task",
		Long:  `Remove blocking status from a task, allowing it to proceed.`,
		Example: `  claude-gtd unblock abc123
  claude-gtd unblock 1a2b`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get task ID (hash or hash prefix)
			taskID := args[0]

			// Get the task to show info
			task, err := repo.GetByID(taskID)
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}

			// Check if it was blocked
			wasBlocked := task.IsBlocked()

			// Unblock the task
			if err := repo.Unblock(task.ID); err != nil {
				return fmt.Errorf("failed to unblock task: %w", err)
			}

			// Output success message
			if wasBlocked {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(),
					"Task %s is no longer blocked: %s\n",
					task.ShortHash(), task.Title); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(),
					"Task %s was not blocked: %s\n",
					task.ShortHash(), task.Title); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
