package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// newInProgressCommand creates the in-progress command
func newInProgressCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "in-progress TASK_ID",
		Short: "Mark a task as in progress",
		Long:  `Mark a task as in progress. This changes the task state to IN_PROGRESS.`,
		Example: `  claude-gtd in-progress 42
  claude-gtd in-progress 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateTaskState(cmd, args[0], models.StateInProgress)
		},
	}
}

// newDoneCommand creates the done command
func newDoneCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "done TASK_ID",
		Short: "Mark a task as done",
		Long: `Mark a task as done. This changes the task state to DONE.
Parent tasks can only be marked as done when all their subtasks are either DONE or CANCELLED.`,
		Example: `  claude-gtd done 42
  claude-gtd done 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateTaskState(cmd, args[0], models.StateDone)
		},
	}
}

// newCancelCommand creates the cancel command
func newCancelCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel TASK_ID",
		Short: "Cancel a task",
		Long:  `Cancel a task. This changes the task state to CANCELLED.`,
		Example: `  claude-gtd cancel 42
  claude-gtd cancel 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateTaskState(cmd, args[0], models.StateCancelled)
		},
	}
}

// updateTaskState is a helper function to update task state
func updateTaskState(cmd *cobra.Command, taskIDStr string, newState string) error {
	// Get the task first to show info
	task, err := repo.GetByID(taskIDStr)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	
	// Update state
	if err := repo.UpdateState(task.ID, newState); err != nil {
		return fmt.Errorf("failed to update task state: %w", err)
	}
	
	// Output success message
	stateVerb := getStateVerb(newState)
	fmt.Fprintf(cmd.OutOrStdout(), "Task %s marked as %s: %s\n", 
		task.ShortHash(), stateVerb, task.Title)
	
	return nil
}

// getStateVerb returns a human-friendly verb for the state
func getStateVerb(state string) string {
	switch state {
	case models.StateInProgress:
		return "in progress"
	case models.StateDone:
		return "done"
	case models.StateCancelled:
		return "cancelled"
	default:
		return state
	}
}