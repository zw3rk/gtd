package services

import (
	"path/filepath"
	"testing"

	"github.com/zw3rk/gtd/internal/database"
	"github.com/zw3rk/gtd/internal/models"
)

// TestTaskServiceCreate tests task creation
func TestTaskServiceCreate(t *testing.T) {
	// Setup with real database for integration
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	repo := models.NewTaskRepository(db)
	service := NewTaskService(repo)

	tests := []struct {
		name    string
		task    *models.Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: models.NewTask(models.KindBug, "Test Bug", "Description"),
			wantErr: false,
		},
		{
			name: "invalid task - no title",
			task: &models.Task{
				Kind:        models.KindBug,
				Description: "Description",
			},
			wantErr: true,
		},
		{
			name: "invalid task - no description",
			task: &models.Task{
				Kind:  models.KindBug,
				Title: "Title",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateTask(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && err == nil {
				// Verify task was created
				retrieved, err := service.GetTask(tt.task.ID)
				if err != nil {
					t.Errorf("Failed to retrieve created task: %v", err)
				}
				if retrieved.Title != tt.task.Title {
					t.Errorf("Retrieved task title = %s, want %s", retrieved.Title, tt.task.Title)
				}
			}
		})
	}
}

// TestTaskServiceStateTransitions tests state transition logic
func TestTaskServiceStateTransitions(t *testing.T) {
	// Setup
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	repo := models.NewTaskRepository(db)
	service := NewTaskService(repo)

	// Create a task
	task := models.NewTask(models.KindBug, "Test Task", "Description")
	if err := service.CreateTask(task); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		fromState string
		toState   string
		wantErr   bool
	}{
		{
			name:      "inbox to new (accept)",
			fromState: models.StateInbox,
			toState:   models.StateNew,
			wantErr:   false,
		},
		{
			name:      "inbox to invalid (reject)",
			fromState: models.StateInbox,
			toState:   models.StateInvalid,
			wantErr:   false,
		},
		{
			name:      "inbox to in_progress (invalid)",
			fromState: models.StateInbox,
			toState:   models.StateInProgress,
			wantErr:   true,
		},
		{
			name:      "new to in_progress",
			fromState: models.StateNew,
			toState:   models.StateInProgress,
			wantErr:   false,
		},
		{
			name:      "in_progress to done",
			fromState: models.StateInProgress,
			toState:   models.StateDone,
			wantErr:   false,
		},
		{
			name:      "done to in_progress (reopen)",
			fromState: models.StateDone,
			toState:   models.StateInProgress,
			wantErr:   false,
		},
		{
			name:      "cancelled to new (reopen)",
			fromState: models.StateCancelled,
			toState:   models.StateNew,
			wantErr:   false,
		},
		{
			name:      "invalid to new (not allowed)",
			fromState: models.StateInvalid,
			toState:   models.StateNew,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new task for each test to avoid state conflicts
			testTask := models.NewTask(models.KindBug, "Test Task", "Description")
			if err := service.CreateTask(testTask); err != nil {
				t.Fatal(err)
			}

			// Move task to the required initial state through valid transitions
			switch tt.fromState {
			case models.StateNew:
				if err := service.AcceptTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
			case models.StateInProgress:
				if err := service.AcceptTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
				if err := service.StartTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
			case models.StateDone:
				if err := service.AcceptTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
				if err := service.CompleteTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
			case models.StateCancelled:
				if err := service.AcceptTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
				if err := service.CancelTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
			case models.StateInvalid:
				if err := service.RejectTask(testTask.ID); err != nil {
					t.Fatal(err)
				}
			case models.StateInbox:
				// Already in inbox
			}

			// Try transition
			err := service.UpdateTaskState(testTask.ID, tt.toState)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTaskState() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify state changed
				updated, _ := service.GetTask(testTask.ID)
				if updated.State != tt.toState {
					t.Errorf("Task state = %s, want %s", updated.State, tt.toState)
				}
			}
		})
	}
}

