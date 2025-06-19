package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "creates new database",
			dbPath:  filepath.Join(t.TempDir(), "test.db"),
			wantErr: false,
		},
		{
			name:    "opens existing database",
			dbPath:  filepath.Join(t.TempDir(), "existing.db"),
			wantErr: false,
		},
		{
			name:    "fails with invalid path",
			dbPath:  "/invalid/path/that/does/not/exist/test.db",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the "opens existing" test, create the file first
			if tt.name == "opens existing database" {
				file, err := os.Create(tt.dbPath)
				if err != nil {
					t.Fatal(err)
				}
				if err := file.Close(); err != nil {
					t.Errorf("failed to close test file: %v", err)
				}
			}

			db, err := New(tt.dbPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if db == nil {
					t.Error("New() returned nil database")
					return
				}
				defer func() {
					if err := db.Close(); err != nil {
						t.Errorf("failed to close database: %v", err)
					}
				}()

				// Verify we can query the database
				var result int
				err := db.DB.QueryRow("SELECT 1").Scan(&result)
				if err != nil {
					t.Errorf("Failed to query database: %v", err)
				}
				if result != 1 {
					t.Errorf("Database query returned %d, want 1", result)
				}

				// Verify database file exists
				if _, err := os.Stat(tt.dbPath); os.IsNotExist(err) {
					t.Error("Database file was not created")
				}
			}
		})
	}
}

func TestDatabase_CreateSchema(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "schema_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	// Create schema
	err = db.CreateSchema()
	if err != nil {
		t.Fatalf("CreateSchema() error = %v", err)
	}

	// Verify tasks table exists
	var tableName string
	err = db.DB.QueryRow(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name='tasks'
	`).Scan(&tableName)
	if err != nil {
		t.Fatalf("Failed to find tasks table: %v", err)
	}
	if tableName != "tasks" {
		t.Errorf("Expected table name 'tasks', got '%s'", tableName)
	}

	// Verify all columns exist
	expectedColumns := []string{
		"id", "parent", "priority", "state", "kind",
		"title", "description", "created", "updated",
		"source", "blocked_by", "tags",
	}

	for _, col := range expectedColumns {
		var cid int
		err := db.DB.QueryRow(`
			SELECT cid FROM pragma_table_info('tasks')
			WHERE name = ?
		`, col).Scan(&cid)
		if err != nil {
			t.Errorf("Column %s not found: %v", col, err)
		}
	}

	// Verify indexes exist
	expectedIndexes := []string{
		"idx_state_priority",
		"idx_parent",
	}

	for _, idx := range expectedIndexes {
		var indexName string
		err := db.DB.QueryRow(`
			SELECT name FROM sqlite_master
			WHERE type='index' AND name=?
		`, idx).Scan(&indexName)
		if err != nil {
			t.Errorf("Index %s not found: %v", idx, err)
		}
	}

	// Test that CreateSchema is idempotent
	err = db.CreateSchema()
	if err != nil {
		t.Fatalf("CreateSchema() should be idempotent, got error: %v", err)
	}
}

func TestDatabase_Transaction(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "tx_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	// Create a simple test table
	_, err = db.DB.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	// Test successful transaction
	t.Run("successful transaction", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}

		_, err = tx.Exec("INSERT INTO test (value) VALUES (?)", "test1")
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				t.Errorf("Failed to rollback transaction: %v", rollbackErr)
			}
			t.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatal(err)
		}

		// Verify the insert
		var count int
		err = db.DB.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("Expected 1 row, got %d", count)
		}
	})

	// Test rolled back transaction
	t.Run("rolled back transaction", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}

		_, err = tx.Exec("INSERT INTO test (value) VALUES (?)", "test2")
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				t.Errorf("Failed to rollback transaction: %v", rollbackErr)
			}
			t.Fatal(err)
		}

		// Rollback instead of commit
		err = tx.Rollback()
		if err != nil {
			t.Fatal(err)
		}

		// Verify the insert was rolled back
		var count int
		err = db.DB.QueryRow("SELECT COUNT(*) FROM test WHERE value = 'test2'").Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("Expected 0 rows after rollback, got %d", count)
		}
	})
}

func TestDatabase_Close(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "close_test.db"))
	if err != nil {
		t.Fatal(err)
	}

	// Close the database
	err = db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify we can't use it after closing
	var result int
	err = db.DB.QueryRow("SELECT 1").Scan(&result)
	if err == nil {
		t.Error("Expected error when querying closed database")
	}
}

// Test that we're using WAL mode for better concurrency
func TestDatabase_WALMode(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "wal_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	var journalMode string
	err = db.DB.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatal(err)
	}

	if journalMode != "wal" {
		t.Errorf("Expected journal_mode=wal, got %s", journalMode)
	}
}

// Test foreign key enforcement
func TestDatabase_ForeignKeys(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "fk_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	// Create schema first
	err = db.CreateSchema()
	if err != nil {
		t.Fatal(err)
	}

	// Verify foreign keys are enabled
	var fkEnabled int
	err = db.DB.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatal(err)
	}
	if fkEnabled != 1 {
		t.Error("Foreign keys should be enabled")
	}

	// Test that we can't insert a task with invalid parent
	_, err = db.DB.Exec(`
		INSERT INTO tasks (parent, kind, title) 
		VALUES (999, 'BUG', 'Test task')
	`)
	if err == nil {
		t.Error("Expected foreign key constraint error")
	}
}
