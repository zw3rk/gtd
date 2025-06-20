package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunMigrations tests the migration logic
func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(db *sql.DB) error
		wantErr   bool
		verify    func(db *sql.DB) error
	}{
		{
			name: "no migration needed - fresh database",
			setupFunc: func(db *sql.DB) error {
				// Fresh database, no setup needed
				return nil
			},
			wantErr: false,
			verify: func(db *sql.DB) error {
				// Verify table has new constraint
				var constraintSQL string
				err := db.QueryRow(`
					SELECT sql FROM sqlite_master 
					WHERE type='table' AND name='tasks' AND sql LIKE '%CHECK(state IN%'
				`).Scan(&constraintSQL)
				if err != nil {
					return fmt.Errorf("failed to find constraint: %w", err)
				}
				// Should have INBOX and INVALID states
				if !contains(constraintSQL, "'INBOX'") {
					return fmt.Errorf("constraint missing INBOX state")
				}
				if !contains(constraintSQL, "'INVALID'") {
					return fmt.Errorf("constraint missing INVALID state")
				}
				return nil
			},
		},
		{
			name: "migration from old schema",
			setupFunc: func(db *sql.DB) error {
				// Create old-style table without INBOX/INVALID states
				_, err := db.Exec(`
					CREATE TABLE tasks (
						id TEXT PRIMARY KEY,
						parent TEXT REFERENCES tasks(id),
						priority TEXT CHECK(priority IN ('high', 'medium', 'low')) DEFAULT 'medium',
						state TEXT CHECK(state IN ('NEW', 'IN_PROGRESS', 'DONE', 'CANCELLED')) DEFAULT 'NEW',
						kind TEXT CHECK(kind IN ('BUG', 'FEATURE', 'REGRESSION')) NOT NULL,
						title TEXT NOT NULL,
						description TEXT,
						author TEXT NOT NULL DEFAULT 'Test User <test@example.com>',
						created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						source TEXT,
						blocked_by TEXT REFERENCES tasks(id),
						tags TEXT
					)
				`)
				if err != nil {
					return err
				}

				// Insert some test data
				_, err = db.Exec(`
					INSERT INTO tasks (id, kind, title, description, state) 
					VALUES 
					('task1', 'BUG', 'Test Bug', 'Description', 'NEW'),
					('task2', 'FEATURE', 'Test Feature', 'Description', 'IN_PROGRESS')
				`)
				return err
			},
			wantErr: false,
			verify: func(db *sql.DB) error {
				// Verify data was preserved
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
				if err != nil {
					return err
				}
				if count != 2 {
					return fmt.Errorf("expected 2 tasks, got %d", count)
				}

				// Verify new constraint exists
				var constraintSQL string
				err = db.QueryRow(`
					SELECT sql FROM sqlite_master 
					WHERE type='table' AND name='tasks' AND sql LIKE '%CHECK(state IN%'
				`).Scan(&constraintSQL)
				if err != nil {
					return fmt.Errorf("failed to find constraint: %w", err)
				}
				if !contains(constraintSQL, "'INBOX'") {
					return fmt.Errorf("constraint missing INBOX state")
				}
				return nil
			},
		},
		{
			name: "migration with parent-child relationships",
			setupFunc: func(db *sql.DB) error {
				// Create old-style table
				_, err := db.Exec(`
					CREATE TABLE tasks (
						id TEXT PRIMARY KEY,
						parent TEXT REFERENCES tasks(id),
						priority TEXT CHECK(priority IN ('high', 'medium', 'low')) DEFAULT 'medium',
						state TEXT CHECK(state IN ('NEW', 'IN_PROGRESS', 'DONE', 'CANCELLED')) DEFAULT 'NEW',
						kind TEXT CHECK(kind IN ('BUG', 'FEATURE', 'REGRESSION')) NOT NULL,
						title TEXT NOT NULL,
						description TEXT,
						author TEXT NOT NULL DEFAULT 'Test User <test@example.com>',
						created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						source TEXT,
						blocked_by TEXT REFERENCES tasks(id),
						tags TEXT
					)
				`)
				if err != nil {
					return err
				}

				// Insert parent and child tasks
				_, err = db.Exec(`
					INSERT INTO tasks (id, kind, title, description, state) 
					VALUES ('parent1', 'FEATURE', 'Parent Task', 'Parent Description', 'NEW')
				`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`
					INSERT INTO tasks (id, parent, kind, title, description, state) 
					VALUES ('child1', 'parent1', 'BUG', 'Child Task', 'Child Description', 'NEW')
				`)
				return err
			},
			wantErr: false,
			verify: func(db *sql.DB) error {
				// Verify parent-child relationship preserved
				var parent string
				err := db.QueryRow("SELECT parent FROM tasks WHERE id = 'child1'").Scan(&parent)
				if err != nil {
					return err
				}
				if parent != "parent1" {
					return fmt.Errorf("parent relationship not preserved, got %s", parent)
				}
				return nil
			},
		},
		{
			name: "migration with blocking relationships",
			setupFunc: func(db *sql.DB) error {
				// Create old-style table
				_, err := db.Exec(`
					CREATE TABLE tasks (
						id TEXT PRIMARY KEY,
						parent TEXT REFERENCES tasks(id),
						priority TEXT CHECK(priority IN ('high', 'medium', 'low')) DEFAULT 'medium',
						state TEXT CHECK(state IN ('NEW', 'IN_PROGRESS', 'DONE', 'CANCELLED')) DEFAULT 'NEW',
						kind TEXT CHECK(kind IN ('BUG', 'FEATURE', 'REGRESSION')) NOT NULL,
						title TEXT NOT NULL,
						description TEXT,
						author TEXT NOT NULL DEFAULT 'Test User <test@example.com>',
						created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						source TEXT,
						blocked_by TEXT REFERENCES tasks(id),
						tags TEXT
					)
				`)
				if err != nil {
					return err
				}

				// Insert tasks with blocking relationship
				_, err = db.Exec(`
					INSERT INTO tasks (id, kind, title, description, state) 
					VALUES 
					('blocker1', 'BUG', 'Blocking Task', 'Must be done first', 'NEW'),
					('blocked1', 'FEATURE', 'Blocked Task', 'Waiting on blocker', 'NEW')
				`)
				if err != nil {
					return err
				}

				// Set blocking relationship
				_, err = db.Exec("UPDATE tasks SET blocked_by = 'blocker1' WHERE id = 'blocked1'")
				return err
			},
			wantErr: false,
			verify: func(db *sql.DB) error {
				// Verify blocking relationship preserved
				var blockedBy sql.NullString
				err := db.QueryRow("SELECT blocked_by FROM tasks WHERE id = 'blocked1'").Scan(&blockedBy)
				if err != nil {
					return err
				}
				if !blockedBy.Valid || blockedBy.String != "blocker1" {
					return fmt.Errorf("blocking relationship not preserved")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh database for each test
			dbPath := filepath.Join(t.TempDir(), "migration_test.db")
			db, err := New(dbPath)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("failed to close database: %v", err)
				}
			}()

			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(db.DB); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Run CreateSchema which includes migrations
			err = db.CreateSchema()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify results if provided
			if tt.verify != nil && !tt.wantErr {
				if err := tt.verify(db.DB); err != nil {
					t.Errorf("verification failed: %v", err)
				}
			}
		})
	}
}

