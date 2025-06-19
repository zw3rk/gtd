package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zw3rk/claude-gtd/internal/database"
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
		INSERT INTO tasks (parent, priority, state, kind, title, description, source, blocked_by, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.DB.Exec(query,
		task.Parent,
		task.Priority,
		task.State,
		task.Kind,
		task.Title,
		task.Description,
		task.Source,
		task.BlockedBy,
		task.Tags,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get task ID: %w", err)
	}

	task.ID = int(id)
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
		    description = ?, source = ?, blocked_by = ?, tags = ?
		WHERE id = ?
	`

	_, err := r.db.DB.Exec(query,
		task.Parent,
		task.Priority,
		task.State,
		task.Kind,
		task.Title,
		task.Description,
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
func (r *TaskRepository) Delete(id int) error {
	_, err := r.db.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// GetByID retrieves a task by its ID
func (r *TaskRepository) GetByID(id int) (*Task, error) {
	task := &Task{}
	query := `
		SELECT id, parent, priority, state, kind, title, description, 
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

// GetChildren retrieves all child tasks of a parent
func (r *TaskRepository) GetChildren(parentID int) ([]*Task, error) {
	query := `
		SELECT id, parent, priority, state, kind, title, description, 
		       created, updated, source, blocked_by, tags
		FROM tasks
		WHERE parent = ?
		ORDER BY priority DESC, created ASC
	`

	rows, err := r.db.DB.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// ListOptions contains filtering options for listing tasks
type ListOptions struct {
	State      string
	Priority   string
	Kind       string
	Tag        string
	Blocked    bool
	ShowDone   bool
	ShowCancelled bool
	Limit      int
	All        bool
}

// List retrieves tasks based on the given options
func (r *TaskRepository) List(opts ListOptions) ([]*Task, error) {
	var conditions []string
	var args []interface{}

	// Default: exclude DONE and CANCELLED unless specifically requested
	if !opts.ShowDone && !opts.ShowCancelled && opts.State == "" {
		conditions = append(conditions, "state NOT IN ('DONE', 'CANCELLED')")
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
		SELECT id, parent, priority, state, kind, title, description, 
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
	defer rows.Close()

	return r.scanTasks(rows)
}

// Search finds tasks by searching in title and description
func (r *TaskRepository) Search(query string) ([]*Task, error) {
	searchQuery := `
		SELECT id, parent, priority, state, kind, title, description, 
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
	defer rows.Close()

	return r.scanTasks(rows)
}

// UpdateState changes the state of a task
func (r *TaskRepository) UpdateState(id int, newState string) error {
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
					return fmt.Errorf("cannot mark parent task as DONE: child task %d is in %s state", child.ID, child.State)
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
func (r *TaskRepository) Block(taskID, blockingTaskID int) error {
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
func (r *TaskRepository) Unblock(taskID int) error {
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