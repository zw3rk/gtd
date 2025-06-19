package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/zw3rk/claude-gtd/internal/models"
)

func TestShowCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()
	
	// Create parent task
	parent := models.NewTask(models.KindFeature, "Parent feature", "This is a detailed description\nwith multiple lines")
	parent.Priority = models.PriorityHigh
	parent.State = models.StateInProgress
	parent.Source = "GitHub:issue/123"
	parent.Tags = "backend,important"
	if err := testRepo.Create(parent); err != nil {
		t.Fatal(err)
	}
	
	// Create subtasks
	subtask1 := models.NewTask(models.KindBug, "First subtask", "Fix the bug")
	subtask1.Parent = &parent.ID
	subtask1.State = models.StateDone
	if err := testRepo.Create(subtask1); err != nil {
		t.Fatal(err)
	}
	
	subtask2 := models.NewTask(models.KindBug, "Second subtask", "")
	subtask2.Parent = &parent.ID
	subtask2.Priority = models.PriorityHigh
	if err := testRepo.Create(subtask2); err != nil {
		t.Fatal(err)
	}
	
	// Create a task that blocks the parent
	blocker := models.NewTask(models.KindBug, "Blocking task", "")
	if err := testRepo.Create(blocker); err != nil {
		t.Fatal(err)
	}
	if err := testRepo.Block(parent.ID, blocker.ID); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
		contains []string
		notContains []string
	}{
		{
			name: "show parent task with subtasks",
			args: []string{fmt.Sprintf("%d", parent.ID)},
			contains: []string{
				"Parent feature",
				"Feature",
				"HIGH",
				"IN_PROGRESS",
				"This is a detailed description",
				"with multiple lines",
				"GitHub:issue/123",
				"backend,important",
				fmt.Sprintf("Blocked by: #%d", blocker.ID),
				"Subtasks:",
				"First subtask",
				"DONE",
				"Second subtask",
				"NEW",
			},
		},
		{
			name: "show simple task without subtasks",
			args: []string{fmt.Sprintf("%d", blocker.ID)},
			contains: []string{
				"Blocking task",
				"Bug",
			},
			notContains: []string{
				"Subtasks:",
			},
		},
		{
			name: "show subtask with parent info",
			args: []string{fmt.Sprintf("%d", subtask1.ID)},
			contains: []string{
				"First subtask",
				"Fix the bug",
				fmt.Sprintf("Parent: #%d", parent.ID),
				"Parent feature",
			},
		},
		{
			name:    "missing task ID",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid task ID",
			args:    []string{"abc"},
			wantErr: true,
			errMsg:  "invalid task ID",
		},
		{
			name:    "non-existent task",
			args:    []string{"999"},
			wantErr: true,
			errMsg:  "task not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			
			cmd := newShowCommand()
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
			
			output := stdout.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("Output does not contain %q\nGot: %s", want, output)
				}
			}
			
			for _, notWant := range tt.notContains {
				if strings.Contains(output, notWant) {
					t.Errorf("Output should not contain %q\nGot: %s", notWant, output)
				}
			}
		})
	}
}