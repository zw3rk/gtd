// Package database provides the SQLite database layer for claude-gtd
package database

import (
	"database/sql"
	"fmt"

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
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure for better performance and concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
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
		state TEXT CHECK(state IN ('NEW', 'IN_PROGRESS', 'DONE', 'CANCELLED')) DEFAULT 'NEW',
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

	return nil
}
