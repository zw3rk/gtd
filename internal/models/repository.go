package models

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/zw3rk/gtd/internal/database"
)

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *database.Database
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *database.Database) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create inserts a new task into the database
func (r *TaskRepository) Create(task *Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		INSERT INTO tasks (id, parent, priority, state, kind, title, description, author, source, blocked_by, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.DB.Exec(query,
		task.ID,
		task.Parent,
		task.Priority,
		task.State,
		task.Kind,
		task.Title,
		task.Description,
		task.Author,
		task.Source,
		task.BlockedBy,
		task.Tags,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// Update modifies an existing task
func (r *TaskRepository) Update(task *Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	query := `
		UPDATE tasks
		SET parent = ?, priority = ?, state = ?, kind = ?, title = ?, 
		    description = ?, author = ?, source = ?, blocked_by = ?, tags = ?
		WHERE id = ?
	`

	_, err := r.db.DB.Exec(query,
		task.Parent,
		task.Priority,
		task.State,
		task.Kind,
		task.Title,
		task.Description,
		task.Author,
		task.Source,
		task.BlockedBy,
		task.Tags,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// Delete removes a task from the database
func (r *TaskRepository) Delete(id string) error {
	_, err := r.db.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// GetByID retrieves a task by its ID or hash prefix
func (r *TaskRepository) GetByID(id string) (*Task, error) {
	// First try exact match
	task, err := r.getByExactID(id)
	if err == nil {
		return task, nil
	}

	// If not found and input looks like a hash prefix (4+ chars), try prefix match
	if len(id) >= 4 && len(id) < 40 {
		return r.getByHashPrefix(id)
	}

	return nil, fmt.Errorf("task not found")
}

// getByExactID retrieves a task by its exact ID
func (r *TaskRepository) getByExactID(id string) (*Task, error) {
	task := &Task{}
	query := `
		SELECT id, parent, priority, state, kind, title, description, author,
		       created, updated, source, blocked_by, tags
		FROM tasks
		WHERE id = ?
	`

	err := r.db.DB.QueryRow(query, id).Scan(
		&task.ID,
		&task.Parent,
		&task.Priority,
		&task.State,
		&task.Kind,
		&task.Title,
		&task.Description,
		&task.Author,
		&task.Created,
		&task.Updated,
		&task.Source,
		&task.BlockedBy,
		&task.Tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// getByHashPrefix retrieves a task by hash prefix (like git)
func (r *TaskRepository) getByHashPrefix(prefix string) (*Task, error) {
	query := `
		SELECT id, parent, priority, state, kind, title, description, author,
		       created, updated, source, blocked_by, tags
		FROM tasks
		WHERE id LIKE ? || '%'
	`

	rows, err := r.db.DB.Query(query, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to search by prefix: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close rows: %v\n", err)
		}
	}()

	tasks, err := r.scanTasks(rows)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("task not found")
	}
	if len(tasks) > 1 {
		return nil, fmt.Errorf("ambiguous hash prefix '%s' matches %d tasks", prefix, len(tasks))
	}

	return tasks[0], nil
}

// GetChildren retrieves all child tasks of a parent
func (r *TaskRepository) GetChildren(parentID string) ([]*Task, error) {
	query := `
		SELECT id, parent, priority, state, kind, title, description, author,
		       created, updated, source, blocked_by, tags
		FROM tasks
		WHERE parent = ?
		ORDER BY priority DESC, created ASC
	`

	rows, err := r.db.DB.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close rows: %v\n", err)
		}
	}()

	return r.scanTasks(rows)
}

// ListOptions contains filtering options for listing tasks
type ListOptions struct {
	State         string
	Priority      string
	Kind          string
	Tag           string
	Blocked       bool
	ShowDone      bool
	ShowCancelled bool
	Limit         int
	All           bool
}

// List retrieves tasks based on the given options
func (r *TaskRepository) List(opts ListOptions) ([]*Task, error) {
	var conditions []string
	var args []interface{}

	// Default: exclude INBOX, DONE, CANCELLED, and INVALID unless specifically requested
	if !opts.All && opts.State == "" {
		excludeStates := []string{}
		if !opts.ShowDone {
			excludeStates = append(excludeStates, "'DONE'")
		}
		if !opts.ShowCancelled {
			excludeStates = append(excludeStates, "'CANCELLED'")
		}
		// Always exclude INBOX and INVALID unless explicitly requested
		excludeStates = append(excludeStates, "'INBOX'", "'INVALID'")

		if len(excludeStates) > 0 {
			conditions = append(conditions, fmt.Sprintf("state NOT IN (%s)", strings.Join(excludeStates, ", ")))
		}
	}

	// Add specific filters
	if opts.State != "" {
		conditions = append(conditions, "state = ?")
		args = append(args, opts.State)
	}
	if opts.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, opts.Priority)
	}
	if opts.Kind != "" {
		conditions = append(conditions, "kind = ?")
		args = append(args, opts.Kind)
	}
	if opts.Tag != "" {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, "%"+opts.Tag+"%")
	}
	if opts.Blocked {
		conditions = append(conditions, "blocked_by IS NOT NULL")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build the query with proper ordering
	query := fmt.Sprintf(`
		SELECT id, parent, priority, state, kind, title, description, author,
		       created, updated, source, blocked_by, tags
		FROM tasks
		%s
		ORDER BY 
			CASE state 
				WHEN 'IN_PROGRESS' THEN 0
				WHEN 'NEW' THEN 1
				ELSE 2
			END,
			CASE priority
				WHEN 'high' THEN 0
				WHEN 'medium' THEN 1
				WHEN 'low' THEN 2
			END,
			created DESC
	`, whereClause)

	// Add limit if not showing all
	if !opts.All && opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := r.db.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close rows: %v\n", err)
		}
	}()

	return r.scanTasks(rows)
}

