package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/database"
	"github.com/zw3rk/gtd/internal/models"
)

func setupTestCommand(t *testing.T) (*database.Database, *models.TaskRepository, func()) {
	// Create test database
	testDB, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	if err := testDB.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	testRepo := models.NewTaskRepository(testDB)

	// Override global variables
	oldDB, oldRepo := db, repo
	db, repo = testDB, testRepo

	cleanup := func() {
		if err := testDB.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
		db, repo = oldDB, oldRepo
	}

	return testDB, testRepo, cleanup
}

func TestAddBugCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	tests := []struct {
		name    string
		args    []string
		input   string
		flags   map[string]string
		wantErr bool
		check   func(t *testing.T, tasks []*models.Task)
	}{
		{
			name:  "create bug with title and default description",
			args:  []string{},
			input: "Fix memory leak\n\nMemory usage increases over time",
			check: func(t *testing.T, tasks []*models.Task) {
				if len(tasks) != 1 {
					t.Fatalf("Expected 1 task, got %d", len(tasks))
				}
				task := tasks[0]
				if task.Title != "Fix memory leak" {
					t.Errorf("Title = %q, want %q", task.Title, "Fix memory leak")
				}
				if task.Kind != models.KindBug {
					t.Errorf("Kind = %q, want %q", task.Kind, models.KindBug)
				}
				if task.Priority != models.PriorityMedium {
					t.Errorf("Priority = %q, want %q", task.Priority, models.PriorityMedium)
				}
				if task.State != models.StateInbox {
					t.Errorf("State = %q, want %q", task.State, models.StateInbox)
				}
			},
		},
		{
			name:  "create bug with title and description",
			args:  []string{},
			input: "Fix memory leak\nMemory usage grows over time in the worker process",
			check: func(t *testing.T, tasks []*models.Task) {
				if len(tasks) != 1 {
					t.Fatalf("Expected 1 task, got %d", len(tasks))
				}
				task := tasks[0]
				if task.Description != "Memory usage grows over time in the worker process" {
					t.Errorf("Description = %q", task.Description)
				}
			},
		},
		{
			name:  "create bug with high priority",
			args:  []string{"--priority", "high"},
			input: "Critical security bug\n\nThis bug needs immediate attention due to security implications",
			check: func(t *testing.T, tasks []*models.Task) {
				task := tasks[0]
				if task.Priority != models.PriorityHigh {
					t.Errorf("Priority = %q, want %q", task.Priority, models.PriorityHigh)
				}
			},
		},
		{
			name:  "create bug with source",
			args:  []string{"--source", "auth.go:42"},
			input: "Fix auth bypass\n\nUsers can bypass authentication in the login flow",
			check: func(t *testing.T, tasks []*models.Task) {
				task := tasks[0]
				if task.Source != "auth.go:42" {
					t.Errorf("Source = %q, want %q", task.Source, "auth.go:42")
				}
			},
		},
		{
			name:  "create bug with tags",
			args:  []string{"--tags", "security,urgent"},
			input: "Security issue\n\nA security vulnerability was discovered in the system",
			check: func(t *testing.T, tasks []*models.Task) {
				task := tasks[0]
				if task.Tags != "security,urgent" {
					t.Errorf("Tags = %q, want %q", task.Tags, "security,urgent")
				}
			},
		},
		{
			name:    "empty input",
			args:    []string{},
			input:   "\n",
			wantErr: true,
		},
		{
			name:    "invalid priority",
			args:    []string{"--priority", "urgent"},
			input:   "Test\n\nTest description",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader(tt.input)

			cmd := newAddBugCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetIn(stdin)
			cmd.SetArgs(tt.args)

			// Clear any existing tasks
			tasks, _ := testRepo.List(models.ListOptions{All: true})
			for _, task := range tasks {
				if err := testRepo.Delete(task.ID); err != nil {
					t.Errorf("Failed to delete task %s: %v", task.ID, err)
				}
			}

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("stderr: %s", stderr.String())
				return
			}

			if !tt.wantErr && tt.check != nil {
				tasks, err := testRepo.List(models.ListOptions{All: true})
				if err != nil {
					t.Fatal(err)
				}
				tt.check(t, tasks)
			}
		})
	}
}

func TestAddFeatureCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("Add dark mode\nImplement theme switching")

	cmd := newAddFeatureCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(stdin)
	cmd.SetArgs([]string{"--priority", "high", "--tags", "ui,enhancement"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tasks, err := testRepo.List(models.ListOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Kind != models.KindFeature {
		t.Errorf("Kind = %q, want %q", task.Kind, models.KindFeature)
	}
	if task.Title != "Add dark mode" {
		t.Errorf("Title = %q, want %q", task.Title, "Add dark mode")
	}
	if task.Description != "Implement theme switching" {
		t.Errorf("Description = %q", task.Description)
	}
	if task.Priority != models.PriorityHigh {
		t.Errorf("Priority = %q, want %q", task.Priority, models.PriorityHigh)
	}
	if task.Tags != "ui,enhancement" {
		t.Errorf("Tags = %q, want %q", task.Tags, "ui,enhancement")
	}
}

func TestAddRegressionCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader("Login broken after update\n\nLogin functionality stopped working after updating to v2.1.0")

	cmd := newAddRegressionCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(stdin)
	cmd.SetArgs([]string{"--source", "v2.1.0"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	tasks, err := testRepo.List(models.ListOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}

	task := tasks[0]
	if task.Kind != models.KindRegression {
		t.Errorf("Kind = %q, want %q", task.Kind, models.KindRegression)
	}
	if task.Source != "v2.1.0" {
		t.Errorf("Source = %q, want %q", task.Source, "v2.1.0")
	}
}
