package output_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/zw3rk/gtd/internal/models"
)

// JSON Formatter interface
type JSONFormatter struct {
	writer bytes.Buffer
}

func (f *JSONFormatter) FormatTask(task *models.Task) error {
	type exportTask struct {
		ID          string  `json:"id"`
		Kind        string  `json:"kind"`
		State       string  `json:"state"`
		Priority    string  `json:"priority"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Tags        string  `json:"tags"`
		Source      string  `json:"source"`
		Parent      *string `json:"parent,omitempty"`
		BlockedBy   *string `json:"blocked_by,omitempty"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
	}

	et := exportTask{
		ID:          task.ID,
		Kind:        task.Kind,
		State:       task.State,
		Priority:    task.Priority,
		Title:       task.Title,
		Description: task.Description,
		Tags:        task.Tags,
		Source:      task.Source,
		Parent:      task.Parent,
		BlockedBy:   task.BlockedBy,
		CreatedAt:   task.Created.Format("2006-01-02 15:04:05"),
		UpdatedAt:   task.Updated.Format("2006-01-02 15:04:05"),
	}

	encoder := json.NewEncoder(&f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(et)
}

func (f *JSONFormatter) FormatTasks(tasks []*models.Task) error {
	type exportTask struct {
		ID          string  `json:"id"`
		Kind        string  `json:"kind"`
		State       string  `json:"state"`
		Priority    string  `json:"priority"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Tags        string  `json:"tags"`
		Source      string  `json:"source"`
		Parent      *string `json:"parent,omitempty"`
		BlockedBy   *string `json:"blocked_by,omitempty"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
	}

	exportTasks := make([]exportTask, len(tasks))
	for i, task := range tasks {
		exportTasks[i] = exportTask{
			ID:          task.ID,
			Kind:        task.Kind,
			State:       task.State,
			Priority:    task.Priority,
			Title:       task.Title,
			Description: task.Description,
			Tags:        task.Tags,
			Source:      task.Source,
			Parent:      task.Parent,
			BlockedBy:   task.BlockedBy,
			CreatedAt:   task.Created.Format("2006-01-02 15:04:05"),
			UpdatedAt:   task.Updated.Format("2006-01-02 15:04:05"),
		}
	}

	encoder := json.NewEncoder(&f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exportTasks)
}

func (f *JSONFormatter) String() string {
	return f.writer.String()
}

// CSV Formatter
type CSVFormatter struct {
	writer bytes.Buffer
}

func (f *CSVFormatter) FormatTask(task *models.Task) error {
	csvWriter := csv.NewWriter(&f.writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Type", "State", "Priority", "Title", "Tags", "Source", "Parent", "BlockedBy", "Created", "Updated"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Write task row
	return f.writeTaskRow(csvWriter, task)
}

func (f *CSVFormatter) FormatTasks(tasks []*models.Task) error {
	csvWriter := csv.NewWriter(&f.writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Type", "State", "Priority", "Title", "Tags", "Source", "Parent", "BlockedBy", "Created", "Updated"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, task := range tasks {
		if err := f.writeTaskRow(csvWriter, task); err != nil {
			return err
		}
	}

	return nil
}

func (f *CSVFormatter) writeTaskRow(w *csv.Writer, task *models.Task) error {
	parentStr := ""
	if task.Parent != nil {
		parentStr = *task.Parent
	}

	blockedByStr := ""
	if task.BlockedBy != nil {
		blockedByStr = *task.BlockedBy
	}

	row := []string{
		task.ID,
		task.Kind,
		task.State,
		task.Priority,
		task.Title,
		task.Tags,
		task.Source,
		parentStr,
		blockedByStr,
		task.Created.Format("2006-01-02 15:04:05"),
		task.Updated.Format("2006-01-02 15:04:05"),
	}

	return w.Write(row)
}

func (f *CSVFormatter) String() string {
	return f.writer.String()
}

// Markdown Formatter
type MarkdownFormatter struct {
	writer bytes.Buffer
}

func (f *MarkdownFormatter) FormatTask(task *models.Task) error {
	fmt.Fprintf(&f.writer, "# Task: %s\n\n", task.Title)
	fmt.Fprintf(&f.writer, "**ID:** %s\n\n", task.ID)
	
	if task.Description != "" {
		fmt.Fprintf(&f.writer, "%s\n\n", task.Description)
	}
	
	fmt.Fprintf(&f.writer, "## Details\n\n")
	fmt.Fprintf(&f.writer, "- **Type:** %s\n", task.Kind)
	fmt.Fprintf(&f.writer, "- **State:** %s %s\n", task.State, getStateEmoji(task.State))
	fmt.Fprintf(&f.writer, "- **Priority:** %s %s\n", task.Priority, getPriorityEmoji(task.Priority))
	
	if task.Tags != "" {
		fmt.Fprintf(&f.writer, "- **Tags:** %s\n", task.Tags)
	}
	
	if task.Source != "" {
		fmt.Fprintf(&f.writer, "- **Source:** %s\n", task.Source)
	}
	
	if task.Parent != nil {
		fmt.Fprintf(&f.writer, "- **Parent:** #%s\n", *task.Parent)
	}
	
	if task.BlockedBy != nil {
		fmt.Fprintf(&f.writer, "- **Blocked by:** #%s\n", *task.BlockedBy)
	}
	
	fmt.Fprintf(&f.writer, "- **Created:** %s\n", task.Created.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&f.writer, "- **Updated:** %s\n", task.Updated.Format("2006-01-02 15:04:05"))
	
	return nil
}

func (f *MarkdownFormatter) FormatTasks(tasks []*models.Task) error {
	fmt.Fprintln(&f.writer, "# Tasks Export")
	fmt.Fprintln(&f.writer)
	fmt.Fprintf(&f.writer, "Total tasks: %d\n", len(tasks))
	fmt.Fprintln(&f.writer)

	// Table header
	fmt.Fprintln(&f.writer, "| ID | Type | State | Priority | Title | Tags | Source | Parent | Blocked By |")
	fmt.Fprintln(&f.writer, "|---|---|---|---|---|---|---|---|---|")

	// Table rows
	for i, task := range tasks {
		parentStr := "-"
		if task.Parent != nil {
			parentStr = fmt.Sprintf("#%s", (*task.Parent)[:7])
		}

		blockedByStr := "-"
		if task.BlockedBy != nil {
			blockedByStr = fmt.Sprintf("#%s", (*task.BlockedBy)[:7])
		}

		tagsStr := "-"
		if task.Tags != "" {
			tagsStr = task.Tags
		}

		sourceStr := "-"
		if task.Source != "" {
			sourceStr = task.Source
		}

		fmt.Fprintf(&f.writer, "| %d | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			i+1,
			task.Kind,
			task.State,
			task.Priority,
			task.Title,
			tagsStr,
			sourceStr,
			parentStr,
			blockedByStr,
		)
	}

	// Add detailed task descriptions
	fmt.Fprintln(&f.writer)
	fmt.Fprintln(&f.writer, "## Task Details")
	fmt.Fprintln(&f.writer)

	for _, task := range tasks {
		fmt.Fprintf(&f.writer, "### #%s: %s\n", task.ID, task.Title)
		fmt.Fprintln(&f.writer)

		if task.Description != "" {
			fmt.Fprintln(&f.writer, task.Description)
			fmt.Fprintln(&f.writer)
		}

		fmt.Fprintf(&f.writer, "- **Type:** %s\n", task.Kind)
		fmt.Fprintf(&f.writer, "- **State:** %s %s\n", task.State, getStateEmoji(task.State))
		fmt.Fprintf(&f.writer, "- **Priority:** %s %s\n", task.Priority, getPriorityEmoji(task.Priority))

		if task.Tags != "" {
			fmt.Fprintf(&f.writer, "- **Tags:** %s\n", task.Tags)
		}

		if task.Source != "" {
			fmt.Fprintf(&f.writer, "- **Source:** %s\n", task.Source)
		}

		if task.Parent != nil {
			fmt.Fprintf(&f.writer, "- **Parent:** #%s\n", *task.Parent)
		}

		if task.BlockedBy != nil {
			fmt.Fprintf(&f.writer, "- **Blocked by:** #%s\n", *task.BlockedBy)
		}

		fmt.Fprintf(&f.writer, "- **Created:** %s\n", task.Created.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&f.writer, "- **Updated:** %s\n", task.Updated.Format("2006-01-02 15:04:05"))
		fmt.Fprintln(&f.writer)
	}

	return nil
}

func (f *MarkdownFormatter) String() string {
	return f.writer.String()
}

// Helper functions
func getStateEmoji(state string) string {
	switch state {
	case models.StateNew:
		return "◆"
	case models.StateInProgress:
		return "▶"
	case models.StateDone:
		return "✓"
	case models.StateCancelled:
		return "✗"
	default:
		return "?"
	}
}

func getPriorityEmoji(priority string) string {
	switch priority {
	case models.PriorityHigh:
		return "!"
	case models.PriorityMedium:
		return "="
	case models.PriorityLow:
		return "-"
	default:
		return "."
	}
}

// Tests
func TestJSONFormatter(t *testing.T) {
	task := createTestTask("json123", "JSON Test Task")
	
	t.Run("FormatTask", func(t *testing.T) {
		formatter := &JSONFormatter{}
		err := formatter.FormatTask(task)
		if err != nil {
			t.Fatalf("FormatTask failed: %v", err)
		}
		
		output := formatter.String()
		
		// Verify JSON structure
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}
		
		// Check required fields
		if result["id"] != task.ID {
			t.Errorf("Expected ID %s, got %v", task.ID, result["id"])
		}
		if result["title"] != task.Title {
			t.Errorf("Expected title %s, got %v", task.Title, result["title"])
		}
		if result["kind"] != task.Kind {
			t.Errorf("Expected kind %s, got %v", task.Kind, result["kind"])
		}
		if result["state"] != task.State {
			t.Errorf("Expected state %s, got %v", task.State, result["state"])
		}
		if result["priority"] != task.Priority {
			t.Errorf("Expected priority %s, got %v", task.Priority, result["priority"])
		}
	})
	
	t.Run("FormatTasks", func(t *testing.T) {
		tasks := []*models.Task{
			createTestTask("json1", "First JSON Task"),
			createTestTask("json2", "Second JSON Task"),
			createTestTask("json3", "Third JSON Task"),
		}
		
		formatter := &JSONFormatter{}
		err := formatter.FormatTasks(tasks)
		if err != nil {
			t.Fatalf("FormatTasks failed: %v", err)
		}
		
		output := formatter.String()
		
		// Verify JSON array
		var result []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Invalid JSON array output: %v", err)
		}
		
		if len(result) != len(tasks) {
			t.Errorf("Expected %d tasks in JSON, got %d", len(tasks), len(result))
		}
		
		// Check each task
		for i, taskData := range result {
			if taskData["id"] != tasks[i].ID {
				t.Errorf("Task %d: Expected ID %s, got %v", i, tasks[i].ID, taskData["id"])
			}
			if taskData["title"] != tasks[i].Title {
				t.Errorf("Task %d: Expected title %s, got %v", i, tasks[i].Title, taskData["title"])
			}
		}
	})
	
	t.Run("FormatTask with optional fields", func(t *testing.T) {
		task := createTestTask("json456", "Task with Options")
		parent := "parent789"
		blocker := "blocker012"
		task.Parent = &parent
		task.BlockedBy = &blocker
		
		formatter := &JSONFormatter{}
		err := formatter.FormatTask(task)
		if err != nil {
			t.Fatalf("FormatTask failed: %v", err)
		}
		
		output := formatter.String()
		
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}
		
		if result["parent"] != parent {
			t.Errorf("Expected parent %s, got %v", parent, result["parent"])
		}
		if result["blocked_by"] != blocker {
			t.Errorf("Expected blocked_by %s, got %v", blocker, result["blocked_by"])
		}
	})
}

