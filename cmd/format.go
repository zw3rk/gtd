package cmd

import (
	"fmt"
	"strings"

	"github.com/zw3rk/claude-gtd/internal/models"
)

// SubtaskStats holds subtask completion statistics
type SubtaskStats struct {
	Total int
	Done  int
}

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

// formatTaskGitStyle formats a task in git log style
func formatTaskGitStyle(task *models.Task, subtaskStats *SubtaskStats) string {
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
	b.WriteString("\n")
	
	// Line 4: Status: state
	b.WriteString("Status: ")
	if useColor {
		b.WriteString(formatStateColor(task.State))
	} else {
		b.WriteString(task.State)
	}
	b.WriteString("\n\n")
	
	// Line 5 (indented): kind(priority): title
	b.WriteString("    ")
	
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
	
	// Body (indented)
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
	
	// Get terminal width for proper padding
	width := getTerminalWidth()
	
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
			blocked := fmt.Sprintf("Blocked by: %s", *task.BlockedBy)
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
		
		// Add description if present, indented
		if task.Description != "" {
			// Split description into lines and indent each
			descLines := strings.Split(task.Description, "\n")
			for _, line := range descLines {
				fmt.Fprintf(&b, "\n    %s", line)
			}
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

// formatSubtask formats a subtask with metadata on the right side
func formatSubtask(task *models.Task) string {
	// Get terminal width for proper padding, account for 2-char indent
	width := getTerminalWidth() - 2
	
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
	
	// Build the metadata part: [ STATE | PRIORITY ]
	var metaParts []string
	metaParts = append(metaParts, task.State)
	metaParts = append(metaParts, strings.ToUpper(task.Priority))
	
	// Add blocked info if needed
	if task.IsBlocked() && task.BlockedBy != nil {
		blocked := fmt.Sprintf("Blocked by: %s", *task.BlockedBy)
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
	
	var result strings.Builder
	if totalLen < width-1 {
		// Add padding between main and meta
		padding := width - totalLen - 1
		fmt.Fprintf(&result, "%s%s%s", mainLine, strings.Repeat(" ", padding), metaLine)
	} else {
		// Too long, put metadata on next line
		fmt.Fprintf(&result, "%s\n%s%s", mainLine, strings.Repeat(" ", 6), metaLine)
	}
	
	return result.String()
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