package output_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/models"
	"github.com/zw3rk/gtd/internal/output"
)

// ColorFormatter simulates the color formatting from cmd/format.go
type ColorFormatter struct {
	useColor bool
}

// ANSI color codes
const (
	colorReset      = "\033[0m"
	colorBold       = "\033[1m"
	colorRed        = "\033[31m"
	colorGreen      = "\033[32m"
	colorYellow     = "\033[33m"
	colorBlue       = "\033[34m"
	colorMagenta    = "\033[35m"
	colorCyan       = "\033[36m"
	colorBrightRed  = "\033[91m"
	colorBrightBlue = "\033[94m"
	colorDim        = "\033[2m"
)

func (f *ColorFormatter) colorize(text, color string) string {
	if !f.useColor {
		return text
	}
	return color + text + colorReset
}

func (f *ColorFormatter) formatStateColor(state string) string {
	icon := getStateIcon(state)
	if !f.useColor {
		return icon
	}
	
	switch state {
	case models.StateNew:
		return f.colorize(icon, colorGreen)
	case models.StateInProgress:
		return f.colorize(icon, colorYellow)
	case models.StateDone:
		return f.colorize(icon, colorGreen)
	case models.StateCancelled:
		return f.colorize(icon, colorDim)
	default:
		return icon
	}
}

func (f *ColorFormatter) formatKindPriorityColor(kind, priority string) string {
	// Format the kind part
	kindLower := strings.ToLower(kind)
	var kindColored string
	switch kind {
	case models.KindBug:
		kindColored = f.colorize(kindLower, colorRed)
	case models.KindFeature:
		kindColored = f.colorize(kindLower, colorGreen)
	case models.KindRegression:
		kindColored = f.colorize(kindLower, colorYellow)
	default:
		kindColored = kindLower
	}
	
	// Format the priority part
	var priorityColored string
	switch priority {
	case models.PriorityHigh:
		priorityColored = f.colorize(priority, colorBrightRed)
	case models.PriorityMedium:
		priorityColored = f.colorize(priority, colorYellow)
	case models.PriorityLow:
		priorityColored = f.colorize(priority, colorGreen)
	default:
		priorityColored = priority
	}
	
	return fmt.Sprintf("%s(%s): ", kindColored, priorityColored)
}

func (f *ColorFormatter) formatTagsColor(tags string) string {
	if !f.useColor {
		return "#" + tags
	}
	return f.colorize("#"+tags, colorBlue)
}