func TestCSVFormatter(t *testing.T) {
	task := createTestTask("csv123", "CSV Test Task")
	
	t.Run("FormatTask", func(t *testing.T) {
		formatter := &CSVFormatter{}
		err := formatter.FormatTask(task)
		if err != nil {
			t.Fatalf("FormatTask failed: %v", err)
		}
		
		output := formatter.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Should have header + 1 data row
		if len(lines) != 2 {
			t.Errorf("Expected 2 lines (header + data), got %d", len(lines))
		}
		
		// Verify header
		expectedHeader := "ID,Type,State,Priority,Title,Tags,Source,Parent,BlockedBy,Created,Updated"
		if lines[0] != expectedHeader {
			t.Errorf("Expected header:\n%s\nGot:\n%s", expectedHeader, lines[0])
		}
		
		// Verify data contains task info
		if !strings.Contains(lines[1], task.ID) {
			t.Errorf("Data row missing task ID")
		}
		if !strings.Contains(lines[1], task.Title) {
			t.Errorf("Data row missing task title")
		}
	})
	
	t.Run("FormatTasks", func(t *testing.T) {
		tasks := []*models.Task{
			createTestTask("csv1", "First CSV Task"),
			createTestTask("csv2", "Second CSV Task"),
			createTestTask("csv3", "Third CSV Task"),
		}
		
		formatter := &CSVFormatter{}
		err := formatter.FormatTasks(tasks)
		if err != nil {
			t.Fatalf("FormatTasks failed: %v", err)
		}
		
		output := formatter.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Should have header + 3 data rows
		if len(lines) != 4 {
			t.Errorf("Expected 4 lines (header + 3 data), got %d", len(lines))
		}
		
		// Verify each task appears
		for _, task := range tasks {
			found := false
			for _, line := range lines[1:] { // Skip header
				if strings.Contains(line, task.ID) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Task %s not found in CSV output", task.ID)
			}
		}
	})
	
	t.Run("CSV escaping", func(t *testing.T) {
		task := createTestTask("csvescape", "Task with, comma and \"quotes\"")
		task.Description = "Line 1\nLine 2"
		
		formatter := &CSVFormatter{}
		err := formatter.FormatTask(task)
		if err != nil {
			t.Fatalf("FormatTask failed: %v", err)
		}
		
		output := formatter.String()
		
		// CSV should properly escape special characters
		if !strings.Contains(output, "\"Task with, comma and \"\"quotes\"\"\"") {
			t.Error("CSV did not properly escape title with comma and quotes")
		}
	})
}

