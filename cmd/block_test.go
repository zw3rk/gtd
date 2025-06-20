package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/models"
)

func TestBlockCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	// Create test tasks
	task1 := models.NewTask(models.KindBug, "Task to be blocked", "This task will be blocked by another task")
	if err := testRepo.Create(task1); err != nil {
		t.Fatal(err)
	}

	task2 := models.NewTask(models.KindFeature, "Blocking task", "This task blocks other tasks")
	if err := testRepo.Create(task2); err != nil {
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
			name: "block task successfully",
			args: []string{task1.ID, "--by", task2.ID},
			check: func(t *testing.T) {
				updated, err := testRepo.GetByID(task1.ID)
				if err != nil {
					t.Fatal(err)
				}
				if !updated.IsBlocked() {
					t.Error("Task should be blocked")
				}
				if updated.BlockedBy == nil || *updated.BlockedBy != task2.ID {
					t.Errorf("BlockedBy = %v, want %s", updated.BlockedBy, task2.ID)
				}
			},
		},
		{
			name:    "missing task ID",
			args:    []string{"--by", "1"},
			wantErr: true,
		},
		{
			name:    "missing --by flag",
			args:    []string{task1.ID},
			wantErr: true,
			errMsg:  "required flag",
		},
		{
			name:    "invalid task ID",
			args:    []string{"abc", "--by", "1"},
			wantErr: true,
			errMsg:  "task not found",
		},
		{
			name:    "invalid blocking task ID",
			args:    []string{"1", "--by", "xyz"},
			wantErr: true,
			errMsg:  "task not found",
		},
		{
			name:    "non-existent task",
			args:    []string{"999", "--by", task2.ID},
			wantErr: true,
			errMsg:  "task not found",
		},
		{
			name:    "non-existent blocking task",
			args:    []string{task1.ID, "--by", "999"},
			wantErr: true,
			errMsg:  "blocking task not found",
		},
		{
			name:    "block by itself",
			args:    []string{task1.ID, "--by", task1.ID},
			wantErr: true,
			errMsg:  "cannot block a task by itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cmd := newBlockCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)

			// Clear any existing blocking
			if err := testRepo.Unblock(task1.ID); err != nil {
				t.Errorf("Failed to unblock task for test setup: %v", err)
			}

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

func TestUnblockCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	// Create test tasks
	blockedTask := models.NewTask(models.KindBug, "Blocked task", "This task is blocked by another task")
	if err := testRepo.Create(blockedTask); err != nil {
		t.Fatal(err)
	}

	blockingTask := models.NewTask(models.KindFeature, "Blocking task", "This task blocks other tasks")
	if err := testRepo.Create(blockingTask); err != nil {
		t.Fatal(err)
	}

	// Block the task
	if err := testRepo.Block(blockedTask.ID, blockingTask.ID); err != nil {
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
			name: "unblock task successfully",
			args: []string{blockedTask.ID},
			check: func(t *testing.T) {
				updated, err := testRepo.GetByID(blockedTask.ID)
				if err != nil {
					t.Fatal(err)
				}
				if updated.IsBlocked() {
					t.Error("Task should not be blocked")
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
			errMsg:  "task not found",
		},
		{
			name:    "non-existent task",
			args:    []string{"999"},
			wantErr: true,
			errMsg:  "task not found",
		},
		{
			name: "unblock already unblocked task",
			args: []string{blockingTask.ID},
			check: func(t *testing.T) {
				// Should succeed without error
				updated, err := testRepo.GetByID(blockingTask.ID)
				if err != nil {
					t.Fatal(err)
				}
				if updated.IsBlocked() {
					t.Error("Task should not be blocked")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cmd := newUnblockCommand()
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