// TestTaskServiceAcceptReject tests accept/reject functionality
func TestTaskServiceAcceptReject(t *testing.T) {
	// Setup
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	repo := models.NewTaskRepository(db)
	service := NewTaskService(repo)

	// Create a task in INBOX
	task := models.NewTask(models.KindBug, "Inbox Task", "Description")
	if err := service.CreateTask(task); err != nil {
		t.Fatal(err)
	}

	t.Run("accept task", func(t *testing.T) {
		err := service.AcceptTask(task.ID)
		if err != nil {
			t.Errorf("AcceptTask() error = %v", err)
		}

		// Verify state
		updated, _ := service.GetTask(task.ID)
		if updated.State != models.StateNew {
			t.Errorf("Accepted task state = %s, want %s", updated.State, models.StateNew)
		}
	})

	t.Run("reject task", func(t *testing.T) {
		// Create another task
		task2 := models.NewTask(models.KindBug, "Reject Task", "Description")
		if err := service.CreateTask(task2); err != nil {
			t.Fatal(err)
		}

		err := service.RejectTask(task2.ID)
		if err != nil {
			t.Errorf("RejectTask() error = %v", err)
		}

		// Verify state
		updated, _ := service.GetTask(task2.ID)
		if updated.State != models.StateInvalid {
			t.Errorf("Rejected task state = %s, want %s", updated.State, models.StateInvalid)
		}
	})

	t.Run("cannot reject done task", func(t *testing.T) {
		// Create and complete a task through proper workflow
		task3 := models.NewTask(models.KindBug, "Done Task", "Description")
		if err := service.CreateTask(task3); err != nil {
			t.Fatal(err)
		}
		// Accept it first
		if err := service.AcceptTask(task3.ID); err != nil {
			t.Fatal(err)
		}
		// Then complete it
		if err := service.CompleteTask(task3.ID); err != nil {
			t.Fatal(err)
		}

		err := service.RejectTask(task3.ID)
		if err == nil {
			t.Error("Expected error rejecting done task")
		}
	})
}

// TestTaskServiceBlocking tests blocking functionality
func TestTaskServiceBlocking(t *testing.T) {
	// Setup
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	repo := models.NewTaskRepository(db)
	service := NewTaskService(repo)

	// Create two tasks
	task1 := models.NewTask(models.KindBug, "Blocker Task", "Must be done first")
	task2 := models.NewTask(models.KindFeature, "Blocked Task", "Depends on blocker")

	if err := service.CreateTask(task1); err != nil {
		t.Fatal(err)
	}
	if err := service.CreateTask(task2); err != nil {
		t.Fatal(err)
	}

	t.Run("block task", func(t *testing.T) {
		err := service.BlockTask(task2.ID, task1.ID)
		if err != nil {
			t.Errorf("BlockTask() error = %v", err)
		}

		// Verify blocking
		updated, _ := service.GetTask(task2.ID)
		if updated.BlockedBy == nil || *updated.BlockedBy != task1.ID {
			t.Error("Task should be blocked")
		}
	})

	t.Run("cannot block by self", func(t *testing.T) {
		err := service.BlockTask(task1.ID, task1.ID)
		if err == nil {
			t.Error("Expected error blocking task by itself")
		}
	})

	t.Run("unblock task", func(t *testing.T) {
		err := service.UnblockTask(task2.ID)
		if err != nil {
			t.Errorf("UnblockTask() error = %v", err)
		}

		// Verify unblocked
		updated, _ := service.GetTask(task2.ID)
		if updated.BlockedBy != nil {
			t.Error("Task should not be blocked")
		}
	})
}

