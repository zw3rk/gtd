package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// readTaskInput reads title and optional description from stdin
// Supports Git-style commit message format: <title>\n\n<body>
func readTaskInput(r io.Reader) (title, description string, err error) {
	scanner := bufio.NewScanner(r)

	// Read title (first line)
	if scanner.Scan() {
		title = strings.TrimSpace(scanner.Text())
	}

	if title == "" {
		return "", "", fmt.Errorf("title cannot be empty")
	}

	// Look for blank line separator (Git-style)
	hasBlankLine := false
	if scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			hasBlankLine = true
		} else {
			// No blank line, this is part of the description
			var descLines []string
			descLines = append(descLines, line)

			// Continue reading remaining lines
			for scanner.Scan() {
				descLines = append(descLines, scanner.Text())
			}

			description = strings.TrimSpace(strings.Join(descLines, "\n"))
		}
	}

	// If we had a blank line, read the body
	if hasBlankLine {
		var descLines []string
		for scanner.Scan() {
			descLines = append(descLines, scanner.Text())
		}
		description = strings.TrimSpace(strings.Join(descLines, "\n"))
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error reading input: %w", err)
	}

	// Work around ZSH heredoc issue - detect and strip "EOF < /dev/null" suffix
	if shell := os.Getenv("SHELL"); strings.Contains(shell, "zsh") {
		// Check if description ends with common heredoc patterns
		patterns := []string{
			"EOF < /dev/null",
			"EOL < /dev/null",
			"END < /dev/null",
			"DONE < /dev/null",
		}
		
		for _, pattern := range patterns {
			if strings.HasSuffix(description, pattern) {
				// Strip the pattern from the end
				description = strings.TrimSuffix(description, pattern)
				description = strings.TrimSpace(description)
				break
			}
		}
	}

	return title, description, nil
}

// formatTaskCreated formats the output message for a created task
func formatTaskCreated(id string, kind string) string {
	return fmt.Sprintf("Created %s task %s", strings.ToLower(kind), id)
}
