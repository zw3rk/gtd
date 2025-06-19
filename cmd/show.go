package cmd

import (
	"fmt"
	"io"
	"strconv"
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
			// Parse task ID
			taskID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", args[0])
			}
			
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
			subtasks, err := repo.GetChildren(taskID)
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
	// Header with ID and title
	fmt.Fprintf(w, "Task #%d: %s\n", task.ID, task.Title)
	fmt.Fprintln(w, strings.Repeat("=", 50))
	
	// Condensed metadata line
	var metadata []string
	
	// Add core info
	metadata = append(metadata, fmt.Sprintf("%s %s", formatKind(task.Kind), getPriorityEmoji(task.Priority)))
	metadata = append(metadata, fmt.Sprintf("%s %s", task.State, getStateEmoji(task.State)))
	metadata = append(metadata, strings.ToUpper(task.Priority))
	metadata = append(metadata, fmt.Sprintf("Created: %s", task.Created.Format("2006-01-02")))
	
	// Add optional fields
	if task.Tags != "" {
		metadata = append(metadata, fmt.Sprintf("Tags: %s", task.Tags))
	}
	
	if task.Source != "" {
		metadata = append(metadata, fmt.Sprintf("Source: %s", task.Source))
	}
	
	if task.IsBlocked() && task.BlockedBy != nil {
		metadata = append(metadata, fmt.Sprintf("Blocked by: #%d", *task.BlockedBy))
	}
	
	// Print the condensed line
	fmt.Fprintf(w, "[ %s ]\n", strings.Join(metadata, " | "))
	
	// Parent info if this is a subtask
	if parent != nil {
		fmt.Fprintf(w, "\nParent: #%d - %s\n", parent.ID, parent.Title)
	}
	
	// Description
	if task.Description != "" {
		fmt.Fprintln(w, "\nDescription:")
		fmt.Fprintln(w, strings.Repeat("-", 30))
		fmt.Fprintln(w, task.Description)
	}
	
	// Subtasks
	if len(subtasks) > 0 {
		fmt.Fprintln(w, "\nSubtasks:")
		fmt.Fprintln(w, strings.Repeat("-", 30))
		
		for _, subtask := range subtasks {
			emoji := getStateEmoji(subtask.State)
			priorityEmoji := getPriorityEmoji(subtask.Priority)
			
			fmt.Fprintf(w, "  [%d] %s %s %s - %s\n", 
				subtask.ID, priorityEmoji, emoji, subtask.State, subtask.Title)
			
			if subtask.Description != "" {
				// Show first line of description
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