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

// formatTask formats a task using the new single-line compact format
func formatTask(task *models.Task, showDetails bool) string {
	return formatTaskCompact(task, showDetails)
}

// formatTaskCompact formats a task in the new compact single-line format
func formatTaskCompact(task *models.Task, showDetails bool) string {
	var b strings.Builder
	
	// Get terminal width for proper padding
	width := getTerminalWidth()
	
	// Build the main line: [ID] priority state KIND title #tags
	var mainParts []string
	
	// ID with brackets
	idPart := fmt.Sprintf("[%d]", task.ID)
	if useColor {
		idPart = colorize(idPart, colorBold)
	}
	mainParts = append(mainParts, idPart)
	
	// Priority indicator
	mainParts = append(mainParts, getPriorityIndicator(task.Priority))
	
	// State indicator
	if useColor {
		mainParts = append(mainParts, formatStateColor(task.State))
	} else {
		mainParts = append(mainParts, getStateEmoji(task.State))
	}
	
	// Task kind
	mainParts = append(mainParts, formatKindColor(formatKind(task.Kind)))
	
	// Title
	title := task.Title
	if useColor {
		title = colorize(title, colorBold)
	}
	mainParts = append(mainParts, title)
	
	// Tags with # prefix
	if task.Tags != "" {
		mainParts = append(mainParts, formatTagsColor(task.Tags))
	}
	
	// Blocked indicator
	if task.IsBlocked() {
		blocked := emojiBlocked
		if useColor {
			blocked = colorize(blocked, colorRed)
		}
		mainParts = append(mainParts, blocked)
	}
	
	// Build main line
	mainLine := strings.Join(mainParts, " ")
	
	if showDetails {
		// Build the metadata part: [ STATE | PRIORITY | Created: date ]
		var metaParts []string
		metaParts = append(metaParts, task.State)
		metaParts = append(metaParts, strings.ToUpper(task.Priority))
		metaParts = append(metaParts, fmt.Sprintf("Created: %s", task.Created.Format("2006-01-02")))
		
		// Add optional metadata
		if task.Source != "" {
			metaParts = append(metaParts, fmt.Sprintf("Source: %s", task.Source))
		}
		
		if task.IsBlocked() && task.BlockedBy != nil {
			blocked := fmt.Sprintf("Blocked by: #%d", *task.BlockedBy)
			if useColor {
				blocked = colorize(blocked, colorRed)
			}
			metaParts = append(metaParts, blocked)
		}
		
		// Format the line with padding
		metaLine := fmt.Sprintf("[ %s ]", strings.Join(metaParts, " | "))
		
		// Calculate padding
		mainLen := visibleLength(mainLine)
		metaLen := visibleLength(metaLine)
		totalLen := mainLen + metaLen
		
		if totalLen < width-1 {
			// Add padding between main and meta
			padding := width - totalLen - 1
			fmt.Fprintf(&b, "%s%s%s", mainLine, strings.Repeat(" ", padding), metaLine)
		} else {
			// Too long, put metadata on next line
			fmt.Fprintf(&b, "%s\n%s%s", mainLine, strings.Repeat(" ", 4), metaLine)
		}
		
		// Add description if present
		if task.Description != "" {
			fmt.Fprintf(&b, "\n\nDescription:\n%s\n%s",
				strings.Repeat("-", 30), task.Description)
		}
	} else {
		// Just the main line for oneline format
		b.WriteString(mainLine)
	}
	
	return b.String()
}

// formatTaskOneline formats a task for oneline output using compact format
func formatTaskOneline(task *models.Task) string {
	return formatTaskCompact(task, false)
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