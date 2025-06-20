package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/zw3rk/gtd/internal/models"
	"github.com/zw3rk/gtd/internal/output"
)

// SubtaskStats is re-exported from output package for compatibility
type SubtaskStats = output.SubtaskStats

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

// formatTaskGitStyle formats a task in git log style - wrapper for compatibility
func formatTaskGitStyle(task *models.Task, subtaskStats *SubtaskStats) string {
	// Use the centralized formatter if colors are disabled
	if !useColor {
		return output.FormatTaskGitStyle(task, subtaskStats)
	}

	// Keep the colored version here for now
	var b strings.Builder

	// Line 1: task <full-hash>
	hashLine := fmt.Sprintf("task %s", task.ID)
	if useColor {
		b.WriteString(colorize("task", colorYellow))
		b.WriteString(" ")
		b.WriteString(colorize(task.ID, colorYellow))
	} else {
		b.WriteString(hashLine)
	}
	b.WriteString("\n")

	// Line 2: Author: Name <email>
	b.WriteString("Author: ")
	b.WriteString(task.Author)
	b.WriteString("\n")

	// Line 3: Date: timestamp
	b.WriteString("Date:   ")
	b.WriteString(task.Created.Format("Mon Jan 2 15:04:05 2006 -0700"))
	b.WriteString("\n\n")

	// Line 4 (indented): state indicator + kind(priority): title
	b.WriteString("  ")

	// State indicator
	if useColor {
		b.WriteString(formatStateColor(task.State))
	} else {
		b.WriteString(getStateEmoji(task.State))
	}
	b.WriteString(" ")

	// Format kind(priority):
	kindPriority := fmt.Sprintf("%s(%s): ", strings.ToLower(task.Kind), task.Priority)
	if useColor {
		b.WriteString(formatKindPriorityColor(task.Kind, task.Priority))
	} else {
		b.WriteString(kindPriority)
	}

	// Title
	title := task.Title
	if subtaskStats != nil && subtaskStats.Total > 0 {
		title = fmt.Sprintf("%s (%d/%d)", task.Title, subtaskStats.Done, subtaskStats.Total)
	}
	if useColor {
		b.WriteString(colorize(title, colorBold))
	} else {
		b.WriteString(title)
	}
	b.WriteString("\n")

	// Body (indented with extra indent)
	if task.Description != "" {
		b.WriteString("\n")
		for _, line := range strings.Split(task.Description, "\n") {
			b.WriteString("    ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Blocked-by (if applicable)
	if task.IsBlocked() && task.BlockedBy != nil {
		b.WriteString("\n    Blocked-by: ")
		if useColor {
			b.WriteString(colorize(*task.BlockedBy, colorRed))
		} else {
			b.WriteString(*task.BlockedBy)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatTaskCompact formats a task in the new compact single-line format
func formatTaskCompact(task *models.Task, showDetails bool) string {
	var b strings.Builder

	// Build the main line: hash state kind(priority): title #tags
	var mainParts []string

	// Hash at the beginning (like git log)
	hash := task.ShortHash()
	if useColor {
		hash = colorize(hash, colorYellow)
	}
	mainParts = append(mainParts, hash)

	// State indicator
	if useColor {
		mainParts = append(mainParts, formatStateColor(task.State))
	} else {
		mainParts = append(mainParts, getStateEmoji(task.State))
	}

	// kind(priority): format
	kindPriority := fmt.Sprintf("%s(%s):", strings.ToLower(task.Kind), task.Priority)
	if useColor {
		kindPriority = formatKindPriorityColor(task.Kind, task.Priority)
	}
	mainParts = append(mainParts, kindPriority)

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
	b.WriteString(mainLine)

	if showDetails {
		// Add description if present, indented
		if task.Description != "" {
			b.WriteString("\n")
			// Split description into lines and indent each
			descLines := strings.Split(task.Description, "\n")
			for _, line := range descLines {
				fmt.Fprintf(&b, "    %s\n", line)
			}
		}

		// Add metadata as part of the body if relevant
		if task.IsBlocked() && task.BlockedBy != nil {
			fmt.Fprintf(&b, "\n    Blocked by: %s\n", *task.BlockedBy)
		}
	}

	return b.String()
}

// formatTaskOneline formats a task for oneline output using compact format
func formatTaskOneline(task *models.Task) string {
	if !useColor {
		return output.FormatTaskOneline(task)
	}
	return formatTaskCompact(task, false)
}

// formatSubtask formats a subtask - wrapper for compatibility
func formatSubtask(task *models.Task) string {
	if !useColor {
		return output.FormatSubtask(task)
	}

	// Keep the colored version here for now
	// Build the main line: hash state kind(priority): title #tags
	var mainParts []string

	// Hash at the beginning (like git log)
	hash := task.ShortHash()
	if useColor {
		hash = colorize(hash, colorYellow)
	}
	mainParts = append(mainParts, hash)

	// State indicator
	if useColor {
		mainParts = append(mainParts, formatStateColor(task.State))
	} else {
		mainParts = append(mainParts, getStateEmoji(task.State))
	}

	// kind(priority): format
	kindPriority := fmt.Sprintf("%s(%s):", strings.ToLower(task.Kind), task.Priority)
	if useColor {
		kindPriority = formatKindPriorityColor(task.Kind, task.Priority)
	}
	mainParts = append(mainParts, kindPriority)

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
	return strings.Join(mainParts, " ")
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

// formatTaskList formats a list of tasks for output
func formatTaskList(w io.Writer, tasks []*models.Task, oneline bool) {
	formatter := output.NewFormatter(w)
	if err := formatter.FormatTaskList(tasks, oneline); err != nil {
		// Ignore write errors for now
		return
	}
}

// formatKindPriorityColor formats kind(priority) with appropriate colors
func formatKindPriorityColor(kind, priority string) string {
	// Format the kind part
	kindLower := strings.ToLower(kind)
	var kindColored string
	switch kind {
	case models.KindBug:
		kindColored = colorize(kindLower, colorRed)
	case models.KindFeature:
		kindColored = colorize(kindLower, colorGreen)
	case models.KindRegression:
		kindColored = colorize(kindLower, colorYellow)
	default:
		kindColored = kindLower
	}

	// Format the priority part
	var priorityColored string
	switch priority {
	case models.PriorityHigh:
		priorityColored = colorize(priority, colorBrightRed)
	case models.PriorityMedium:
		priorityColored = colorize(priority, colorYellow)
	case models.PriorityLow:
		priorityColored = colorize(priority, colorGreen)
	default:
		priorityColored = priority
	}

	return fmt.Sprintf("%s(%s): ", kindColored, priorityColored)
}
