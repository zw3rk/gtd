package output_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/models"
	"github.com/zw3rk/gtd/internal/output"
)

// FormatterFactory simulates a GetFormatter function
type FormatterFactory struct{}

type Format string

const (
	FormatStandard Format = "standard"
	FormatOneline  Format = "oneline"
	FormatJSON     Format = "json"
	FormatCSV      Format = "csv"
	FormatMarkdown Format = "markdown"
)

// TaskFormatter interface that all formatters implement
type TaskFormatter interface {
	FormatTask(task *models.Task) error
	FormatTasks(tasks []*models.Task) error
	String() string
}

// GetFormatter returns the appropriate formatter based on format string
func (f *FormatterFactory) GetFormatter(format string) (TaskFormatter, error) {
	switch Format(strings.ToLower(format)) {
	case FormatJSON:
		return &JSONFormatter{}, nil
	case FormatCSV:
		return &CSVFormatter{}, nil
	case FormatMarkdown:
		return &MarkdownFormatter{}, nil
	case FormatStandard, FormatOneline, "":
		// For standard/oneline, we'd return a different formatter
		// but for testing we'll use a simple one
		return &StandardFormatter{oneline: format == string(FormatOneline)}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// StandardFormatter handles standard and oneline formats
type StandardFormatter struct {
	writer  bytes.Buffer
	oneline bool
}

func (f *StandardFormatter) FormatTask(task *models.Task) error {
	if f.oneline {
		fmt.Fprintln(&f.writer, output.FormatTaskOneline(task))
	} else {
		fmt.Fprint(&f.writer, output.FormatTaskGitStyle(task, nil))
	}
	return nil
}

func (f *StandardFormatter) FormatTasks(tasks []*models.Task) error {
	for i, task := range tasks {
		if i > 0 && !f.oneline {
			fmt.Fprintln(&f.writer) // Empty line between tasks
		}
		if err := f.FormatTask(task); err != nil {
			return err
		}
	}
	return nil
}

func (f *StandardFormatter) String() string {
	return f.writer.String()
}

// Tests
func TestGetFormatter(t *testing.T) {
	factory := &FormatterFactory{}
	
	tests := []struct {
		format      string
		expectError bool
		formatType  string
	}{
		{"json", false, "JSON"},
		{"JSON", false, "JSON"},
		{"csv", false, "CSV"},
		{"CSV", false, "CSV"},
		{"markdown", false, "Markdown"},
		{"MARKDOWN", false, "Markdown"},
		{"standard", false, "Standard"},
		{"oneline", false, "Standard"},
		{"", false, "Standard"}, // Default
		{"unknown", true, ""},
		{"xml", true, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			formatter, err := factory.GetFormatter(tt.format)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for format %s, got none", tt.format)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for format %s: %v", tt.format, err)
				return
			}
			
			if formatter == nil {
				t.Errorf("GetFormatter returned nil for format %s", tt.format)
			}
			
			// Test that the formatter works
			task := createTestTask("test123", "Test Task")
			if err := formatter.FormatTask(task); err != nil {
				t.Errorf("FormatTask failed for %s formatter: %v", tt.format, err)
			}
		})
	}
}

func TestFormatterFactoryIntegration(t *testing.T) {
	factory := &FormatterFactory{}
	
	// Create test tasks
	tasks := []*models.Task{
		createTestTask("task1", "First Task"),
		createTestTask("task2", "Second Task"),
		createTestTask("task3", "Third Task"),
	}
	
	// Test each formatter type
	formats := []string{"json", "csv", "markdown", "standard", "oneline"}
	
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			formatter, err := factory.GetFormatter(format)
			if err != nil {
				t.Fatalf("Failed to get %s formatter: %v", format, err)
			}
			
			// Format tasks
			err = formatter.FormatTasks(tasks)
			if err != nil {
				t.Fatalf("FormatTasks failed for %s: %v", format, err)
			}
			
			output := formatter.String()
			if output == "" {
				t.Errorf("Empty output from %s formatter", format)
			}
			
			// Verify all tasks appear in output
			for _, task := range tasks {
				if !strings.Contains(output, task.ID) && !strings.Contains(output, task.ShortHash()) {
					t.Errorf("%s formatter output missing task %s", format, task.ID)
				}
			}
		})
	}
}

