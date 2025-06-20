package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/models"
)

func TestListCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	// Create test tasks with various states and priorities
	tasks := []struct {
		kind      string
		title     string
		state     string
		priority  string
		source    string
		tags      string
		blockedBy *string
	}{
		{models.KindBug, "High priority bug in progress", models.StateInProgress, models.PriorityHigh, "app.go:42", "backend,urgent", nil},
		{models.KindFeature, "Medium priority feature", models.StateNew, models.PriorityMedium, "", "ui", nil},
		{models.KindBug, "Low priority bug", models.StateNew, models.PriorityLow, "", "", nil},
		{models.KindRegression, "Done regression", models.StateDone, models.PriorityHigh, "", "", nil},
		{models.KindBug, "Cancelled bug", models.StateCancelled, models.PriorityMedium, "", "", nil},
	}

	createdTasks := make([]*models.Task, 0, len(tasks))
	for _, tt := range tasks {
		task := models.NewTask(tt.kind, tt.title, "Description for "+tt.title)
		task.State = tt.state
		task.Priority = tt.priority
		task.Source = tt.source
		task.Tags = tt.tags
		task.BlockedBy = tt.blockedBy
		if err := testRepo.Create(task); err != nil {
			t.Fatal(err)
		}
		createdTasks = append(createdTasks, task)
	}

	// Create blocked task
	blockedTask := models.NewTask(models.KindFeature, "Blocked feature", "This feature is blocked by another task")
	blockedTask.State = models.StateNew
	blockedTask.Priority = models.PriorityHigh
	blockedTask.BlockedBy = &createdTasks[0].ID
	if err := testRepo.Create(blockedTask); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		contains    []string
		notContains []string
	}{
		{
			name: "default list (no done/cancelled)",
			args: []string{},
			contains: []string{
				"High priority bug in progress",
				"Medium priority feature",
				"Low priority bug",
				"Blocked feature",
			},
			notContains: []string{
				"Done regression",
				"Cancelled bug",
			},
		},
		{
			name: "oneline format",
			args: []string{"--oneline"},
			contains: []string{
				"High priority bug in progress",
				"▶", // IN_PROGRESS symbol
			},
			notContains: []string{
				"Source:",
				"Tags:",
			},
		},
		{
			name: "show all tasks",
			args: []string{"--all"},
			contains: []string{
				"High priority bug in progress",
				"Done regression",
				"Cancelled bug",
			},
		},
		{
			name: "filter by state",
			args: []string{"--state", "NEW"},
			contains: []string{
				"Medium priority feature",
				"Low priority bug",
			},
			notContains: []string{
				"High priority bug in progress",
			},
		},
		{
			name: "filter by priority",
			args: []string{"--priority", "high"},
			contains: []string{
				"High priority bug in progress",
				"Blocked feature",
			},
			notContains: []string{
				"Medium priority feature",
				"Low priority bug",
			},
		},
		{
			name: "filter by kind",
			args: []string{"--kind", "bug"},
			contains: []string{
				"High priority bug in progress",
				"Low priority bug",
			},
			notContains: []string{
				"Medium priority feature",
				"Blocked feature",
			},
		},
		{
			name: "filter by tag",
			args: []string{"--tag", "backend"},
			contains: []string{
				"High priority bug in progress",
			},
			notContains: []string{
				"Medium priority feature",
			},
		},
		{
			name: "show blocked tasks",
			args: []string{"--blocked"},
			contains: []string{
				"Blocked feature",
				"Blocked-by:",
			},
			notContains: []string{
				"High priority bug in progress",
				"Medium priority feature",
			},
		},
		{
			name:    "invalid state filter",
			args:    []string{"--state", "INVALID"},
			wantErr: true,
		},
		{
			name:    "invalid priority filter",
			args:    []string{"--priority", "urgent"},
			wantErr: true,
		},
		{
			name:    "invalid kind filter",
			args:    []string{"--kind", "task"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cmd := newListCommand()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
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

func TestListDoneCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	// Create tasks
	doneTask1 := models.NewTask(models.KindBug, "Fixed bug", "This bug has been fixed")
	doneTask1.State = models.StateDone
	if err := testRepo.Create(doneTask1); err != nil {
		t.Fatal(err)
	}

	doneTask2 := models.NewTask(models.KindFeature, "Completed feature", "This feature has been completed")
	doneTask2.State = models.StateDone
	doneTask2.Priority = models.PriorityHigh
	if err := testRepo.Create(doneTask2); err != nil {
		t.Fatal(err)
	}

	activeTask := models.NewTask(models.KindBug, "Active bug", "This bug is currently active")
	if err := testRepo.Create(activeTask); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	cmd := newListDoneCommand()
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Should contain done tasks
	if !strings.Contains(output, "Fixed bug") {
		t.Error("Output should contain 'Fixed bug'")
	}
	if !strings.Contains(output, "Completed feature") {
		t.Error("Output should contain 'Completed feature'")
	}

	// Should not contain active task
	if strings.Contains(output, "Active bug") {
		t.Error("Output should not contain 'Active bug'")
	}
}

func TestListCancelledCommand(t *testing.T) {
	_, testRepo, cleanup := setupTestCommand(t)
	defer cleanup()

	// Create tasks
	cancelledTask := models.NewTask(models.KindBug, "Cancelled bug", "This bug was cancelled")
	cancelledTask.State = models.StateCancelled
	if err := testRepo.Create(cancelledTask); err != nil {
		t.Fatal(err)
	}

	activeTask := models.NewTask(models.KindBug, "Active bug", "This bug is currently active")
	if err := testRepo.Create(activeTask); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	cmd := newListCancelledCommand()
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Should contain cancelled task
	if !strings.Contains(output, "Cancelled bug") {
		t.Error("Output should contain 'Cancelled bug'")
	}

	// Should not contain active task
	if strings.Contains(output, "Active bug") {
		t.Error("Output should not contain 'Active bug'")
	}
}
