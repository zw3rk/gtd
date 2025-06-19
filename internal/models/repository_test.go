package models

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/internal/database"
)

func setupTestDB(t *testing.T) *TaskRepository {
	db, err := database.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	return NewTaskRepository(db)
}

func TestTaskRepository_Create(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	task := NewTask(KindBug, "Test bug", "Test description")
	task.Priority = PriorityHigh
	task.Source = "test.go:42"
	task.Tags = "test,bug"

	err := repo.Create(task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if task.ID == 0 {
		t.Error("Create() did not set task ID")
	}

	// Verify the task was saved
	saved, err := repo.GetByID(task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if saved.Title != task.Title {
		t.Errorf("Saved task title = %v, want %v", saved.Title, task.Title)
	}
	if saved.Priority != task.Priority {
		t.Errorf("Saved task priority = %v, want %v", saved.Priority, task.Priority)
	}
}

func TestTaskRepository_CreateWithParent(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	// Create parent task
	parent := NewTask(KindFeature, "Parent feature", "")
	if err := repo.Create(parent); err != nil {
		t.Fatal(err)
	}

	// Create child task
	child := NewTask(KindBug, "Child bug", "")
	child.Parent = &parent.ID
	if err := repo.Create(child); err != nil {
		t.Fatal(err)
	}

	// Verify parent relationship
	saved, err := repo.GetByID(child.ID)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Parent == nil || *saved.Parent != parent.ID {
		t.Error("Child task parent relationship not saved correctly")
	}
}

func TestTaskRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	task := NewTask(KindBug, "Original title", "")
	if err := repo.Create(task); err != nil {
		t.Fatal(err)
	}

	// Update the task
	task.Title = "Updated title"
	task.State = StateInProgress
	task.Priority = PriorityHigh

	if err := repo.Update(task); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify updates
	saved, err := repo.GetByID(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Title != "Updated title" {
		t.Errorf("Updated title = %v, want %v", saved.Title, "Updated title")
	}
	if saved.State != StateInProgress {
		t.Errorf("Updated state = %v, want %v", saved.State, StateInProgress)
	}
	if saved.Priority != PriorityHigh {
		t.Errorf("Updated priority = %v, want %v", saved.Priority, PriorityHigh)
	}
}

func TestTaskRepository_Delete(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	task := NewTask(KindBug, "To be deleted", "")
	if err := repo.Create(task); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err := repo.GetByID(task.ID)
	if err == nil {
		t.Error("Expected error getting deleted task")
	}
}

func TestTaskRepository_GetChildren(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	// Create parent
	parent := NewTask(KindFeature, "Parent", "")
	if err := repo.Create(parent); err != nil {
		t.Fatal(err)
	}

	// Create children
	for i := 0; i < 3; i++ {
		child := NewTask(KindBug, "Child task", "")
		child.Parent = &parent.ID
		if err := repo.Create(child); err != nil {
			t.Fatal(err)
		}
	}

	// Get children
	children, err := repo.GetChildren(parent.ID)
	if err != nil {
		t.Fatalf("GetChildren() error = %v", err)
	}

	if len(children) != 3 {
		t.Errorf("GetChildren() returned %d tasks, want 3", len(children))
	}

	for _, child := range children {
		if child.Parent == nil || *child.Parent != parent.ID {
			t.Error("Child task has incorrect parent reference")
		}
	}
}

func TestTaskRepository_List(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	// Create tasks with different states and priorities
	tasks := []struct {
		state    string
		priority string
		title    string
	}{
		{StateInProgress, PriorityHigh, "High priority in progress"},
		{StateInProgress, PriorityLow, "Low priority in progress"},
		{StateNew, PriorityHigh, "High priority new"},
		{StateNew, PriorityMedium, "Medium priority new"},
		{StateDone, PriorityHigh, "Done task"},
	}

	for _, tt := range tasks {
		task := NewTask(KindBug, tt.title, "")
		task.State = tt.state
		task.Priority = tt.priority
		if err := repo.Create(task); err != nil {
			t.Fatal(err)
		}
	}

	// Test default listing (IN_PROGRESS first, then NEW)
	opts := ListOptions{Limit: 20}
	result, err := repo.List(opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(result) != 4 { // Should not include DONE task
		t.Errorf("List() returned %d tasks, want 4", len(result))
	}

	// Verify ordering: IN_PROGRESS tasks should come first
	if result[0].State != StateInProgress {
		t.Error("First task should be IN_PROGRESS")
	}
}

func TestTaskRepository_ListWithFilters(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	// Create test tasks
	bug := NewTask(KindBug, "Bug task", "")
	bug.Tags = "backend,urgent"
	if err := repo.Create(bug); err != nil {
		t.Fatal(err)
	}

	feature := NewTask(KindFeature, "Feature task", "")
	feature.Priority = PriorityHigh
	feature.Tags = "frontend"
	if err := repo.Create(feature); err != nil {
		t.Fatal(err)
	}

	// Test kind filter
	opts := ListOptions{Kind: KindBug}
	result, err := repo.List(opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Kind != KindBug {
		t.Error("Kind filter not working correctly")
	}

	// Test priority filter
	opts = ListOptions{Priority: PriorityHigh}
	result, err = repo.List(opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Priority != PriorityHigh {
		t.Error("Priority filter not working correctly")
	}

	// Test tag filter
	opts = ListOptions{Tag: "backend"}
	result, err = repo.List(opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || !strings.Contains(result[0].Tags, "backend") {
		t.Error("Tag filter not working correctly")
	}
}

func TestTaskRepository_Search(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	// Create tasks with searchable content
	task1 := NewTask(KindBug, "Database connection error", "Connection pool exhausted")
	if err := repo.Create(task1); err != nil {
		t.Fatal(err)
	}

	task2 := NewTask(KindFeature, "Add connection pooling", "Implement database connection pooling")
	if err := repo.Create(task2); err != nil {
		t.Fatal(err)
	}

	task3 := NewTask(KindBug, "Unrelated bug", "Something else entirely")
	if err := repo.Create(task3); err != nil {
		t.Fatal(err)
	}

	// Search for "connection"
	results, err := repo.Search("connection")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Search() returned %d results, want 2", len(results))
	}

	// Verify both matching tasks are returned
	foundTitles := make(map[string]bool)
	for _, task := range results {
		foundTitles[task.Title] = true
	}

	if !foundTitles["Database connection error"] || !foundTitles["Add connection pooling"] {
		t.Error("Search did not return expected tasks")
	}
}

func TestTaskRepository_UpdateState(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	// Create parent and child tasks
	parent := NewTask(KindFeature, "Parent feature", "")
	if err := repo.Create(parent); err != nil {
		t.Fatal(err)
	}

	child := NewTask(KindBug, "Child bug", "")
	child.Parent = &parent.ID
	if err := repo.Create(child); err != nil {
		t.Fatal(err)
	}

	// Try to mark parent as DONE (should fail)
	err := repo.UpdateState(parent.ID, StateDone)
	if err == nil {
		t.Error("Expected error marking parent DONE with incomplete children")
	}

	// Mark child as DONE
	if err := repo.UpdateState(child.ID, StateDone); err != nil {
		t.Fatalf("UpdateState() error = %v", err)
	}

	// First move parent to IN_PROGRESS
	if err := repo.UpdateState(parent.ID, StateInProgress); err != nil {
		t.Fatalf("UpdateState() error = %v", err)
	}

	// Now parent can be marked as DONE
	if err := repo.UpdateState(parent.ID, StateDone); err != nil {
		t.Fatalf("UpdateState() error = %v", err)
	}

	// Verify states
	updatedParent, _ := repo.GetByID(parent.ID)
	if updatedParent.State != StateDone {
		t.Error("Parent state not updated to DONE")
	}
}

func TestTaskRepository_Block(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.db.Close()

	// Create two tasks
	blocker := NewTask(KindBug, "Blocking task", "")
	if err := repo.Create(blocker); err != nil {
		t.Fatal(err)
	}

	blocked := NewTask(KindFeature, "Blocked task", "")
	if err := repo.Create(blocked); err != nil {
		t.Fatal(err)
	}

	// Block the second task
	if err := repo.Block(blocked.ID, blocker.ID); err != nil {
		t.Fatalf("Block() error = %v", err)
	}

	// Verify blocking
	updated, err := repo.GetByID(blocked.ID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.BlockedBy == nil || *updated.BlockedBy != blocker.ID {
		t.Error("Task not properly blocked")
	}

	// Test unblock
	if err := repo.Unblock(blocked.ID); err != nil {
		t.Fatalf("Unblock() error = %v", err)
	}

	updated, err = repo.GetByID(blocked.ID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.BlockedBy != nil {
		t.Error("Task not properly unblocked")
	}
}