func TestFormatterConsistency(t *testing.T) {
	factory := &FormatterFactory{}
	task := createTestTask("consist123", "Consistency Test")
	
	// Get output from each formatter
	outputs := make(map[string]string)
	formats := []string{"json", "csv", "markdown", "standard", "oneline"}
	
	for _, format := range formats {
		formatter, err := factory.GetFormatter(format)
		if err != nil {
			t.Fatalf("Failed to get %s formatter: %v", format, err)
		}
		
		err = formatter.FormatTask(task)
		if err != nil {
			t.Fatalf("FormatTask failed for %s: %v", format, err)
		}
		
		outputs[format] = formatter.String()
	}
	
	// Verify each formatter produced unique output
	seen := make(map[string]string)
	for format, output := range outputs {
		if prevFormat, exists := seen[output]; exists {
			t.Errorf("Formatters %s and %s produced identical output", format, prevFormat)
		}
		seen[output] = format
	}
	
	// Verify essential information is in all outputs
	// Note: Some formatters use lowercase for kind/state
	essentials := []string{
		task.Title,
		task.Priority,
	}
	
	for format, output := range outputs {
		for _, essential := range essentials {
			if !strings.Contains(output, essential) {
				t.Errorf("%s formatter missing essential info: %s", format, essential)
			}
		}
		
		// Check for kind and state (might be lowercase)
		if !strings.Contains(output, task.Kind) && !strings.Contains(output, strings.ToLower(task.Kind)) {
			t.Errorf("%s formatter missing task kind", format)
		}
		// State might be represented as an icon (◆ for NEW) instead of text
		stateIcon := getStateEmoji(task.State)
		if !strings.Contains(output, task.State) && !strings.Contains(output, strings.ToLower(task.State)) && !strings.Contains(output, stateIcon) {
			t.Errorf("%s formatter missing task state (looked for %s or %s or icon %s)", format, task.State, strings.ToLower(task.State), stateIcon)
		}
	}
}

// Test case sensitivity
func TestFormatterCaseInsensitive(t *testing.T) {
	factory := &FormatterFactory{}
	
	formats := [][]string{
		{"json", "JSON", "Json", "jSoN"},
		{"csv", "CSV", "Csv", "cSv"},
		{"markdown", "MARKDOWN", "Markdown", "MarkDown"},
	}
	
	for _, variations := range formats {
		baseFormat := variations[0]
		
		// Get formatter with base format
		_, err := factory.GetFormatter(baseFormat)
		if err != nil {
			t.Fatalf("Failed to get %s formatter: %v", baseFormat, err)
		}
		
		// Verify all variations return a formatter
		for _, variation := range variations[1:] {
			formatter, err := factory.GetFormatter(variation)
			if err != nil {
				t.Errorf("Failed to get formatter for %s (variation of %s): %v", 
					variation, baseFormat, err)
			}
			if formatter == nil {
				t.Errorf("Got nil formatter for %s", variation)
			}
		}
	}
}

// Test formatter reusability
func TestFormatterReusability(t *testing.T) {
	factory := &FormatterFactory{}
	
	formatter, err := factory.GetFormatter("json")
	if err != nil {
		t.Fatalf("Failed to get JSON formatter: %v", err)
	}
	
	// Use formatter multiple times
	task1 := createTestTask("reuse1", "First Use")
	task2 := createTestTask("reuse2", "Second Use")
	
	// First use
	err = formatter.FormatTask(task1)
	if err != nil {
		t.Fatalf("First FormatTask failed: %v", err)
	}
	output1 := formatter.String()
	
	// Create new formatter for second use (simulating fresh state)
	formatter2, _ := factory.GetFormatter("json")
	err = formatter2.FormatTask(task2)
	if err != nil {
		t.Fatalf("Second FormatTask failed: %v", err)
	}
	output2 := formatter2.String()
	
	// Outputs should be different (different tasks)
	if output1 == output2 {
		t.Error("Different tasks produced identical output")
	}
	
	// Both should be valid JSON
	if !strings.Contains(output1, task1.ID) {
		t.Error("First output missing task1 ID")
	}
	if !strings.Contains(output2, task2.ID) {
		t.Error("Second output missing task2 ID")
	}
}

// Test error propagation
func TestFormatterErrorHandling(t *testing.T) {
	factory := &FormatterFactory{}
	
	// Test with invalid format
	_, err := factory.GetFormatter("invalid-format")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Error message should mention unsupported format, got: %v", err)
	}
	
	// Test with empty format (should default to standard)
	formatter, err := factory.GetFormatter("")
	if err != nil {
		t.Errorf("Empty format should default to standard, got error: %v", err)
	}
	if formatter == nil {
		t.Error("Empty format should return a valid formatter")
	}
}

// Benchmark different formatters
func BenchmarkFormatters(b *testing.B) {
	factory := &FormatterFactory{}
	
	// Create a substantial task
	task := createTestTask("bench123", "Benchmark Task")
	task.Description = strings.Repeat("This is a long description line.\n", 10)
	task.Tags = "tag1,tag2,tag3,tag4,tag5"
	
	formats := []string{"json", "csv", "markdown", "standard", "oneline"}
	
	for _, format := range formats {
		b.Run(format, func(b *testing.B) {
			_, err := factory.GetFormatter(format)
			if err != nil {
				b.Fatalf("Failed to get %s formatter: %v", format, err)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create new formatter each time to avoid accumulation
				f, _ := factory.GetFormatter(format)
				_ = f.FormatTask(task)
				_ = f.String()
			}
		})
	}
}