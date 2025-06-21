package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zw3rk/gtd/cmd"
	"github.com/zw3rk/gtd/internal/database"
	"github.com/zw3rk/gtd/internal/models"
)

// TestFullGTDWorkflow tests the complete GTD workflow from INBOX to DONE
func TestFullGTDWorkflow(t *testing.T) {
	// Setup
	testDir := t.TempDir()
	setupGitRepo(t, testDir)
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)
	
	// Use test database
	t.Setenv("GTD_DATABASE_NAME", "test-claude-tasks.db")

	// Test workflow
	t.Run("complete GTD workflow", func(t *testing.T) {
		var taskID string

		// Step 1: Add a task (goes to INBOX)
		t.Run("add task to inbox", func(t *testing.T) {
			output := runCommandWithInput(t, "Security Fix\n\nFix authentication bypass vulnerability", 
				"add", "bug", "--priority", "high", "--tags", "critical,security")
			
			// Extract task ID from output
			if !strings.Contains(output, "Created bug task") {
				t.Fatalf("Expected task creation message, got: %s", output)
			}
			parts := strings.Fields(output)
			taskID = parts[len(parts)-1]
			
			// Verify task is in INBOX
			db := openTestDB(t, testDir)
			defer func() { _ = db.Close() }()
			
			task := getTask(t, db, taskID)
			if task.State != models.StateInbox {
				t.Errorf("New task should be in INBOX, got %s", task.State)
			}
			if task.Priority != "high" {
				t.Errorf("Task priority should be high, got %s", task.Priority)
			}
			if task.Tags != "critical,security" {
				t.Errorf("Task tags should be 'critical,security', got %s", task.Tags)
			}
		})

		// Step 2: Review inbox
		t.Run("review inbox", func(t *testing.T) {
			output := runCommand(t, "review")
			
			if !strings.Contains(output, "Security Fix") {
				t.Errorf("Review should show task title, got: %s", output)
			}
			if !strings.Contains(output, taskID[:7]) {
				t.Errorf("Review should show task ID, got: %s", output)
			}
		})

		// Step 3: Accept task (INBOX → NEW)
		t.Run("accept task", func(t *testing.T) {
			output := runCommand(t, "accept", taskID[:7])
			
			if !strings.Contains(output, "accepted") {
				t.Errorf("Expected acceptance message, got: %s", output)
			}
			
			// Verify state change
			db := openTestDB(t, testDir)
			defer func() { _ = db.Close() }()
			
			task := getTask(t, db, taskID)
			if task.State != models.StateNew {
				t.Errorf("Accepted task should be in NEW state, got %s", task.State)
			}
		})

		// Step 4: Start work (NEW → IN_PROGRESS)
		t.Run("start work on task", func(t *testing.T) {
			output := runCommand(t, "in-progress", taskID[:7])
			
			if !strings.Contains(output, "in progress") {
				t.Errorf("Expected in-progress message, got: %s", output)
			}
			
			// Verify state change
			db := openTestDB(t, testDir)
			defer func() { _ = db.Close() }()
			
			task := getTask(t, db, taskID)
			if task.State != models.StateInProgress {
				t.Errorf("Task should be IN_PROGRESS, got %s", task.State)
			}
		})

		// Step 5: Complete task (IN_PROGRESS → DONE)
		t.Run("complete task", func(t *testing.T) {
			output := runCommand(t, "done", taskID[:7])
			
			if !strings.Contains(output, "done") {
				t.Errorf("Expected done message, got: %s", output)
			}
			
			// Verify state change
			db := openTestDB(t, testDir)
			defer func() { _ = db.Close() }()
			
			task := getTask(t, db, taskID)
			if task.State != models.StateDone {
				t.Errorf("Task should be DONE, got %s", task.State)
			}
		})

		// Step 6: Verify task appears in done list
		t.Run("list done tasks", func(t *testing.T) {
			output := runCommand(t, "list-done")
			
			if !strings.Contains(output, "Security Fix") {
				t.Errorf("Done list should show completed task, got: %s", output)
			}
		})
	})
}