// ListByState retrieves all tasks with a specific state
func (r *TaskRepository) ListByState(state string) ([]*Task, error) {
	query := `
		SELECT id, parent, priority, state, kind, title, description, author,
		       created, updated, source, blocked_by, tags
		FROM tasks
		WHERE state = ?
		ORDER BY created DESC
	`

	rows, err := r.db.DB.Query(query, state)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks by state: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close rows: %v\n", err)
		}
	}()

	return r.scanTasks(rows)
}

// Search finds tasks by searching in title and description
func (r *TaskRepository) Search(query string) ([]*Task, error) {
	searchQuery := `
		SELECT id, parent, priority, state, kind, title, description, author,
		       created, updated, source, blocked_by, tags
		FROM tasks
		WHERE LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)
		ORDER BY created DESC
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.DB.Query(searchQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close rows: %v\n", err)
		}
	}()

	return r.scanTasks(rows)
}

// UpdateState changes the state of a task
func (r *TaskRepository) UpdateState(id string, newState string) error {
	// Get the task first
	task, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Get children if any
	children, err := r.GetChildren(id)
	if err != nil {
		return err
	}

	// Check if transition is allowed
	if !task.CanTransitionTo(newState, children) {
		// Provide more detailed error for parent/child state conflicts
		if newState == StateDone && len(children) > 0 {
			for _, child := range children {
				if child.State != StateDone && child.State != StateCancelled {
					return fmt.Errorf("cannot mark parent task as DONE: child task %s is in %s state", child.ID, child.State)
				}
			}
		}
		return fmt.Errorf("cannot transition from %s to %s", task.State, newState)
	}

	// Update the state
	_, err = r.db.DB.Exec("UPDATE tasks SET state = ? WHERE id = ?", newState, id)
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}

// Block sets a task as blocked by another task
func (r *TaskRepository) Block(taskID, blockingTaskID string) error {
	// Verify both tasks exist
	if _, err := r.GetByID(taskID); err != nil {
		return fmt.Errorf("task to block not found: %w", err)
	}
	if _, err := r.GetByID(blockingTaskID); err != nil {
		return fmt.Errorf("blocking task not found: %w", err)
	}

	_, err := r.db.DB.Exec("UPDATE tasks SET blocked_by = ? WHERE id = ?", blockingTaskID, taskID)
	if err != nil {
		return fmt.Errorf("failed to block task: %w", err)
	}

	return nil
}

// Unblock removes the blocking relationship from a task
func (r *TaskRepository) Unblock(taskID string) error {
	_, err := r.db.DB.Exec("UPDATE tasks SET blocked_by = NULL WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to unblock task: %w", err)
	}

	return nil
}

// scanTasks is a helper to scan multiple task rows
func (r *TaskRepository) scanTasks(rows *sql.Rows) ([]*Task, error) {
	var tasks []*Task

	for rows.Next() {
		task := &Task{}
		err := rows.Scan(
			&task.ID,
			&task.Parent,
			&task.Priority,
			&task.State,
			&task.Kind,
			&task.Title,
			&task.Description,
			&task.Author,
			&task.Created,
			&task.Updated,
			&task.Source,
			&task.BlockedBy,
			&task.Tags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return tasks, nil
}
