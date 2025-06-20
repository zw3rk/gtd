package database

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewEdgeCases tests edge cases for database creation
func TestNewEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		wantErr bool
	}{
		{
			name: "read-only directory",
			setup: func() (string, func()) {
				dir := t.TempDir()
				dbPath := filepath.Join(dir, "readonly.db")
				// Make directory read-only
				if err := os.Chmod(dir, 0444); err != nil {
					t.Skip("Cannot change directory permissions")
				}
				cleanup := func() {
					_ = os.Chmod(dir, 0755)
				}
				return dbPath, cleanup
			},
			wantErr: true,
		},
		{
			name: "very long path",
			setup: func() (string, func()) {
				dir := t.TempDir()
				// Create a very long path
				longName := ""
				for i := 0; i < 50; i++ {
					longName += "verylongdirectoryname/"
				}
				dbPath := filepath.Join(dir, longName, "test.db")
				return dbPath, func() {}
			},
			wantErr: true,
		},
		{
			name: "special characters in filename",
			setup: func() (string, func()) {
				dir := t.TempDir()
				dbPath := filepath.Join(dir, "test@#$%.db")
				return dbPath, func() {}
			},
			wantErr: false,
		},
		{
			name: "unicode filename",
			setup: func() (string, func()) {
				dir := t.TempDir()
				dbPath := filepath.Join(dir, "测试数据库.db")
				return dbPath, func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath, cleanup := tt.setup()
			defer cleanup()

			db, err := New(dbPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && db != nil {
				defer func() {
					if err := db.Close(); err != nil {
						t.Errorf("failed to close database: %v", err)
					}
				}()
			}
		})
	}
}

// TestDatabasePragmas tests that all pragmas are set correctly
func TestDatabasePragmas(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "pragma_test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	pragmas := []struct {
		name     string
		query    string
		expected interface{}
	}{
		{
			name:     "journal_mode",
			query:    "PRAGMA journal_mode",
			expected: "wal",
		},
		{
			name:     "foreign_keys",
			query:    "PRAGMA foreign_keys",
			expected: 1,
		},
		{
			name:     "busy_timeout",
			query:    "PRAGMA busy_timeout",
			expected: 5000,
		},
		{
			name:     "synchronous",
			query:    "PRAGMA synchronous",
			expected: 1, // NORMAL in SQLite
		},
	}

	for _, p := range pragmas {
		t.Run(p.name, func(t *testing.T) {
			var result interface{}
			switch p.expected.(type) {
			case string:
				var s string
				if err := db.DB.QueryRow(p.query).Scan(&s); err != nil {
					t.Fatalf("Failed to scan string result: %v", err)
				}
				result = s
			case int:
				var i int
				if err := db.DB.QueryRow(p.query).Scan(&i); err != nil {
					t.Fatalf("Failed to scan int result: %v", err)
				}
				result = i
			}

			if result != p.expected {
				t.Errorf("Pragma %s = %v, want %v", p.name, result, p.expected)
			}
		})
	}
}

// TestCloseMultipleTimes tests that Close() can be called multiple times safely
func TestCloseMultipleTimes(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "multiclose_test.db"))
	if err != nil {
		t.Fatal(err)
	}

	// Close multiple times
	for i := 0; i < 3; i++ {
		err := db.Close()
		if i == 0 && err != nil {
			t.Errorf("First Close() failed: %v", err)
		}
		// Subsequent closes might return an error, which is fine
	}
}

// TestDatabaseMemoryMode tests in-memory database
func TestDatabaseMemoryMode(t *testing.T) {
	// Test with :memory: database
	db, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}()

	// Should be able to create schema
	if err := db.CreateSchema(); err != nil {
		t.Fatalf("CreateSchema() failed for memory database: %v", err)
	}

	// Should be able to insert data
	_, err = db.DB.Exec(`
		INSERT INTO tasks (id, kind, title, author) 
		VALUES ('test1', 'BUG', 'Test Task', 'Test User')
	`)
	if err != nil {
		t.Errorf("Failed to insert into memory database: %v", err)
	}

	// Verify data exists
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("Expected 1 task in memory database, got %d", count)
	}
}

// TestTransactionIsolation tests transaction isolation
func TestTransactionIsolation(t *testing.T) {
	db, err := New(filepath.Join(t.TempDir(), "isolation_test.db"))
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

	// Start a transaction
	tx1, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx1.Rollback()
	}()

	// Insert in transaction
	_, err = tx1.Exec(`
		INSERT INTO tasks (id, kind, title, author) 
		VALUES ('tx_test', 'BUG', 'Transaction Test', 'Test User')
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Query outside transaction - should not see the insert
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = 'tx_test'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Uncommitted transaction data visible outside transaction")
	}

	// Commit transaction
	if err := tx1.Commit(); err != nil {
		t.Fatal(err)
	}

	// Now it should be visible
	err = db.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = 'tx_test'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("Committed transaction data not visible")
	}
}