func (f *ColorFormatter) FormatTask(task *models.Task, subtaskStats *output.SubtaskStats) string {
	var b strings.Builder
	
	// Line 1: task <full-hash>
	if f.useColor {
		b.WriteString(f.colorize("task", colorYellow))
		b.WriteString(" ")
		b.WriteString(f.colorize(task.ID, colorYellow))
	} else {
		b.WriteString(fmt.Sprintf("task %s", task.ID))
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
	b.WriteString(f.formatStateColor(task.State))
	b.WriteString(" ")
	
	// Format kind(priority):
	if f.useColor {
		b.WriteString(f.formatKindPriorityColor(task.Kind, task.Priority))
	} else {
		b.WriteString(fmt.Sprintf("%s(%s): ", strings.ToLower(task.Kind), task.Priority))
	}
	
	// Title
	title := task.Title
	if subtaskStats != nil && subtaskStats.Total > 0 {
		title = fmt.Sprintf("%s (%d/%d)", task.Title, subtaskStats.Done, subtaskStats.Total)
	}
	if f.useColor {
		b.WriteString(f.colorize(title, colorBold))
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
		if f.useColor {
			b.WriteString(f.colorize(*task.BlockedBy, colorRed))
		} else {
			b.WriteString(*task.BlockedBy)
		}
		b.WriteString("\n")
	}
	
	return b.String()
}

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

// Tests
func TestColorFormatting(t *testing.T) {
	task := createTestTask("color123", "Colorful Task")
	
	t.Run("Colors enabled", func(t *testing.T) {
		formatter := &ColorFormatter{useColor: true}
		output := formatter.FormatTask(task, nil)
		
		// Should contain ANSI color codes
		if !strings.Contains(output, "\033[") {
			t.Error("Expected ANSI color codes in output with colors enabled")
		}
		
		// Check specific color codes
		if !strings.Contains(output, colorYellow) {
			t.Error("Expected yellow color for task hash")
		}
		if !strings.Contains(output, colorBold) {
			t.Error("Expected bold color for title")
		}
	})
	
	t.Run("Colors disabled", func(t *testing.T) {
		formatter := &ColorFormatter{useColor: false}
		output := formatter.FormatTask(task, nil)
		
		// Should NOT contain ANSI color codes
		if strings.Contains(output, "\033[") {
			t.Error("Should not contain ANSI color codes with colors disabled")
		}
		
		// Should still contain all the content
		if !strings.Contains(output, task.ID) {
			t.Error("Missing task ID in non-colored output")
		}
		if !strings.Contains(output, task.Title) {
			t.Error("Missing task title in non-colored output")
		}
	})
}

func TestColoredStateIcons(t *testing.T) {
	states := []struct {
		state         string
		expectedIcon  string
		expectedColor string
	}{
		{models.StateNew, "◆", colorGreen},
		{models.StateInProgress, "▶", colorYellow},
		{models.StateDone, "✓", colorGreen},
		{models.StateCancelled, "✗", colorDim},
	}
	
	formatter := &ColorFormatter{useColor: true}
	
	for _, tt := range states {
		t.Run(tt.state, func(t *testing.T) {
			colored := formatter.formatStateColor(tt.state)
			
			// Should contain the icon
			if !strings.Contains(colored, tt.expectedIcon) {
				t.Errorf("Missing icon %s for state %s", tt.expectedIcon, tt.state)
			}
			
			// Should contain the color code
			if !strings.Contains(colored, tt.expectedColor) {
				t.Errorf("Missing color code %s for state %s", tt.expectedColor, tt.state)
			}
			
			// Should end with reset
			if !strings.Contains(colored, colorReset) {
				t.Error("Missing color reset code")
			}
		})
	}
}

func TestColoredKindPriority(t *testing.T) {
	tests := []struct {
		kind         string
		priority     string
		kindColor    string
		priorityColor string
	}{
		{models.KindBug, models.PriorityHigh, colorRed, colorBrightRed},
		{models.KindFeature, models.PriorityMedium, colorGreen, colorYellow},
		{models.KindRegression, models.PriorityLow, colorYellow, colorGreen},
	}
	
	formatter := &ColorFormatter{useColor: true}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.kind, tt.priority), func(t *testing.T) {
			colored := formatter.formatKindPriorityColor(tt.kind, tt.priority)
			
			// Should contain kind in lowercase
			if !strings.Contains(colored, strings.ToLower(tt.kind)) {
				t.Errorf("Missing kind %s", tt.kind)
			}
			
			// Should contain priority
			if !strings.Contains(colored, tt.priority) {
				t.Errorf("Missing priority %s", tt.priority)
			}
			
			// Should contain color codes
			if !strings.Contains(colored, tt.kindColor) {
				t.Errorf("Missing kind color %s", tt.kindColor)
			}
			if !strings.Contains(colored, tt.priorityColor) {
				t.Errorf("Missing priority color %s", tt.priorityColor)
			}
		})
	}
}

func TestColoredTags(t *testing.T) {
	formatter := &ColorFormatter{useColor: true}
	
	tags := "backend,api,urgent"
	colored := formatter.formatTagsColor(tags)
	
	// Should have # prefix
	if !strings.HasPrefix(strings.TrimPrefix(colored, colorBlue), "#") {
		t.Error("Tags should have # prefix")
	}
	
	// Should be blue
	if !strings.Contains(colored, colorBlue) {
		t.Error("Tags should be colored blue")
	}
	
	// Should contain the actual tags
	if !strings.Contains(colored, tags) {
		t.Error("Missing tag content")
	}
}