func TestMarkdownFormatter(t *testing.T) {
	task := createTestTask("md123", "Markdown Test Task")
	
	t.Run("FormatTask", func(t *testing.T) {
		formatter := &MarkdownFormatter{}
		err := formatter.FormatTask(task)
		if err != nil {
			t.Fatalf("FormatTask failed: %v", err)
		}
		
		output := formatter.String()
		
		// Check markdown structure
		expectedElements := []string{
			"# Task: Markdown Test Task",
			"**ID:** md123",
			"## Details",
			"- **Type:** " + task.Kind,
			"- **State:** " + task.State,
			"- **Priority:** " + task.Priority,
			"- **Tags:** " + task.Tags,
			"- **Source:** " + task.Source,
		}
		
		for _, expected := range expectedElements {
			if !strings.Contains(output, expected) {
				t.Errorf("Missing expected element: %s", expected)
			}
		}
	})
	
	t.Run("FormatTasks", func(t *testing.T) {
		tasks := []*models.Task{
			createTestTask("md1", "First MD Task"),
			createTestTask("md2", "Second MD Task"),
			createTestTask("md3", "Third MD Task"),
		}
		
		formatter := &MarkdownFormatter{}
		err := formatter.FormatTasks(tasks)
		if err != nil {
			t.Fatalf("FormatTasks failed: %v", err)
		}
		
		output := formatter.String()
		
		// Check for header
		if !strings.Contains(output, "# Tasks Export") {
			t.Error("Missing main header")
		}
		
		// Check for total count
		if !strings.Contains(output, "Total tasks: 3") {
			t.Error("Missing total tasks count")
		}
		
		// Check for table
		if !strings.Contains(output, "| ID | Type | State | Priority | Title | Tags | Source | Parent | Blocked By |") {
			t.Error("Missing table header")
		}
		
		// Check for task details section
		if !strings.Contains(output, "## Task Details") {
			t.Error("Missing task details section")
		}
		
		// Check each task appears in details
		for _, task := range tasks {
			expectedHeader := fmt.Sprintf("### #%s: %s", task.ID, task.Title)
			if !strings.Contains(output, expectedHeader) {
				t.Errorf("Missing task detail header: %s", expectedHeader)
			}
		}
	})
	
	t.Run("Markdown with special characters", func(t *testing.T) {
		task := createTestTask("mdspecial", "Task with **bold** and _italic_ text")
		task.Description = "Description with [link](http://example.com) and `code`"
		
		formatter := &MarkdownFormatter{}
		err := formatter.FormatTask(task)
		if err != nil {
			t.Fatalf("FormatTask failed: %v", err)
		}
		
		output := formatter.String()
		
		// Markdown special characters should be preserved
		if !strings.Contains(output, "**bold**") {
			t.Error("Markdown bold syntax not preserved")
		}
		if !strings.Contains(output, "_italic_") {
			t.Error("Markdown italic syntax not preserved")
		}
		if !strings.Contains(output, "[link](http://example.com)") {
			t.Error("Markdown link syntax not preserved")
		}
		if !strings.Contains(output, "`code`") {
			t.Error("Markdown code syntax not preserved")
		}
	})
}