// TestCreateSchemaIdempotent verifies CreateSchema can be called multiple times
func TestCreateSchemaIdempotent(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "idempotent_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	// Call CreateSchema multiple times
	for i := 0; i < 3; i++ {
		if err := db.CreateSchema(); err != nil {
			t.Errorf("CreateSchema() call %d failed: %v", i+1, err)
		}
	}

	// Verify only one tasks table exists
	var count int
	err = db.DB.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='tasks'
	`).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("Expected 1 tasks table, got %d", count)
	}
}

// TestDatabaseConstraints tests database constraints
func TestDatabaseConstraints(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "constraints_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name: "valid task insert",
			query: `INSERT INTO tasks (id, kind, title, author) 
					VALUES ('test1', 'BUG', 'Test Task', 'Test User <test@example.com>')`,
			wantErr: false,
		},
		{
			name: "invalid state",
			query: `INSERT INTO tasks (id, kind, title, author, state) 
					VALUES ('test2', 'BUG', 'Test Task', 'Test User', 'INVALID_STATE')`,
			wantErr: true,
		},
		{
			name: "invalid priority",
			query: `INSERT INTO tasks (id, kind, title, author, priority) 
					VALUES ('test3', 'BUG', 'Test Task', 'Test User', 'extreme')`,
			wantErr: true,
		},
		{
			name: "invalid kind",
			query: `INSERT INTO tasks (id, kind, title, author) 
					VALUES ('test4', 'INVALID_KIND', 'Test Task', 'Test User')`,
			wantErr: true,
		},
		{
			name: "null title",
			query: `INSERT INTO tasks (id, kind, author) 
					VALUES ('test5', 'BUG', 'Test User')`,
			wantErr: true,
		},
		{
			name: "null kind",
			query: `INSERT INTO tasks (id, title, author) 
					VALUES ('test6', 'Test Task', 'Test User')`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.DB.Exec(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("query error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDatabaseConcurrency tests concurrent access
func TestDatabaseConcurrency(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "concurrent_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	if err := db.CreateSchema(); err != nil {
		t.Fatal(err)
	}

	// Run multiple goroutines inserting data
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			taskID := fmt.Sprintf("task%d", id)
			_, err := db.DB.Exec(`
				INSERT INTO tasks (id, kind, title, author) 
				VALUES (?, 'BUG', ?, 'Test User')
			`, taskID, fmt.Sprintf("Task %d", id))
			
			if err != nil {
				t.Errorf("Concurrent insert %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all inserts succeeded
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 10 {
		t.Errorf("Expected 10 tasks, got %d", count)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && strings.Contains(s, substr)
}