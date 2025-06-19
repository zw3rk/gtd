package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/models"
)

func TestInProgressCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()
	
	// Create test tasks
	newTask := models.NewTask(models.KindBug, "New bug", "")
	if err := testRepo.Create(newTask); err != nil {
		t.Fatal(err)
	}
	
	doneTask := models.NewTask(models.KindFeature, "Done feature", "")
	doneTask.State = models.StateDone
	if err := testRepo.Create(doneTask); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
		check   func(t *testing.T)
	}{
		{
			name: "mark new task as in progress",
			args: []string{fmt.Sprintf("%d", newTask.ID)},
			check: func(t *testing.T) {
				task, err := testRepo.GetByID(newTask.ID)
				if err != nil {
					t.Fatal(err)
				}
				if task.State != models.StateInProgress {
					t.Errorf("State = %q, want %q", task.State, models.StateInProgress)
				}
			},
		},
		{
			name: "mark done task as in progress",
			args: []string{fmt.Sprintf("%d", doneTask.ID)},
			check: func(t *testing.T) {
				task, err := testRepo.GetByID(doneTask.ID)
				if err != nil {
					t.Fatal(err)
				}
				if task.State != models.StateInProgress {
					t.Errorf("State = %q, want %q", task.State, models.StateInProgress)
				}
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
			
			cmd := newInProgressCommand()
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
			
			if !tt.wantErr && tt.check != nil {
				tt.check(t)
			}
		})
	}
}

func TestDoneCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()
	
	// Create parent and child tasks
	parent := models.NewTask(models.KindFeature, "Parent feature", "")
	parent.State = models.StateInProgress
	if err := testRepo.Create(parent); err != nil {
		t.Fatal(err)
	}
	
	child1 := models.NewTask(models.KindBug, "Child bug 1", "")
	child1.Parent = &parent.ID
	child1.State = models.StateInProgress
	if err := testRepo.Create(child1); err != nil {
		t.Fatal(err)
	}
	
	child2 := models.NewTask(models.KindBug, "Child bug 2", "")
	child2.Parent = &parent.ID
	child2.State = models.StateNew
	if err := testRepo.Create(child2); err != nil {
		t.Fatal(err)
	}
	
	simpleTask := models.NewTask(models.KindBug, "Simple task", "")
	simpleTask.State = models.StateInProgress
	if err := testRepo.Create(simpleTask); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name    string
		args    []string
		setup   func()
		wantErr bool
		errMsg  string
		check   func(t *testing.T)
	}{
		{
			name: "mark simple task as done",
			args: []string{fmt.Sprintf("%d", simpleTask.ID)},
			check: func(t *testing.T) {
				task, err := testRepo.GetByID(simpleTask.ID)
				if err != nil {
					t.Fatal(err)
				}
				if task.State != models.StateDone {
					t.Errorf("State = %q, want %q", task.State, models.StateDone)
				}
			},
		},
		{
			name:    "cannot mark parent as done with incomplete children",
			args:    []string{fmt.Sprintf("%d", parent.ID)},
			wantErr: true,
			errMsg:  "child task",
		},
		{
			name: "mark parent as done after children are complete",
			args: []string{fmt.Sprintf("%d", parent.ID)},
			setup: func() {
				// Mark children as done
				testRepo.UpdateState(child1.ID, models.StateDone)
				testRepo.UpdateState(child2.ID, models.StateDone)
			},
			check: func(t *testing.T) {
				task, err := testRepo.GetByID(parent.ID)
				if err != nil {
					t.Fatal(err)
				}
				if task.State != models.StateDone {
					t.Errorf("State = %q, want %q", task.State, models.StateDone)
				}
			},
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
			if tt.setup != nil {
				tt.setup()
			}
			
			var stdout, stderr bytes.Buffer
			
			cmd := newDoneCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("stderr: %s", stderr.String())
				return
			}
			
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Error = %v, want error containing %q", err, tt.errMsg)
			}
			
			if !tt.wantErr && tt.check != nil {
				tt.check(t)
			}
		})
	}
}

func TestCancelCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()
	
	// Create test task
	task := models.NewTask(models.KindBug, "Bug to cancel", "")
	task.State = models.StateInProgress
	if err := testRepo.Create(task); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(t *testing.T)
	}{
		{
			name: "cancel in-progress task",
			args: []string{fmt.Sprintf("%d", task.ID)},
			check: func(t *testing.T) {
				updated, err := testRepo.GetByID(task.ID)
				if err != nil {
					t.Fatal(err)
				}
				if updated.State != models.StateCancelled {
					t.Errorf("State = %q, want %q", updated.State, models.StateCancelled)
				}
			},
		},
		{
			name:    "non-existent task",
			args:    []string{"999"},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			
			cmd := newCancelCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.check != nil {
				tt.check(t)
			}
		})
	}
}