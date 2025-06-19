package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// newReviewCommand creates the review command
func newReviewCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review tasks in INBOX state",
		Long: `Display all tasks currently in INBOX state for review.
		
Tasks in INBOX are items that need to be triaged and processed.
Use 'gtd accept <task-id>' to accept a task (move from INBOX to NEW).
Use 'gtd reject <task-id>' to reject a task (mark as INVALID).

Note: You should complete your current active tasks before reviewing INBOX items.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for active tasks first
			activeTasks, err := repo.List(models.ListOptions{
				ShowDone:      false,
				ShowCancelled: false,
				All:           false,
			})
			if err != nil {
				return fmt.Errorf("failed to check active tasks: %w", err)
			}

			if len(activeTasks) > 0 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Warning: You have %d active tasks. Consider completing them before reviewing INBOX.\n", len(activeTasks))
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "   Use 'gtd list' to see your active tasks.\n\n")
			}
			tasks, err := repo.ListByState(models.StateInbox)
			if err != nil {
				return fmt.Errorf("failed to list inbox tasks: %w", err)
			}

			if len(tasks) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No tasks in INBOX.")
				return nil
			}

			switch outputFormat {
			case "json":
				return exportJSON(cmd.OutOrStdout(), tasks)
			case "csv":
				return exportCSV(cmd.OutOrStdout(), tasks)
			case "markdown":
				return exportMarkdown(cmd.OutOrStdout(), tasks)
			default:
				formatTaskList(cmd.OutOrStdout(), tasks, outputFormat == "oneline")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format: json, csv, markdown, oneline")

	return cmd
}

// newAcceptCommand creates the accept command to move tasks from INBOX to NEW
func newAcceptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept <task-id>",
		Short: "Accept task from INBOX (move to NEW state)",
		Long:  `Accept a task from INBOX state by moving it to NEW state, indicating it has been reviewed and accepted for work.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Find the task
			task, err := repo.GetByID(taskID)
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}

			// Check current state
			if task.State != models.StateInbox {
				return fmt.Errorf("task %s is not in INBOX state (current: %s)", task.ID[:7], task.State)
			}

			// Update to NEW state
			if err := repo.UpdateState(task.ID, models.StateNew); err != nil {
				return fmt.Errorf("failed to update task state: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Task %s accepted (moved from INBOX to NEW)\n", task.ID[:7])
			return nil
		},
	}

	return cmd
}

// newRejectCommand creates the reject command to mark tasks as INVALID
func newRejectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject <task-id>",
		Short: "Reject task from INBOX (mark as INVALID)",
		Long:  `Reject a task from INBOX state by marking it as INVALID, indicating it should not be worked on.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Find the task
			task, err := repo.GetByID(taskID)
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}

			// Check if task can be marked invalid
			if task.State == models.StateDone {
				return fmt.Errorf("cannot mark completed task as invalid")
			}

			// Update to INVALID state
			if err := repo.UpdateState(task.ID, models.StateInvalid); err != nil {
				return fmt.Errorf("failed to update task state: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Task %s rejected (marked as INVALID)\n", task.ID[:7])
			return nil
		},
	}

	return cmd
}
