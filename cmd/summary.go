package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// newSummaryCommand creates the summary command
func newSummaryCommand() *cobra.Command {
	var activeOnly bool

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show task summary statistics",
		Long:  `Display a summary of all tasks, showing counts by state, type, and priority.`,
		Example: `  claude-gtd summary
  claude-gtd summary --active`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get all tasks
			opts := models.ListOptions{
				All:           true,
				ShowDone:      true,
				ShowCancelled: true,
				State:         "", // Include all states
			}

			tasks, err := repo.List(opts)
			if err != nil {
				return fmt.Errorf("failed to get tasks: %w", err)
			}

			// Generate and display summary
			formatSummary(cmd.OutOrStdout(), tasks, activeOnly)

			return nil
		},
	}

	cmd.Flags().BoolVar(&activeOnly, "active", false, "Show only active tasks (exclude DONE and CANCELLED)")

	return cmd
}

// formatSummary formats and displays task statistics
func formatSummary(w io.Writer, tasks []*models.Task, activeOnly bool) {
	// Initialize counters
	stateCounts := make(map[string]int)
	typeCounts := make(map[string]int)
	priorityCounts := make(map[string]int)
	blockedCount := 0
	parentCount := 0
	subtaskCount := 0

	// Count tasks
	activeTasks := 0
	for _, task := range tasks {
		// Skip done/cancelled if activeOnly
		if activeOnly && (task.State == models.StateDone || task.State == models.StateCancelled) {
			continue
		}

		stateCounts[task.State]++
		typeCounts[formatKind(task.Kind)]++
		priorityCounts[task.Priority]++

		if task.IsBlocked() {
			blockedCount++
		}

		// Count parents and subtasks
		hasChildren := false
		for _, other := range tasks {
			if other.Parent != nil && *other.Parent == task.ID {
				hasChildren = true
				break
			}
		}
		if hasChildren {
			parentCount++
		}
		if task.Parent != nil {
			subtaskCount++
		}

		if task.State == models.StateNew || task.State == models.StateInProgress {
			activeTasks++
		}
	}

	// Calculate total
	total := 0
	for _, count := range stateCounts {
		total += count
	}

	// Display summary
	if activeOnly {
		_, _ = fmt.Fprintf(w, "Active Tasks: %d\n", activeTasks)
	} else {
		_, _ = fmt.Fprintf(w, "Task Summary\n")
		_, _ = fmt.Fprintln(w, strings.Repeat("=", 50))
		_, _ = fmt.Fprintf(w, "Total Tasks: %d\n", total)
	}
	_, _ = fmt.Fprintln(w)

	// By State
	_, _ = fmt.Fprintln(w, "By State:")
	if !activeOnly {
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "INBOX:", stateCounts[models.StateInbox])
	}
	_, _ = fmt.Fprintf(w, "  %-12s %d\n", "NEW:", stateCounts[models.StateNew])
	_, _ = fmt.Fprintf(w, "  %-12s %d\n", "IN_PROGRESS:", stateCounts[models.StateInProgress])
	if !activeOnly {
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "DONE:", stateCounts[models.StateDone])
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "CANCELLED:", stateCounts[models.StateCancelled])
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "INVALID:", stateCounts[models.StateInvalid])
	}

	if !activeOnly {
		_, _ = fmt.Fprintln(w)

		// By Type
		_, _ = fmt.Fprintln(w, "By Type:")
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "Bug:", typeCounts["Bug"])
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "Feature:", typeCounts["Feature"])
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "Regression:", typeCounts["Regression"])
		_, _ = fmt.Fprintln(w)

		// By Priority
		_, _ = fmt.Fprintln(w, "By Priority:")
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "High:", priorityCounts[models.PriorityHigh])
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "Medium:", priorityCounts[models.PriorityMedium])
		_, _ = fmt.Fprintf(w, "  %-12s %d\n", "Low:", priorityCounts[models.PriorityLow])
		_, _ = fmt.Fprintln(w)

		// Special categories
		_, _ = fmt.Fprintln(w, "Special:")
		_, _ = fmt.Fprintf(w, "  %-13s %d\n", "Blocked:", blockedCount)
		_, _ = fmt.Fprintf(w, "  %-13s %d\n", "Parent tasks:", parentCount)
		_, _ = fmt.Fprintf(w, "  %-13s %d\n", "Subtasks:", subtaskCount)
	}
}