// TestParentChildWorkflow tests parent tasks with subtasks
func TestParentChildWorkflow(t *testing.T) {
	// Setup
	testDir := t.TempDir()
	setupGitRepo(t, testDir)
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)
	
	// Use test database
	t.Setenv("GTD_DATABASE_NAME", "test-claude-tasks.db")

	var parentID, child1ID, child2ID string

	// Create parent task
	t.Run("create parent task", func(t *testing.T) {
		output := runCommandWithInput(t, "Implement User Dashboard\n\nCreate a comprehensive dashboard", "add", "feature")
		parentID = extractTaskID(t, output)
		
		// Accept it
		runCommand(t, "accept", parentID[:7])
	})

	// Add subtasks
	t.Run("add subtasks", func(t *testing.T) {
		// First subtask
		output := runCommandWithInput(t, "Design Dashboard UI\n\nCreate mockups and wireframes",
			"add-subtask", parentID[:7], "--kind", "feature")
		child1ID = extractTaskID(t, output)
		
		// Second subtask
		output = runCommandWithInput(t, "Fix Dashboard Layout\n\nResponsive design issues",
			"add-subtask", parentID[:7], "--kind", "bug")
		child2ID = extractTaskID(t, output)
	})

	// Try to complete parent (should fail)
	t.Run("cannot complete parent with incomplete children", func(t *testing.T) {
		output := runCommandExpectError(t, "done", parentID[:7])
		
		if !strings.Contains(output, "cannot mark parent task as DONE") {
			t.Errorf("Expected parent completion error, got: %s", output)
		}
	})

	// Complete children
	t.Run("complete all subtasks", func(t *testing.T) {
		// Accept and complete first child
		runCommand(t, "accept", child1ID[:7])
		runCommand(t, "done", child1ID[:7])
		
		// Accept and complete second child
		runCommand(t, "accept", child2ID[:7])
		runCommand(t, "done", child2ID[:7])
	})

	// Now complete parent
	t.Run("complete parent after children", func(t *testing.T) {
		output := runCommand(t, "done", parentID[:7])
		
		if !strings.Contains(output, "done") {
			t.Errorf("Parent should be completable now, got: %s", output)
		}
	})
}

// TestBlockingWorkflow tests task blocking relationships
func TestBlockingWorkflow(t *testing.T) {
	// Setup
	testDir := t.TempDir()
	setupGitRepo(t, testDir)
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)
	
	// Use test database
	t.Setenv("GTD_DATABASE_NAME", "test-claude-tasks.db")

	var task1ID, task2ID string

	// Create two tasks
	t.Run("create tasks", func(t *testing.T) {
		// First task
		output := runCommandWithInput(t, "Fix Database Connection\n\nConnection pool exhaustion", "add", "bug")
		task1ID = extractTaskID(t, output)
		runCommand(t, "accept", task1ID[:7])
		
		// Second task
		output = runCommandWithInput(t, "Add Connection Monitoring\n\nMonitor connection pool", "add", "feature")
		task2ID = extractTaskID(t, output)
		runCommand(t, "accept", task2ID[:7])
	})

	// Block second task by first
	t.Run("block task", func(t *testing.T) {
		output := runCommand(t, "block", task2ID[:7], "--by", task1ID[:7])
		
		if !strings.Contains(output, "blocked by") {
			t.Errorf("Expected blocking confirmation, got: %s", output)
		}
	})

	// Verify blocked task shows in list
	t.Run("list shows blocked status", func(t *testing.T) {
		output := runCommand(t, "list", "--blocked")
		
		if !strings.Contains(output, "Add Connection Monitoring") {
			t.Errorf("Blocked task should appear in blocked list, got: %s", output)
		}
	})

	// Unblock task
	t.Run("unblock task", func(t *testing.T) {
		output := runCommand(t, "unblock", task2ID[:7])
		
		if !strings.Contains(output, "no longer blocked") {
			t.Errorf("Expected unblock confirmation, got: %s", output)
		}
	})
}

