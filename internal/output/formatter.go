package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/zw3rk/gtd/internal/models"
)

// Formatter handles formatting of tasks for display
type Formatter struct {
	writer io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(w io.Writer) *Formatter {
	return &Formatter{writer: w}
}

// FormatTask formats a single task in git-style format
func (f *Formatter) FormatTask(task *models.Task, stats *SubtaskStats) error {
	output := FormatTaskGitStyle(task, stats)
	_, err := fmt.Fprint(f.writer, output)
	return err
}

// FormatTaskList formats a list of tasks
func (f *Formatter) FormatTaskList(tasks []*models.Task, oneline bool) error {
	if oneline {
		return f.formatTasksOneline(tasks)
	}
	return f.formatTasksStandard(tasks)
}

// formatTasksStandard formats tasks in standard multi-line format
func (f *Formatter) formatTasksStandard(tasks []*models.Task) error {
	for i, task := range tasks {
		if i > 0 {
			if _, err := fmt.Fprintln(f.writer); err != nil {
				return err
			}
		}

		output := FormatTaskGitStyle(task, nil)
		if _, err := fmt.Fprint(f.writer, output); err != nil {
			return err
		}
	}
	return nil
}

// formatTasksOneline formats tasks in compact one-line format
func (f *Formatter) formatTasksOneline(tasks []*models.Task) error {
	for _, task := range tasks {
		line := FormatTaskOneline(task)
		if _, err := fmt.Fprintln(f.writer, line); err != nil {
			return err
		}
	}
	return nil
}

// SubtaskStats holds statistics about subtasks
type SubtaskStats struct {
	Total int
	Done  int
}

// FormatTaskGitStyle formats a task in git-log style
func FormatTaskGitStyle(task *models.Task, stats *SubtaskStats) string {
	var sb strings.Builder

	// Header line
	fmt.Fprintf(&sb, "task %s\n", task.ID)
	fmt.Fprintf(&sb, "Author: %s\n", task.Author)
	fmt.Fprintf(&sb, "Date:   %s\n", task.Created.Format(time.RFC1123Z))

	// Parent reference if subtask
	if task.Parent != nil {
		fmt.Fprintf(&sb, "Parent: %s\n", *task.Parent)
	}

	// Empty line before content
	sb.WriteString("\n")

	// Status icon and metadata
	icon := getStateIcon(task.State)
	fmt.Fprintf(&sb, "  %s %s(%s): %s", icon, strings.ToLower(task.Kind), task.Priority, task.Title)

	// Add subtask progress if parent
	if stats != nil && stats.Total > 0 {
		fmt.Fprintf(&sb, " [%d/%d]", stats.Done, stats.Total)
	}

	sb.WriteString("\n")

	// Body with proper indentation
	if task.Description != "" {
		for _, line := range strings.Split(task.Description, "\n") {
			fmt.Fprintf(&sb, "    %s\n", line)
		}
	}

	// Metadata section
	var metadata []string
	if task.Source != "" {
		metadata = append(metadata, fmt.Sprintf("Source: %s", task.Source))
	}
	if task.BlockedBy != nil {
		metadata = append(metadata, fmt.Sprintf("Blocked-by: %s", (*task.BlockedBy)[:7]))
	}
	if task.Tags != "" {
		metadata = append(metadata, fmt.Sprintf("Tags: %s", task.Tags))
	}

	if len(metadata) > 0 {
		sb.WriteString("\n")
		for _, meta := range metadata {
			fmt.Fprintf(&sb, "    %s\n", meta)
		}
	}

	return sb.String()
}

// FormatTaskOneline formats a task in a single line
func FormatTaskOneline(task *models.Task) string {
	icon := getStateIcon(task.State)
	line := fmt.Sprintf("%s %s %s(%s): %s",
		task.ShortHash(),
		icon,
		strings.ToLower(task.Kind),
		task.Priority,
		task.Title)

	if task.IsBlocked() {
		line += " [BLOCKED]"
	}

	return line
}

// FormatSubtask formats a subtask with metadata
func FormatSubtask(task *models.Task) string {
	// Format with metadata on the right
	icon := getStateIcon(task.State)
	base := fmt.Sprintf("%s %s - %s", task.ShortHash(), icon, task.Title)

	// Add metadata to the right
	var metadata []string
	metadata = append(metadata, strings.ToLower(task.Kind))
	metadata = append(metadata, task.Priority)

	if task.IsBlocked() {
		metadata = append(metadata, "blocked")
	}

	// Calculate padding for alignment
	const targetWidth = 80
	baseLen := len(base)
	metaStr := strings.Join(metadata, ", ")
	padding := targetWidth - baseLen - len(metaStr) - 3 // 3 for " | "

	if padding < 2 {
		padding = 2
	}

	return fmt.Sprintf("%s%s| %s", base, strings.Repeat(" ", padding), metaStr)
}

// getStateIcon returns an icon for the task state
func getStateIcon(state string) string {
	switch state {
	case models.StateInbox:
		return "?"
	case models.StateNew:
		return "◆"
	case models.StateInProgress:
		return "▶"
	case models.StateDone:
		return "✓"
	case models.StateCancelled:
		return "✗"
	case models.StateInvalid:
		return "⊘"
	default:
		return "·"
	}
}