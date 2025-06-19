package cmd

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorDim   = "\033[2m"

	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"

	colorBrightRed    = "\033[91m"
	colorBrightGreen  = "\033[92m"
	colorBrightYellow = "\033[93m"
)

var (
	// Check if we should use colors
	useColor = isColorTerminal()
)

// isColorTerminal checks if the terminal supports colors
func isColorTerminal() bool {
	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	// Check TERM environment variable
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		return false
	}

	// Check NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	return true
}

// getTerminalWidth returns the terminal width or a default
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < 40 {
		return 80 // default width
	}
	return width
}

// colorize applies color to text if colors are enabled
func colorize(text, color string) string {
	if !useColor {
		return text
	}
	return color + text + colorReset
}

// formatPriorityColor returns colored priority indicator
func formatPriorityColor(priority string) string {
	switch priority {
	case "high":
		return colorize("!", colorBrightRed)
	case "medium":
		return colorize("=", colorYellow)
	case "low":
		return colorize("-", colorGreen)
	default:
		return "."
	}
}

// formatStateColor returns colored state indicator
func formatStateColor(state string) string {
	switch state {
	case "NEW":
		return colorize("◆", colorCyan)
	case "IN_PROGRESS":
		return colorize("▶", colorBrightYellow)
	case "DONE":
		return colorize("✓", colorBrightGreen)
	case "CANCELLED":
		return colorize("✗", colorGray)
	default:
		return "?"
	}
}

// formatKindColor returns colored task kind
func formatKindColor(kind string) string {
	switch kind {
	case "BUG":
		return colorize("BUG", colorRed)
	case "FEATURE":
		return colorize("FEATURE", colorGreen)
	case "REGRESSION":
		return colorize("REGRESSION", colorYellow)
	default:
		return kind
	}
}

// formatTagsColor returns colored tags with # prefix
func formatTagsColor(tags string) string {
	if tags == "" {
		return ""
	}

	tagList := strings.Split(tags, ",")
	coloredTags := make([]string, len(tagList))
	for i, tag := range tagList {
		coloredTags[i] = colorize("#"+strings.TrimSpace(tag), colorBlue)
	}
	return strings.Join(coloredTags, " ")
}

// padRight pads string to the right with spaces
func padRight(s string, length int) string {
	// Account for ANSI color codes when calculating visible length
	visibleLen := visibleLength(s)
	if visibleLen >= length {
		return s
	}
	return s + strings.Repeat(" ", length-visibleLen)
}

// visibleLength returns the visible length of a string (excluding ANSI codes)
func visibleLength(s string) int {
	// Simple approach: count runes, ignoring ANSI escape sequences
	visible := 0
	inEscape := false

	for _, r := range s {
		if r == '\033' {
			inEscape = true
		} else if inEscape {
			if r == 'm' {
				inEscape = false
			}
		} else {
			visible++
		}
	}

	return visible
}
