package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/models"
)

func TestAddSubtaskCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	// Create parent tasks
	parentBug := models.NewTask(models.KindBug, "Parent bug", "")
	if err := testRepo.Create(parentBug); err != nil {
		t.Fatal(err)
	}

	parentFeature := models.NewTask(models.KindFeature, "Parent feature", "")
	if err := testRepo.Create(parentFeature); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    []string
		input   string
		wantErr bool
		errMsg  string
		check   func(t *testing.T)
	}{
		{
			name:  "create bug subtask",
			args:  []string{fmt.Sprintf("%s", parentBug.ID), "--kind", "bug"},
			input: "Fix memory leak in auth module\n",
			check: func(t *testing.T) {
				children, err := testRepo.GetChildren(parentBug.ID)
				if err != nil {
					t.Fatal(err)
				}
				if len(children) != 1 {
					t.Fatalf("Expected 1 child, got %d", len(children))
				}
				child := children[0]
				if child.Title != "Fix memory leak in auth module" {
					t.Errorf("Title = %q", child.Title)
				}
				if child.Kind != models.KindBug {
					t.Errorf("Kind = %q, want %q", child.Kind, models.KindBug)
				}
				if child.Parent == nil || *child.Parent != parentBug.ID {
					t.Error("Parent not set correctly")
				}
			},
		},
		{
			name:  "create feature subtask with priority",
			args:  []string{fmt.Sprintf("%s", parentFeature.ID), "--kind", "feature", "--priority", "high"},
			input: "Add theme switcher\nImplement UI for theme selection",
			check: func(t *testing.T) {
				children, err := testRepo.GetChildren(parentFeature.ID)
				if err != nil {
					t.Fatal(err)
				}
				child := children[0]
				if child.Priority != models.PriorityHigh {
					t.Errorf("Priority = %q, want %q", child.Priority, models.PriorityHigh)
				}
				if child.Description != "Implement UI for theme selection" {
					t.Errorf("Description = %q", child.Description)
				}
			},
		},
		{
			name:    "missing parent ID",
			args:    []string{"--kind", "bug"},
			input:   "Test\n",
			wantErr: true,
		},
		{
			name:    "invalid parent ID",
			args:    []string{"abc", "--kind", "bug"},
			input:   "Test\n",
			wantErr: true,
			errMsg:  "invalid parent ID",
		},
		{
			name:    "non-existent parent",
			args:    []string{"999", "--kind", "bug"},
			input:   "Test\n",
			wantErr: true,
			errMsg:  "parent task not found",
		},
		{
			name:    "missing kind flag",
			args:    []string{fmt.Sprintf("%s", parentBug.ID)},
			input:   "Test\n",
			wantErr: true,
			errMsg:  "required flag",
		},
		{
			name:    "invalid kind",
			args:    []string{fmt.Sprintf("%s", parentBug.ID), "--kind", "invalid"},
			input:   "Test\n",
			wantErr: true,
			errMsg:  "invalid kind",
		},
		{
			name:    "empty input",
			args:    []string{fmt.Sprintf("%s", parentBug.ID), "--kind", "bug"},
			input:   "\n",
			wantErr: true,
			errMsg:  "title cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader(tt.input)

			cmd := newAddSubtaskCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetIn(stdin)
			cmd.SetArgs(tt.args)

			// Clear any existing children
			children, _ := testRepo.GetChildren(parentBug.ID)
			for _, child := range children {
				if err := testRepo.Delete(child.ID); err != nil {
					t.Errorf("Failed to delete child task: %v", err)
				}
			}
			children, _ = testRepo.GetChildren(parentFeature.ID)
			for _, child := range children {
				if err := testRepo.Delete(child.ID); err != nil {
					t.Errorf("Failed to delete child task: %v", err)
				}
			}

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