// Test empty/nil cases for all formatters
func TestFormattersEdgeCases(t *testing.T) {
	t.Run("Empty task list", func(t *testing.T) {
		type formatter interface {
			FormatTasks(tasks []*models.Task) error
			String() string
		}
		
		formatters := []formatter{
			&JSONFormatter{},
			&CSVFormatter{},
			&MarkdownFormatter{},
		}
		
		for _, formatter := range formatters {
			err := formatter.FormatTasks([]*models.Task{})
			if err != nil {
				t.Errorf("%T: FormatTasks with empty list failed: %v", formatter, err)
			}
			
			output := formatter.String()
			if output == "" {
				t.Errorf("%T: Expected some output for empty list", formatter)
			}
		}
	})
	
	t.Run("Task with nil optional fields", func(t *testing.T) {
		task := &models.Task{
			ID:       "nil123",
			Kind:     models.KindBug,
			State:    models.StateNew,
			Priority: models.PriorityHigh,
			Title:    "Task with nil fields",
			Author:   "Test",
			Created:  time.Now(),
			Updated:  time.Now(),
			// Parent and BlockedBy are nil
		}
		
		// Test JSON
		jsonFormatter := &JSONFormatter{}
		if err := jsonFormatter.FormatTask(task); err != nil {
			t.Errorf("JSON FormatTask failed with nil fields: %v", err)
		}
		
		// Test CSV
		csvFormatter := &CSVFormatter{}
		if err := csvFormatter.FormatTask(task); err != nil {
			t.Errorf("CSV FormatTask failed with nil fields: %v", err)
		}
		
		// Test Markdown
		mdFormatter := &MarkdownFormatter{}
		if err := mdFormatter.FormatTask(task); err != nil {
			t.Errorf("Markdown FormatTask failed with nil fields: %v", err)
		}
	})
}