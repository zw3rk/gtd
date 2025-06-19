package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// newBlockCommand creates the block command
func newBlockCommand() *cobra.Command {
	var blockingTaskID int
	
	cmd := &cobra.Command{
		Use:   "block TASK_ID --by BLOCKING_TASK_ID",
		Short: "Mark a task as blocked by another task",
		Long: `Mark a task as blocked by another task.
This indicates that the task cannot proceed until the blocking task is completed.`,
		Example: `  claude-gtd block 42 --by 10
  claude-gtd block 5 --by 3`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse task ID
			taskID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", args[0])
			}
			
			// Validate blocking task ID was provided
			if blockingTaskID == 0 {
				return fmt.Errorf("blocking task ID is required (use --by flag)")
			}
			
			// Validate not blocking by itself
			if taskID == blockingTaskID {
				return fmt.Errorf("cannot block a task by itself")
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
			
			// Block the task
			if err := repo.Block(taskID, blockingTaskID); err != nil {
				return fmt.Errorf("failed to block task: %w", err)
			}
			
			// Output success message
			fmt.Fprintf(cmd.OutOrStdout(), 
				"Task #%d is now blocked by task #%d\n  %s\n  blocked by: %s\n", 
				task.ID, blockingTask.ID, task.Title, blockingTask.Title)
			
			return nil
		},
	}
	
	cmd.Flags().IntVar(&blockingTaskID, "by", 0, "ID of the task that is blocking this task")
	cmd.MarkFlagRequired("by")
	
	return cmd
}

// newUnblockCommand creates the unblock command
func newUnblockCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unblock TASK_ID",
		Short: "Remove blocking status from a task",
		Long:  `Remove blocking status from a task, allowing it to proceed.`,
		Example: `  claude-gtd unblock 42
  claude-gtd unblock 5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse task ID
			taskID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", args[0])
			}
			
			// Get the task to show info
			task, err := repo.GetByID(taskID)
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}
			
			// Check if it was blocked
			wasBlocked := task.IsBlocked()
			
			// Unblock the task
			if err := repo.Unblock(taskID); err != nil {
				return fmt.Errorf("failed to unblock task: %w", err)
			}
			
			// Output success message
			if wasBlocked {
				fmt.Fprintf(cmd.OutOrStdout(), 
					"Task #%d is no longer blocked: %s\n", 
					task.ID, task.Title)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), 
					"Task #%d was not blocked: %s\n", 
					task.ID, task.Title)
			}
			
			return nil
		},
	}
}