package cmd

import (
	"fmt"
	"strings"

	"github.com/zw3rk/claude-gtd/internal/models"
)

// Priority indicators
const (
	emojiHigh   = "!" // Exclamation mark for high priority
	emojiMedium = "=" // Equals sign for medium priority
	emojiLow    = "-" // Hyphen for low priority
)

// State indicators
const (
	emojiNew        = "◆" // U+25C6 - Black Diamond (was 📋)
	emojiInProgress = "▶" // U+25B6 - Black Right-Pointing Triangle (was 🔄)
	emojiDone       = "✓" // U+2713 - Check Mark (was ✅)
	emojiCancelled  = "✗" // U+2717 - Ballot X (was ❌)
	emojiBlocked    = "⊘" // U+2298 - Circled Division Slash (was 🚫)
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
	var parts []string
	
	// ID with brackets
	idPart := fmt.Sprintf("[%d]", task.ID)
	if useColor {
		idPart = colorize(idPart, colorBold)
	}
	parts = append(parts, idPart)
	
	// Priority with color
	parts = append(parts, getPriorityIndicator(task.Priority))
	
	// State
	state := task.State
	if useColor {
		state = formatStateColor(task.State)
	} else {
		state = fmt.Sprintf("%s %s", state, getStateEmoji(task.State))
	}
	parts = append(parts, state)
	
	// Title
	title := task.Title
	if useColor {
		title = colorize(title, colorBold)
	}
	parts = append(parts, title)
	
	// Blocked indicator
	if task.IsBlocked() {
		blocked := emojiBlocked
		if useColor {
			blocked = colorize(blocked, colorRed)
		}
		parts = append(parts, blocked)
	}
	
	return strings.Join(parts, " ")
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
		return "." // Period for unknown/default priority
	}
}

// getPriorityIndicator returns the priority indicator (with color if enabled)
func getPriorityIndicator(priority string) string {
	if useColor {
		return formatPriorityColor(priority)
	}
	return getPriorityEmoji(priority)
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
		return "?" // Question mark
	}
}

// formatTaskCount formats a count with proper pluralization
func formatTaskCount(count int, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}

// formatKind formats a task kind for display
func formatKind(kind string) string {
	switch kind {
	case models.KindBug:
		return "Bug"
	case models.KindFeature:
		return "Feature"
	case models.KindRegression:
		return "Regression"
	default:
		return kind
	}
}