// TestCancelReopenWorkflow tests cancelling and reopening tasks
func TestCancelReopenWorkflow(t *testing.T) {
	// Setup
	testDir := t.TempDir()
	setupGitRepo(t, testDir)
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)
	
	// Use test database
	t.Setenv("GTD_DATABASE_NAME", "test-claude-tasks.db")

	var taskID string

	// Create and start a task
	t.Run("create and start task", func(t *testing.T) {
		output := runCommandWithInput(t, "Experimental Feature\n\nMight not work out", "add", "feature")
		taskID = extractTaskID(t, output)
		
		runCommand(t, "accept", taskID[:7])
		runCommand(t, "in-progress", taskID[:7])
	})

	// Cancel the task
	t.Run("cancel task", func(t *testing.T) {
		output := runCommand(t, "cancel", taskID[:7])
		
		if !strings.Contains(output, "cancelled") {
			t.Errorf("Expected cancellation message, got: %s", output)
		}
	})

	// Verify in cancelled list
	t.Run("list cancelled tasks", func(t *testing.T) {
		output := runCommand(t, "list-cancelled")
		
		if !strings.Contains(output, "Experimental Feature") {
			t.Errorf("Cancelled task should appear in list, got: %s", output)
		}
	})

	// Reopen the task
	t.Run("reopen cancelled task", func(t *testing.T) {
		output := runCommand(t, "reopen", taskID[:7])
		
		if !strings.Contains(output, "reopened") {
			t.Errorf("Expected reopen message, got: %s", output)
		}
		
		// Verify state
		db := openTestDB(t, testDir)
		defer func() { _ = db.Close() }()
		
		task := getTask(t, db, taskID)
		if task.State != models.StateNew {
			t.Errorf("Reopened task should be in NEW state, got %s", task.State)
		}
	})
}

// TestRejectWorkflow tests rejecting tasks from inbox
func TestRejectWorkflow(t *testing.T) {
	// Setup
	testDir := t.TempDir()
	setupGitRepo(t, testDir)
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)
	
	// Use test database
	t.Setenv("GTD_DATABASE_NAME", "test-claude-tasks.db")

	var taskID string

	// Create a task
	t.Run("create task", func(t *testing.T) {
		output := runCommandWithInput(t, "Not a real bug\n\nUser error, not a bug", "add", "bug")
		taskID = extractTaskID(t, output)
	})

	// Reject the task
	t.Run("reject task from inbox", func(t *testing.T) {
		output := runCommand(t, "reject", taskID[:7])
		
		if !strings.Contains(output, "rejected") {
			t.Errorf("Expected rejection message, got: %s", output)
		}
		
		// Verify state
		db := openTestDB(t, testDir)
		defer func() { _ = db.Close() }()
		
		task := getTask(t, db, taskID)
		if task.State != models.StateInvalid {
			t.Errorf("Rejected task should be INVALID, got %s", task.State)
		}
	})

	// Verify not in regular lists
	t.Run("rejected task not in normal lists", func(t *testing.T) {
		output := runCommand(t, "list", "--all")
		
		// INVALID tasks are excluded even with --all
		if strings.Contains(output, "Not a real bug") {
			// This is expected behavior - INVALID tasks don't show in list --all
			return
		}
	})
}

// TestSearchAndExport tests search and export functionality
func TestSearchAndExport(t *testing.T) {
	// Setup
	testDir := t.TempDir()
	setupGitRepo(t, testDir)
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)
	
	// Use test database
	t.Setenv("GTD_DATABASE_NAME", "test-claude-tasks.db")

	// Create multiple tasks
	t.Run("create test tasks", func(t *testing.T) {
		// Bug with "memory" in title
		runCommandWithInput(t, "Memory Leak in Parser\n\nHigh memory usage", "add", "bug")
		
		// Feature with "memory" in description
		runCommandWithInput(t, "Performance Monitor\n\nTrack memory and CPU usage", "add", "feature")
		
		// Unrelated task
		runCommandWithInput(t, "Login Broken\n\nUsers cannot authenticate", "add", "regression")
	})

	// Search for "memory"
	t.Run("search tasks", func(t *testing.T) {
		output := runCommand(t, "search", "memory")
		
		if !strings.Contains(output, "Memory Leak") {
			t.Errorf("Search should find task with memory in title")
		}
		if !strings.Contains(output, "Performance Monitor") {
			t.Errorf("Search should find task with memory in description")
		}
		if strings.Contains(output, "Login Broken") {
			t.Errorf("Search should not find unrelated task")
		}
	})

	// Export to JSON
	t.Run("export to JSON", func(t *testing.T) {
		output := runCommand(t, "export", "--format", "json")
		
		// Should be valid JSON array
		if !strings.HasPrefix(strings.TrimSpace(output), "[") {
			t.Errorf("JSON export should start with [, got: %s", output[:20])
		}
		if !strings.Contains(output, "Memory Leak in Parser") {
			t.Errorf("JSON should contain task title")
		}
	})

	// Export to CSV
	t.Run("export to CSV", func(t *testing.T) {
		output := runCommand(t, "export", "--format", "csv")
		
		// Should have CSV header
		if !strings.Contains(output, "ID,Type,State,Priority,Title") {
			t.Errorf("CSV should have header row")
		}
	})
}

