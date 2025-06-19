package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/models"
)

func TestSummaryCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()
	
	// Create a variety of tasks for statistics
	tasks := []struct {
		kind     string
		state    string
		priority string
		blocked  bool
	}{
		// Bugs
		{models.KindBug, models.StateNew, models.PriorityHigh, false},
		{models.KindBug, models.StateNew, models.PriorityMedium, true},
		{models.KindBug, models.StateInProgress, models.PriorityHigh, false},
		{models.KindBug, models.StateDone, models.PriorityLow, false},
		{models.KindBug, models.StateDone, models.PriorityMedium, false},
		{models.KindBug, models.StateCancelled, models.PriorityLow, false},
		
		// Features
		{models.KindFeature, models.StateNew, models.PriorityHigh, true},
		{models.KindFeature, models.StateInProgress, models.PriorityMedium, false},
		{models.KindFeature, models.StateInProgress, models.PriorityHigh, false},
		{models.KindFeature, models.StateDone, models.PriorityMedium, false},
		
		// Regressions
		{models.KindRegression, models.StateNew, models.PriorityHigh, false},
		{models.KindRegression, models.StateInProgress, models.PriorityHigh, false},
	}
	
	var blocker *models.Task
	for i, tt := range tasks {
		task := models.NewTask(tt.kind, fmt.Sprintf("Task %d", i+1), "")
		task.State = tt.state
		task.Priority = tt.priority
		if err := testRepo.Create(task); err != nil {
			t.Fatal(err)
		}
		
		// Use first task as blocker for others
		if i == 0 {
			blocker = task
		} else if tt.blocked && blocker != nil {
			if err := testRepo.Block(task.ID, blocker.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	
	// Create parent-child relationship
	parent := models.NewTask(models.KindFeature, "Parent task", "")
	if err := testRepo.Create(parent); err != nil {
		t.Fatal(err)
	}
	
	child := models.NewTask(models.KindBug, "Child task", "")
	child.Parent = &parent.ID
	if err := testRepo.Create(child); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "show summary statistics",
			args: []string{},
			contains: []string{
				"Task Summary",
				"Total Tasks: 14",
				
				// By State
				"By State:",
				"NEW:         6",
				"IN_PROGRESS: 4", 
				"DONE:        3",
				"CANCELLED:   1",
				
				// By Type
				"By Type:",
				"Bug:         7",
				"Feature:     5",
				"Regression:  2",
				
				// By Priority
				"By Priority:",
				"High:        6",
				"Medium:      6",
				"Low:         2",
				
				// Special categories
				"Special:",
				"Blocked:      2",
				"Parent tasks: 1",
				"Subtasks:     1",
			},
		},
		{
			name: "show active tasks summary",
			args: []string{"--active"},
			contains: []string{
				"Active Tasks: 10",
				"NEW:         6",
				"IN_PROGRESS: 4",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			
			cmd := newSummaryCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}
			
			output := stdout.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("Output does not contain %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

