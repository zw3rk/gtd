package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// readTaskInput reads title and optional description from stdin
func readTaskInput(r io.Reader) (title, description string, err error) {
	scanner := bufio.NewScanner(r)
	
	// Read title (first line)
	if scanner.Scan() {
		title = strings.TrimSpace(scanner.Text())
	}
	
	if title == "" {
		return "", "", fmt.Errorf("title cannot be empty")
	}
	
	// Read description (remaining lines)
	var descLines []string
	for scanner.Scan() {
		line := scanner.Text()
		// Stop reading if we encounter EOF marker or empty line after content
		if line == "EOF" || (line == "" && len(descLines) > 0) {
			break
		}
		descLines = append(descLines, line)
	}
	
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error reading input: %w", err)
	}
	
	description = strings.TrimSpace(strings.Join(descLines, "\n"))
	
	return title, description, nil
}

// formatTaskCreated formats the output message for a created task
func formatTaskCreated(id int, kind string) string {
	return fmt.Sprintf("Created %s task #%d", strings.ToLower(kind), id)
}