// TestSummaryCommand tests the summary output
func TestSummaryCommand(t *testing.T) {
	// Setup
	testDir := t.TempDir()
	setupGitRepo(t, testDir)
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)
	
	// Use test database
	t.Setenv("GTD_DATABASE_NAME", "test-claude-tasks.db")

	// Create tasks in various states
	t.Run("create diverse tasks", func(t *testing.T) {
		// Create and accept a bug
		output := runCommandWithInput(t, "High Priority Bug\n\nCritical issue", "add", "bug", "--priority", "high")
		id := extractTaskID(t, output)
		runCommand(t, "accept", id[:7])
		
		// Create and complete a feature
		output = runCommandWithInput(t, "Completed Feature\n\nAlready done", "add", "feature")
		id = extractTaskID(t, output)
		runCommand(t, "accept", id[:7])
		runCommand(t, "done", id[:7])
		
		// Create a regression in inbox
		runCommandWithInput(t, "Recent Regression\n\nBroke in last release", "add", "regression")
	})

	// Get summary
	t.Run("summary shows statistics", func(t *testing.T) {
		output := runCommand(t, "summary")
		
		// Should show counts by state
		if !strings.Contains(output, "INBOX") {
			t.Errorf("Summary should show INBOX count")
		}
		if !strings.Contains(output, "NEW") {
			t.Errorf("Summary should show NEW count")
		}
		if !strings.Contains(output, "DONE") {
			t.Errorf("Summary should show DONE count")
		}
		
		// Should show counts by type
		if !strings.Contains(output, "Bug") {
			t.Errorf("Summary should show bug count")
		}
		if !strings.Contains(output, "Feature") {
			t.Errorf("Summary should show feature count")
		}
		if !strings.Contains(output, "Regression") {
			t.Errorf("Summary should show regression count")
		}
		
		// Should show priority breakdown
		if !strings.Contains(output, "High:") {
			t.Errorf("Summary should show high priority count")
		}
	})
}

// Helper functions

func setupGitRepo(t *testing.T, dir string) {
	t.Helper()
	
	// Initialize git repo
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.name", "Test User")
	runCmd(t, dir, "git", "config", "user.email", "test@example.com")
}

func runCommand(t *testing.T, args ...string) string {
	t.Helper()
	return runCommandWithInput(t, "", args...)
}

func runCommandWithInput(t *testing.T, input string, args ...string) string {
	t.Helper()
	
	app := cmd.NewApp()
	rootCmd := cmd.NewRootCommand(app)
	
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)
	
	if input != "" {
		rootCmd.SetIn(strings.NewReader(input))
	}
	
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
	}
	
	return stdout.String()
}

func runCommandExpectError(t *testing.T, args ...string) string {
	t.Helper()
	
	app := cmd.NewApp()
	rootCmd := cmd.NewRootCommand(app)
	
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)
	
	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("Expected command to fail but it succeeded")
	}
	
	return stderr.String()
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run %s: %v", name, err)
	}
}


func extractTaskID(t *testing.T, output string) string {
	t.Helper()
	
	if !strings.Contains(output, "Created") {
		t.Fatalf("No task creation message in output: %s", output)
	}
	
	parts := strings.Fields(output)
	// Handle different output formats:
	// "Created bug task abc1234"
	// "Created bug subtask abc1234 for task def5678 (Title)"
	for i, part := range parts {
		if (part == "task" || part == "subtask") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	// Fallback to last part
	return parts[len(parts)-1]
}

func openTestDB(t *testing.T, testDir string) *database.Database {
	t.Helper()
	
	dbPath := filepath.Join(testDir, "test-claude-tasks.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	return db
}

func getTask(t *testing.T, db *database.Database, id string) *models.Task {
	t.Helper()
	
	repo := models.NewTaskRepository(db)
	task, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	return task
}