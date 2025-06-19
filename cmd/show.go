package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zw3rk/claude-gtd/internal/models"
)

// newShowCommand creates the show command
func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show TASK_ID",
		Short: "Show task details",
		Long:  `Show detailed information about a task, including description, metadata, and subtasks.`,
		Example: `  claude-gtd show 42
  claude-gtd show 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get task ID (hash or hash prefix)
			taskID := args[0]
			
			// Get the task
			task, err := repo.GetByID(taskID)
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}
			
			// Get parent if this is a subtask
			var parent *models.Task
			if task.Parent != nil {
				parent, _ = repo.GetByID(*task.Parent)
			}
			
			// Get subtasks
			subtasks, err := repo.GetChildren(task.ID)
			if err != nil {
				return fmt.Errorf("failed to get subtasks: %w", err)
			}
			
			// Format and output
			formatTaskDetails(cmd.OutOrStdout(), task, parent, subtasks)
			
			return nil
		},
	}
}

// formatTaskDetails formats detailed task information
func formatTaskDetails(w io.Writer, task *models.Task, parent *models.Task, subtasks []*models.Task) {
	// Calculate subtask stats
	var stats *SubtaskStats
	if len(subtasks) > 0 {
		stats = &SubtaskStats{Total: len(subtasks)}
		for _, st := range subtasks {
			if st.State == models.StateDone {
				stats.Done++
			}
		}
	}
	
	// Use git-style format
	fmt.Fprint(w, formatTaskGitStyle(task, stats))
	
	// Parent info if this is a subtask
	if parent != nil {
		fmt.Fprintf(w, "\nParent: %s - %s\n", parent.ShortHash(), parent.Title)
	}
	
	// Subtasks
	if len(subtasks) > 0 {
		fmt.Fprintln(w, "\nSubtasks:")
		fmt.Fprintln(w, strings.Repeat("-", 30))
		
		for _, subtask := range subtasks {
			// Use the subtask format with metadata on the right
			subtaskLine := formatSubtask(subtask)
			fmt.Fprintf(w, "  %s\n", subtaskLine)
			
			if subtask.Description != "" {
				// Show first line of description, indented
				lines := strings.Split(subtask.Description, "\n")
				fmt.Fprintf(w, "      %s\n", lines[0])
			}
		}
		
		// Summary
		fmt.Fprintf(w, "\n%s\n", formatSubtaskSummary(subtasks))
	}
}

// formatSubtaskSummary creates a summary of subtask states
func formatSubtaskSummary(subtasks []*models.Task) string {
	counts := make(map[string]int)
	for _, task := range subtasks {
		counts[task.State]++
	}
	
	var parts []string
	if n := counts[models.StateDone]; n > 0 {
		parts = append(parts, fmt.Sprintf("%d done", n))
	}
	if n := counts[models.StateInProgress]; n > 0 {
		parts = append(parts, fmt.Sprintf("%d in progress", n))
	}
	if n := counts[models.StateNew]; n > 0 {
		parts = append(parts, fmt.Sprintf("%d new", n))
	}
	if n := counts[models.StateCancelled]; n > 0 {
		parts = append(parts, fmt.Sprintf("%d cancelled", n))
	}
	
	total := len(subtasks)
	return fmt.Sprintf("Total: %d subtasks (%s)", total, strings.Join(parts, ", "))
}