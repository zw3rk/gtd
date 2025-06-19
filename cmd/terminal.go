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

// colorize applies color to text if colors are enabled
func colorize(text, color string) string {
	if !useColor {
		return text
	}
	return color + text + colorReset
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