// TestTaskServiceParentChild tests parent-child relationships
func TestTaskServiceParentChild(t *testing.T) {
	// Setup
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	repo := models.NewTaskRepository(db)
	service := NewTaskService(repo)

	// Create parent task
	parent := models.NewTask(models.KindFeature, "Parent Task", "Has subtasks")
	if err := service.CreateTask(parent); err != nil {
		t.Fatal(err)
	}

	// Accept parent
	if err := service.AcceptTask(parent.ID); err != nil {
		t.Fatal(err)
	}

	// Create child tasks
	child1 := models.NewTask(models.KindBug, "Child 1", "First subtask")
	child1.Parent = &parent.ID
	if err := service.CreateTask(child1); err != nil {
		t.Fatal(err)
	}

	child2 := models.NewTask(models.KindBug, "Child 2", "Second subtask")
	child2.Parent = &parent.ID
	if err := service.CreateTask(child2); err != nil {
		t.Fatal(err)
	}

	t.Run("get subtasks", func(t *testing.T) {
		subtasks, err := service.GetSubtasks(parent.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(subtasks) != 2 {
			t.Errorf("Expected 2 subtasks, got %d", len(subtasks))
		}
	})

	t.Run("cannot complete parent with incomplete children", func(t *testing.T) {
		// Accept children
		if err := service.AcceptTask(child1.ID); err != nil {
			t.Fatal(err)
		}
		if err := service.AcceptTask(child2.ID); err != nil {
			t.Fatal(err)
		}

		// Try to complete parent
		err := service.CompleteTask(parent.ID)
		if err == nil {
			t.Error("Expected error completing parent with incomplete children")
		}
	})

	t.Run("can complete parent after children", func(t *testing.T) {
		// Complete children
		if err := service.CompleteTask(child1.ID); err != nil {
			t.Fatal(err)
		}
		if err := service.CompleteTask(child2.ID); err != nil {
			t.Fatal(err)
		}

		// Now complete parent
		err := service.CompleteTask(parent.ID)
		if err != nil {
			t.Errorf("CompleteTask() error = %v", err)
		}

		// Verify
		updated, _ := service.GetTask(parent.ID)
		if updated.State != models.StateDone {
			t.Errorf("Parent state = %s, want %s", updated.State, models.StateDone)
		}
	})
}

// TestTaskServiceSearch tests search functionality
func TestTaskServiceSearch(t *testing.T) {
	// Setup
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	repo := models.NewTaskRepository(db)
	service := NewTaskService(repo)

	// Create test tasks
	tasks := []*models.Task{
		models.NewTask(models.KindBug, "Memory leak in parser", "High memory usage"),
		models.NewTask(models.KindFeature, "Add caching", "Improve memory efficiency"),
		models.NewTask(models.KindBug, "Crash on startup", "Application crashes"),
	}

	for _, task := range tasks {
		if err := service.CreateTask(task); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "search by title",
			query:    "memory leak",
			expected: 1,
		},
		{
			name:     "search by title and description",
			query:    "memory",
			expected: 2, // Matches both title and description
		},
		{
			name:     "search not found",
			query:    "nonexistent",
			expected: 0,
		},
		{
			name:     "search partial match",
			query:    "crash",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.SearchTasks(tt.query)
			if err != nil {
				t.Fatal(err)
			}

			if len(results) != tt.expected {
				t.Errorf("Search results = %d, want %d", len(results), tt.expected)
			}
		})
	}
}

// TestTaskServiceReopen tests the reopen functionality
func TestTaskServiceReopen(t *testing.T) {
	// Setup
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	repo := models.NewTaskRepository(db)
	service := NewTaskService(repo)

	// Create and cancel a task
	task := models.NewTask(models.KindFeature, "Cancelled Feature", "Was cancelled")
	if err := service.CreateTask(task); err != nil {
		t.Fatal(err)
	}

	// Accept then cancel
	if err := service.AcceptTask(task.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.CancelTask(task.ID); err != nil {
		t.Fatal(err)
	}

	t.Run("reopen cancelled task", func(t *testing.T) {
		err := service.ReopenTask(task.ID)
		if err != nil {
			t.Errorf("ReopenTask() error = %v", err)
		}

		// Verify state
		updated, _ := service.GetTask(task.ID)
		if updated.State != models.StateNew {
			t.Errorf("Reopened task state = %s, want %s", updated.State, models.StateNew)
		}
	})

	t.Run("cannot reopen non-cancelled task", func(t *testing.T) {
		// Task is now NEW, try to reopen again
		err := service.ReopenTask(task.ID)
		if err == nil {
			t.Error("Expected error reopening non-cancelled task")
		}
	})
}
