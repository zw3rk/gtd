// Package database provides the SQLite database layer for claude-gtd
package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Database wraps the SQL database connection
type Database struct {
	DB *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	// Open database with foreign key support
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to connect to database: %w (also failed to close: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure for better performance and concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to set WAL mode: %w (also failed to close: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	return &Database{DB: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// Begin starts a new transaction
func (d *Database) Begin() (*sql.Tx, error) {
	return d.DB.Begin()
}

// CreateSchema creates the database schema
func (d *Database) CreateSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		parent TEXT REFERENCES tasks(id),
		priority TEXT CHECK(priority IN ('high', 'medium', 'low')) DEFAULT 'medium',
		state TEXT CHECK(state IN ('INBOX', 'NEW', 'IN_PROGRESS', 'DONE', 'CANCELLED', 'INVALID')) DEFAULT 'INBOX',
		kind TEXT CHECK(kind IN ('BUG', 'FEATURE', 'REGRESSION')) NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		author TEXT NOT NULL,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		source TEXT,
		blocked_by TEXT REFERENCES tasks(id),
		tags TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_state_priority ON tasks(state, priority);
	CREATE INDEX IF NOT EXISTS idx_parent ON tasks(parent);
	CREATE INDEX IF NOT EXISTS idx_id_prefix ON tasks(substr(id, 1, 7));
	CREATE INDEX IF NOT EXISTS idx_kind_state ON tasks(kind, state);
	CREATE INDEX IF NOT EXISTS idx_blocked_by ON tasks(blocked_by) WHERE blocked_by IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_created ON tasks(created);
	CREATE INDEX IF NOT EXISTS idx_updated ON tasks(updated);
	CREATE INDEX IF NOT EXISTS idx_tags ON tasks(tags) WHERE tags IS NOT NULL;

	-- Trigger to update the updated timestamp
	CREATE TRIGGER IF NOT EXISTS update_task_timestamp 
	AFTER UPDATE ON tasks
	BEGIN
		UPDATE tasks SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;
	`

	_, err := d.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Run migrations
	if err := d.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations runs database migrations to update schema
func (d *Database) runMigrations() error {
	// Check if we need to add INBOX and INVALID states
	// First, check if the constraint exists with the old states
	var constraintSQL string
	err := d.DB.QueryRow(`
		SELECT sql FROM sqlite_master 
		WHERE type='table' AND name='tasks' AND sql LIKE '%CHECK(state IN%'
	`).Scan(&constraintSQL)

	if err == nil && constraintSQL != "" {
		// Check if INBOX is already in the constraint
		if !strings.Contains(constraintSQL, "'INBOX'") {
			// We need to migrate - this requires recreating the table
			tx, err := d.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin migration transaction: %w", err)
			}
			defer func() {
				if err != nil {
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						fmt.Fprintf(os.Stderr, "Failed to rollback migration: %v\n", rollbackErr)
					}
				}
			}()

			// Create new table with updated schema
			_, err = tx.Exec(`
				CREATE TABLE tasks_new (
					id TEXT PRIMARY KEY,
					parent TEXT REFERENCES tasks_new(id),
					priority TEXT CHECK(priority IN ('high', 'medium', 'low')) DEFAULT 'medium',
					state TEXT CHECK(state IN ('INBOX', 'NEW', 'IN_PROGRESS', 'DONE', 'CANCELLED', 'INVALID')) DEFAULT 'INBOX',
					kind TEXT CHECK(kind IN ('BUG', 'FEATURE', 'REGRESSION')) NOT NULL,
					title TEXT NOT NULL,
					description TEXT,
					author TEXT NOT NULL,
					created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					source TEXT,
					blocked_by TEXT REFERENCES tasks_new(id),
					tags TEXT
				)
			`)
			if err != nil {
				return fmt.Errorf("failed to create new table: %w", err)
			}

			// Copy data from old table
			_, err = tx.Exec(`
				INSERT INTO tasks_new 
				SELECT * FROM tasks
			`)
			if err != nil {
				return fmt.Errorf("failed to copy data: %w", err)
			}

			// Drop old table
			_, err = tx.Exec(`DROP TABLE tasks`)
			if err != nil {
				return fmt.Errorf("failed to drop old table: %w", err)
			}

			// Rename new table
			_, err = tx.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
			if err != nil {
				return fmt.Errorf("failed to rename table: %w", err)
			}

			// Recreate indexes
			_, err = tx.Exec(`
				CREATE INDEX idx_state_priority ON tasks(state, priority);
				CREATE INDEX idx_parent ON tasks(parent);
				CREATE INDEX idx_id_prefix ON tasks(substr(id, 1, 7));
				CREATE INDEX idx_kind_state ON tasks(kind, state);
				CREATE INDEX idx_blocked_by ON tasks(blocked_by) WHERE blocked_by IS NOT NULL;
				CREATE INDEX idx_created ON tasks(created);
				CREATE INDEX idx_updated ON tasks(updated);
				CREATE INDEX idx_tags ON tasks(tags) WHERE tags IS NOT NULL;
			`)
			if err != nil {
				return fmt.Errorf("failed to recreate indexes: %w", err)
			}

			// Recreate trigger
			_, err = tx.Exec(`
				CREATE TRIGGER update_task_timestamp 
				AFTER UPDATE ON tasks
				BEGIN
					UPDATE tasks SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END;
			`)
			if err != nil {
				return fmt.Errorf("failed to recreate trigger: %w", err)
			}

			if err = tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration: %w", err)
			}
		}
	}

	// Add new performance indices if they don't exist
	newIndices := []string{
		"CREATE INDEX IF NOT EXISTS idx_kind_state ON tasks(kind, state)",
		"CREATE INDEX IF NOT EXISTS idx_blocked_by ON tasks(blocked_by) WHERE blocked_by IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_created ON tasks(created)",
		"CREATE INDEX IF NOT EXISTS idx_updated ON tasks(updated)",
		"CREATE INDEX IF NOT EXISTS idx_tags ON tasks(tags) WHERE tags IS NOT NULL",
	}

	for _, indexSQL := range newIndices {
		if _, err := d.DB.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
