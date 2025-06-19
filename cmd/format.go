package cmd

import (
	"fmt"
	"strings"

	"github.com/zw3rk/claude-gtd/internal/models"
)

// Priority emojis
const (
	emojiHigh   = "🔴"
	emojiMedium = "🟡"
	emojiLow    = "🟢"
)

// State emojis
const (
	emojiNew        = "📋"
	emojiInProgress = "🔄"
	emojiDone       = "✅"
	emojiCancelled  = "❌"
	emojiBlocked    = "🚫"
)

// formatTask formats a task for standard output
func formatTask(task *models.Task, showDetails bool) string {
	var b strings.Builder
	
	// First line: [ID] PRIORITY STATE [BLOCKED] TITLE
	priorityEmoji := getPriorityEmoji(task.Priority)
	
	// Basic format
	fmt.Fprintf(&b, "[%d] %s %s %s", task.ID, priorityEmoji, task.State, task.Title)
	
	// Add blocked indicator if blocked
	if task.IsBlocked() {
		b.WriteString(" " + emojiBlocked)
	}
	
	// Add details on subsequent lines if requested
	if showDetails {
		if task.Source != "" {
			fmt.Fprintf(&b, "\n     Source: %s", task.Source)
		}
		if task.Tags != "" {
			fmt.Fprintf(&b, "\n     Tags: %s", task.Tags)
		}
		if task.IsBlocked() && task.BlockedBy != nil {
			fmt.Fprintf(&b, "\n     Blocked by: #%d", *task.BlockedBy)
		}
		if task.Description != "" {
			// Show first line of description
			descLines := strings.Split(task.Description, "\n")
			fmt.Fprintf(&b, "\n     %s", descLines[0])
			if len(descLines) > 1 {
				b.WriteString("...")
			}
		}
	}
	
	return b.String()
}

// formatTaskOneline formats a task for oneline output
func formatTaskOneline(task *models.Task) string {
	priorityEmoji := getPriorityEmoji(task.Priority)
	
	blocked := ""
	if task.IsBlocked() {
		blocked = " " + emojiBlocked
	}
	
	return fmt.Sprintf("[%d] %s %s %s%s", 
		task.ID, priorityEmoji, task.State, task.Title, blocked)
}

// getPriorityEmoji returns the emoji for a priority level
func getPriorityEmoji(priority string) string {
	switch priority {
	case models.PriorityHigh:
		return emojiHigh
	case models.PriorityMedium:
		return emojiMedium
	case models.PriorityLow:
		return emojiLow
	default:
		return "⚪"
	}
}

// getStateEmoji returns the emoji for a state
func getStateEmoji(state string) string {
	switch state {
	case models.StateNew:
		return emojiNew
	case models.StateInProgress:
		return emojiInProgress
	case models.StateDone:
		return emojiDone
	case models.StateCancelled:
		return emojiCancelled
	default:
		return "❓"
	}
}

// formatTaskCount formats a count with proper pluralization
func formatTaskCount(count int, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}