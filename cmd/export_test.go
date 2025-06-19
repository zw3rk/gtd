package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/zw3rk/claude-gtd/internal/models"
)

func TestExportCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()
	
	// Create test tasks
	task1 := models.NewTask(models.KindBug, "Bug 1", "Description 1")
	task1.Priority = models.PriorityHigh
	task1.Tags = "tag1,tag2"
	task1.Source = "GitHub:issue/123"
	if err := testRepo.Create(task1); err != nil {
		t.Fatal(err)
	}
	
	task2 := models.NewTask(models.KindFeature, "Feature 1", "Description 2")
	task2.Priority = models.PriorityMedium
	task2.State = models.StateInProgress
	if err := testRepo.Create(task2); err != nil {
		t.Fatal(err)
	}
	
	// Create parent-child relationship
	task3 := models.NewTask(models.KindRegression, "Subtask 1", "Subtask description")
	task3.Parent = &task1.ID
	task3.State = models.StateDone
	if err := testRepo.Create(task3); err != nil {
		t.Fatal(err)
	}
	
	// Block task2 by task1
	if err := testRepo.Block(task2.ID, task1.ID); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, output string)
	}{
		{
			name: "export to JSON",
			args: []string{"--format", "json"},
			validate: func(t *testing.T, output string) {
				var tasks []map[string]interface{}
				if err := json.Unmarshal([]byte(output), &tasks); err != nil {
					t.Errorf("Failed to parse JSON output: %v", err)
					return
				}
				
				if len(tasks) != 3 {
					t.Errorf("Expected 3 tasks, got %d", len(tasks))
				}
				
				// Check that we have the expected tasks
				foundBug := false
				for _, task := range tasks {
					if task["title"] == "Bug 1" && task["kind"] == "BUG" {
						foundBug = true
						break
					}
				}
				if !foundBug {
					t.Error("Expected to find Bug 1 with kind BUG")
				}
			},
		},
		{
			name: "export to CSV",
			args: []string{"--format", "csv"},
			validate: func(t *testing.T, output string) {
				reader := csv.NewReader(strings.NewReader(output))
				records, err := reader.ReadAll()
				if err != nil {
					t.Errorf("Failed to parse CSV output: %v", err)
					return
				}
				
				// Check header
				if len(records) < 1 {
					t.Errorf("No records in CSV output")
					return
				}
				
				header := records[0]
				expectedHeaders := []string{"ID", "Type", "State", "Priority", "Title", "Tags", "Source", "Parent", "BlockedBy", "Created", "Updated"}
				if len(header) != len(expectedHeaders) {
					t.Errorf("Expected %d columns, got %d", len(expectedHeaders), len(header))
				}
				
				// Check data rows
				if len(records) != 4 { // header + 3 tasks
					t.Errorf("Expected 4 rows (header + 3 tasks), got %d", len(records))
				}
			},
		},
		{
			name: "export to Markdown",
			args: []string{"--format", "markdown"},
			validate: func(t *testing.T, output string) {
				// Check for markdown table elements
				if !strings.Contains(output, "# Tasks Export") {
					t.Error("Missing markdown header")
				}
				if !strings.Contains(output, "| ID | Type | State | Priority | Title |") {
					t.Error("Missing table header")
				}
				if !strings.Contains(output, "|---|---|---|---|---|") {
					t.Error("Missing table separator")
				}
				if !strings.Contains(output, "| 1 | BUG | NEW | high | Bug 1 |") {
					t.Error("Missing task row")
				}
			},
		},
		{
			name: "export active tasks only",
			args: []string{"--format", "json", "--active"},
			validate: func(t *testing.T, output string) {
				var tasks []map[string]interface{}
				if err := json.Unmarshal([]byte(output), &tasks); err != nil {
					t.Errorf("Failed to parse JSON output: %v", err)
					return
				}
				
				// Should only have 2 tasks (exclude DONE task)
				if len(tasks) != 2 {
					t.Errorf("Expected 2 active tasks, got %d", len(tasks))
				}
				
				// Check no DONE tasks
				for _, task := range tasks {
					if task["state"] == "DONE" {
						t.Error("Found DONE task in active export")
					}
				}
			},
		},
		{
			name: "export with state filter",
			args: []string{"--format", "json", "--state", "done"},
			validate: func(t *testing.T, output string) {
				var tasks []map[string]interface{}
				if err := json.Unmarshal([]byte(output), &tasks); err != nil {
					t.Errorf("Failed to parse JSON output: %v", err)
					return
				}
				
				// Should only have 1 task
				if len(tasks) != 1 {
					t.Errorf("Expected 1 done task, got %d", len(tasks))
				}
				
				if tasks[0]["state"] != "DONE" {
					t.Errorf("Expected DONE state, got %v", tasks[0]["state"])
				}
			},
		},
		{
			name: "export with kind filter",
			args: []string{"--format", "json", "--kind", "bug"},
			validate: func(t *testing.T, output string) {
				var tasks []map[string]interface{}
				if err := json.Unmarshal([]byte(output), &tasks); err != nil {
					t.Errorf("Failed to parse JSON output: %v", err)
					return
				}
				
				// Should only have bug tasks
				if len(tasks) != 1 {
					t.Errorf("Expected 1 bug task, got %d", len(tasks))
				}
				
				if tasks[0]["kind"] != "BUG" {
					t.Errorf("Expected BUG kind, got %v", tasks[0]["kind"])
				}
			},
		},
		{
			name: "default format is JSON",
			args: []string{},
			validate: func(t *testing.T, output string) {
				var tasks []map[string]interface{}
				if err := json.Unmarshal([]byte(output), &tasks); err != nil {
					t.Errorf("Failed to parse JSON output: %v", err)
				}
			},
		},
		{
			name: "invalid format",
			args: []string{"--format", "xml"},
			wantErr: true,
			errMsg: "unsupported format",
		},
		{
			name: "export to file",
			args: []string{"--format", "json", "--output", "/tmp/tasks.json"},
			validate: func(t *testing.T, output string) {
				// Should show success message
				if !strings.Contains(output, "Exported") || !strings.Contains(output, "/tmp/tasks.json") {
					t.Error("Missing export success message")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			
			cmd := newExportCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error = %v, want error containing %q", err, tt.errMsg)
			}
			
			if tt.validate != nil {
				tt.validate(t, stdout.String())
			}
		})
	}
}