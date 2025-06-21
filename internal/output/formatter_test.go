package output_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zw3rk/gtd/internal/models"
	"github.com/zw3rk/gtd/internal/output"
)

// Helper function to create a test task
func createTestTask(id, title string) *models.Task {
	return &models.Task{
		ID:          id,
		Kind:        models.KindFeature,
		State:       models.StateNew,
		Priority:    models.PriorityMedium,
		Title:       title,
		Description: "Test description",
		Tags:        "test,sample",
		Source:      "test.go:42",
		Author:      "Test User <test@example.com>",
		Created:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Updated:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
}

func TestNewFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := output.NewFormatter(&buf)
	if formatter == nil {
		t.Fatal("NewFormatter returned nil")
	}
}

func TestFormatterFormatTask(t *testing.T) {
	tests := []struct {
		name     string
		task     *models.Task
		stats    *output.SubtaskStats
		contains []string
	}{
		{
			name: "basic task",
			task: createTestTask("abc123def456", "Test Task"),
			contains: []string{
				"task abc123def456",
				"Author: Test User <test@example.com>",
				"Date:   Mon, 15 Jan 2024",
				"◆ feature(medium): Test Task",
				"Test description",
				"Source: test.go:42",
				"Tags: test,sample",
			},
		},
		{
			name: "task with subtask stats",
			task: createTestTask("abc123def456", "Parent Task"),
			stats: &output.SubtaskStats{
				Total: 5,
				Done:  3,
			},
			contains: []string{
				"◆ feature(medium): Parent Task [3/5]",
			},
		},
		{
			name: "subtask with parent",
			task: func() *models.Task {
				task := createTestTask("child123", "Child Task")
				parent := "parent456"
				task.Parent = &parent
				return task
			}(),
			contains: []string{
				"Parent: parent456",
				"◆ feature(medium): Child Task",
			},
		},
		{
			name: "blocked task",
			task: func() *models.Task {
				task := createTestTask("blocked123", "Blocked Task")
				blocker := "blocker456"
				task.BlockedBy = &blocker
				return task
			}(),
			contains: []string{
				"◆ feature(medium): Blocked Task",
				"Blocked-by: blocker",
			},
		},
		{
			name: "task with different states",
			task: func() *models.Task {
				task := createTestTask("state123", "State Task")
				task.State = models.StateInProgress
				return task
			}(),
			contains: []string{
				"▶ feature(medium): State Task",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := output.NewFormatter(&buf)

			err := formatter.FormatTask(tt.task, tt.stats)
			if err != nil {
				t.Fatalf("FormatTask failed: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Output missing expected content: %q\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

func TestFormatterFormatTaskList(t *testing.T) {
	tasks := []*models.Task{
		createTestTask("task1", "First Task"),
		createTestTask("task2", "Second Task"),
		createTestTask("task3", "Third Task"),
	}

	t.Run("standard format", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := output.NewFormatter(&buf)

		err := formatter.FormatTaskList(tasks, false)
		if err != nil {
			t.Fatalf("FormatTaskList failed: %v", err)
		}

		output := buf.String()
		// Check for all task IDs
		for _, task := range tasks {
			if !strings.Contains(output, task.ID) {
				t.Errorf("Output missing task ID: %s", task.ID)
			}
			if !strings.Contains(output, task.Title) {
				t.Errorf("Output missing task title: %s", task.Title)
			}
		}

		// Check for task separation (empty lines between tasks)
		lines := strings.Split(output, "\n")
		emptyLineCount := 0
		for _, line := range lines {
			if line == "" {
				emptyLineCount++
			}
		}
		// Should have 2 empty lines between 3 tasks
		if emptyLineCount < 2 {
			t.Errorf("Expected task separation, got %d empty lines", emptyLineCount)
		}
	})

	t.Run("oneline format", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := output.NewFormatter(&buf)

		err := formatter.FormatTaskList(tasks, true)
		if err != nil {
			t.Fatalf("FormatTaskList failed: %v", err)
		}

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Should have exactly 3 lines
		if len(lines) != 3 {
			t.Errorf("Expected 3 lines in oneline format, got %d", len(lines))
		}

		// Each line should contain task info
		for i, line := range lines {
			if !strings.Contains(line, tasks[i].ShortHash()) {
				t.Errorf("Line %d missing short hash", i)
			}
			if !strings.Contains(line, tasks[i].Title) {
				t.Errorf("Line %d missing title", i)
			}
		}
	})
}

func TestFormatTaskGitStyle(t *testing.T) {
	task := createTestTask("abc123def456", "Test Task")
	
	output := output.FormatTaskGitStyle(task, nil)
	
	// Check structure
	lines := strings.Split(output, "\n")
	if len(lines) < 4 {
		t.Fatalf("Expected at least 4 lines, got %d", len(lines))
	}
	
	// Check header
	if !strings.HasPrefix(lines[0], "task ") {
		t.Errorf("First line should start with 'task ', got: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "Author: ") {
		t.Errorf("Second line should start with 'Author: ', got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "Date:   ") {
		t.Errorf("Third line should start with 'Date:   ', got: %s", lines[2])
	}
	
	// Check empty line after header
	if lines[3] != "" {
		t.Errorf("Expected empty line after header, got: %s", lines[3])
	}
}

func TestFormatTaskOneline(t *testing.T) {
	tests := []struct {
		name     string
		task     *models.Task
		expected []string
	}{
		{
			name: "basic task",
			task: createTestTask("abc123def456", "Basic Task"),
			expected: []string{
				"abc123d", // short hash
				"◆",       // NEW state icon
				"feature(medium):",
				"Basic Task",
			},
		},
		{
			name: "blocked task",
			task: func() *models.Task {
				task := createTestTask("blocked123", "Blocked Task")
				blocker := "blocker456"
				task.BlockedBy = &blocker
				return task
			}(),
			expected: []string{
				"blocked",
				"◆",
				"feature(medium):",
				"Blocked Task",
				"[BLOCKED]",
			},
		},
		{
			name: "different states",
			task: func() *models.Task {
				task := createTestTask("done123", "Done Task")
				task.State = models.StateDone
				return task
			}(),
			expected: []string{
				"done123",
				"✓", // DONE state icon
				"feature(medium):",
				"Done Task",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := output.FormatTaskOneline(tt.task)
			
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Output missing expected content: %q\nGot: %s", expected, output)
				}
			}
		})
	}
}

func TestFormatSubtask(t *testing.T) {
	task := createTestTask("sub123def456", "Subtask Title")
	
	output := output.FormatSubtask(task)
	
	// Check basic structure
	if !strings.Contains(output, task.ShortHash()) {
		t.Errorf("Output missing short hash: %s", task.ShortHash())
	}
	if !strings.Contains(output, "◆") {
		t.Errorf("Output missing state icon")
	}
	if !strings.Contains(output, task.Title) {
		t.Errorf("Output missing title: %s", task.Title)
	}
	
	// Check metadata alignment
	if !strings.Contains(output, "| ") {
		t.Errorf("Output missing metadata separator")
	}
	if !strings.Contains(output, "feature, medium") {
		t.Errorf("Output missing metadata")
	}
}

func TestGetStateIcon(t *testing.T) {
	tests := []struct {
		state    string
		expected string
	}{
		{models.StateInbox, "?"},
		{models.StateNew, "◆"},
		{models.StateInProgress, "▶"},
		{models.StateDone, "✓"},
		{models.StateCancelled, "✗"},
		{models.StateInvalid, "⊘"},
		{"unknown", "·"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			task := createTestTask("test123", "Test")
			task.State = tt.state
			
			output := output.FormatTaskOneline(task)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected state icon %q for state %s, not found in: %s", 
					tt.expected, tt.state, output)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("nil task", func(t *testing.T) {
		// FormatTaskGitStyle should handle nil gracefully
		defer func() {
			if r := recover(); r == nil {
				// If no panic, check the output
				output := output.FormatTaskGitStyle(nil, nil)
				if output != "" {
					// Should either panic or return empty string
					t.Logf("FormatTaskGitStyle(nil) returned: %s", output)
				}
			}
		}()
		
		_ = output.FormatTaskGitStyle(nil, nil)
	})

	t.Run("empty task list", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := output.NewFormatter(&buf)
		
		err := formatter.FormatTaskList([]*models.Task{}, false)
		if err != nil {
			t.Fatalf("FormatTaskList failed with empty list: %v", err)
		}
		
		if buf.String() != "" {
			t.Errorf("Expected empty output for empty task list, got: %s", buf.String())
		}
	})

	t.Run("task with empty fields", func(t *testing.T) {
		task := &models.Task{
			ID:       "empty123",
			Kind:     models.KindBug,
			State:    models.StateNew,
			Priority: models.PriorityLow,
			Title:    "Empty Fields",
			Author:   "Test",
			Created:  time.Now(),
			Updated:  time.Now(),
			// Leave Description, Tags, Source empty
		}
		
		output := output.FormatTaskGitStyle(task, nil)
		
		// Should not have metadata section for empty fields
		if strings.Contains(output, "Tags:") {
			t.Error("Output should not contain Tags: when tags are empty")
		}
		if strings.Contains(output, "Source:") {
			t.Error("Output should not contain Source: when source is empty")
		}
	})

	t.Run("very long title", func(t *testing.T) {
		longTitle := strings.Repeat("Very Long Title ", 20)
		task := createTestTask("long123", longTitle)
		
		output := output.FormatTaskOneline(task)
		
		// Should contain the full title
		if !strings.Contains(output, longTitle) {
			t.Error("Long title should not be truncated")
		}
	})

	t.Run("multiline description", func(t *testing.T) {
		task := createTestTask("multi123", "Multiline Task")
		task.Description = "Line 1\nLine 2\nLine 3"
		
		output := output.FormatTaskGitStyle(task, nil)
		
		// Each line should be indented
		for i := 1; i <= 3; i++ {
			expected := fmt.Sprintf("    Line %d", i)
			if !strings.Contains(output, expected) {
				t.Errorf("Missing properly indented line: %s", expected)
			}
		}
	})
}

func TestFormatterWriteError(t *testing.T) {
	// Test formatter behavior when writer returns an error
	errorWriter := &errorWriter{err: fmt.Errorf("write error")}
	formatter := output.NewFormatter(errorWriter)
	
	task := createTestTask("error123", "Error Task")
	
	err := formatter.FormatTask(task, nil)
	if err == nil {
		t.Error("Expected error from formatter when writer fails")
	}
}

// errorWriter is a writer that always returns an error
type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

// TestAllTaskKinds tests formatting for all task kinds
func TestAllTaskKinds(t *testing.T) {
	kinds := []string{
		models.KindBug,
		models.KindFeature,
		models.KindRegression,
	}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			task := createTestTask("kind123", fmt.Sprintf("%s Task", kind))
			task.Kind = kind
			
			output := output.FormatTaskOneline(task)
			
			// Should contain lowercase kind
			if !strings.Contains(output, strings.ToLower(kind)) {
				t.Errorf("Output missing kind %s: %s", kind, output)
			}
		})
	}
}

// TestAllPriorities tests formatting for all priority levels
func TestAllPriorities(t *testing.T) {
	priorities := []string{
		models.PriorityHigh,
		models.PriorityMedium,
		models.PriorityLow,
	}

	for _, priority := range priorities {
		t.Run(priority, func(t *testing.T) {
			task := createTestTask("pri123", fmt.Sprintf("%s Priority Task", priority))
			task.Priority = priority
			
			output := output.FormatTaskOneline(task)
			
			// Should contain priority in parentheses
			expected := fmt.Sprintf("(%s)", priority)
			if !strings.Contains(output, expected) {
				t.Errorf("Output missing priority %s: %s", priority, output)
			}
		})
	}
}