func TestColoredBlockedBy(t *testing.T) {
	task := createTestTask("blocked123", "Blocked Task")
	blocker := "blocker456"
	task.BlockedBy = &blocker
	
	formatter := &ColorFormatter{useColor: true}
	output := formatter.FormatTask(task, nil)
	
	// Should contain blocked-by in red
	if !strings.Contains(output, "Blocked-by:") {
		t.Error("Missing Blocked-by label")
	}
	
	// The blocker ID should be in red
	blockedSection := output[strings.Index(output, "Blocked-by:"):]
	if !strings.Contains(blockedSection, colorRed) {
		t.Error("Blocked-by ID should be colored red")
	}
}

func TestColorFormatterComparison(t *testing.T) {
	// Test that colored and non-colored output have the same content
	task := createTestTask("compare123", "Comparison Task")
	task.Description = "Multi-line\ndescription\nfor testing"
	blocker := "blocker789"
	task.BlockedBy = &blocker
	
	coloredFormatter := &ColorFormatter{useColor: true}
	plainFormatter := &ColorFormatter{useColor: false}
	
	coloredOutput := coloredFormatter.FormatTask(task, nil)
	plainOutput := plainFormatter.FormatTask(task, nil)
	
	// Strip ANSI codes from colored output
	strippedColored := stripANSI(coloredOutput)
	
	// Should be identical after stripping colors
	if strippedColored != plainOutput {
		t.Errorf("Colored and plain output differ:\nColored (stripped):\n%s\nPlain:\n%s", 
			strippedColored, plainOutput)
	}
}

// Helper to strip ANSI color codes
func stripANSI(s string) string {
	// Simple ANSI code stripper
	result := s
	for strings.Contains(result, "\033[") {
		start := strings.Index(result, "\033[")
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

// Test indentation behavior
func TestIndentationWithColors(t *testing.T) {
	task := createTestTask("indent123", "Indented Task")
	task.Description = "Line 1\nLine 2\nLine 3"
	
	formatter := &ColorFormatter{useColor: true}
	output := formatter.FormatTask(task, nil)
	
	lines := strings.Split(output, "\n")
	
	// Find description lines
	descStarted := false
	for _, line := range lines {
		stripped := stripANSI(line)
		if strings.TrimSpace(stripped) == "" && descStarted {
			// Empty line after title, before description
			continue
		}
		if strings.Contains(stripped, "Line ") {
			descStarted = true
			// Each description line should start with 4 spaces
			if !strings.HasPrefix(stripped, "    ") {
				t.Errorf("Description line not properly indented: %q", stripped)
			}
		}
	}
}

// Test that output package formatter matches non-colored formatter
func TestOutputPackageCompatibility(t *testing.T) {
	task := createTestTask("compat123", "Compatibility Test")
	
	// Get output from package formatter
	packageOutput := output.FormatTaskGitStyle(task, nil)
	
	// Get output from our non-colored formatter
	formatter := &ColorFormatter{useColor: false}
	ourOutput := formatter.FormatTask(task, nil)
	
	// They should be very similar (might have minor formatting differences)
	// Check that both contain the essential elements
	essentials := []string{
		"task " + task.ID,
		"Author: " + task.Author,
		"Date:",
		getStateIcon(task.State),
		strings.ToLower(task.Kind) + "(" + task.Priority + "):",
		task.Title,
		task.Description,
	}
	
	for _, essential := range essentials {
		if !strings.Contains(packageOutput, essential) {
			t.Errorf("Package output missing: %s", essential)
		}
		if !strings.Contains(ourOutput, essential) {
			t.Errorf("Our output missing: %s", essential)
		}
	}
}