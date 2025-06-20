package services

import (
	"fmt"

	"github.com/zw3rk/gtd/internal/models"
)

// TaskService defines the interface for task operations
type TaskService interface {
	// Task CRUD operations
	CreateTask(task *models.Task) error
	GetTask(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error

	// Task state operations
	UpdateTaskState(id, newState string) error
	AcceptTask(id string) error
	RejectTask(id string) error
	StartTask(id string) error
	CompleteTask(id string) error
	CancelTask(id string) error
	ReopenTask(id string) error

	// Task relationships
	BlockTask(taskID, blockingTaskID string) error
	UnblockTask(taskID string) error
	GetSubtasks(parentID string) ([]*models.Task, error)

	// Task queries
	ListTasks(opts models.ListOptions) ([]*models.Task, error)
	ListByState(state string) ([]*models.Task, error)
	SearchTasks(query string) ([]*models.Task, error)
}

// taskService is the default implementation of TaskService
type taskService struct {
	repo *models.TaskRepository
}

// NewTaskService creates a new task service
func NewTaskService(repo *models.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

// CreateTask creates a new task
func (s *taskService) CreateTask(task *models.Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Create(task)
}

// GetTask retrieves a task by ID
func (s *taskService) GetTask(id string) (*models.Task, error) {
	return s.repo.GetByID(id)
}

// UpdateTask updates an existing task
func (s *taskService) UpdateTask(task *models.Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Update(task)
}

// DeleteTask deletes a task
func (s *taskService) DeleteTask(id string) error {
	return s.repo.Delete(id)
}

// UpdateTaskState updates the state of a task with validation
func (s *taskService) UpdateTaskState(id, newState string) error {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get children for validation
	children, err := s.repo.GetChildren(id)
	if err != nil {
		return fmt.Errorf("failed to get children: %w", err)
	}

	// Validate transition
	if !task.CanTransitionTo(newState, children) {
		return s.getTransitionError(task.State, newState, children)
	}

	return s.repo.UpdateState(id, newState)
}

// AcceptTask moves a task from INBOX to NEW
func (s *taskService) AcceptTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	if task.State != models.StateInbox {
		return fmt.Errorf("task %s is not in INBOX state (current: %s)", task.ShortHash(), task.State)
	}

	return s.UpdateTaskState(id, models.StateNew)
}

// RejectTask marks a task as INVALID
func (s *taskService) RejectTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	if task.State == models.StateDone {
		return fmt.Errorf("cannot mark completed task as invalid")
	}

	return s.UpdateTaskState(id, models.StateInvalid)
}

// StartTask moves a task to IN_PROGRESS
func (s *taskService) StartTask(id string) error {
	return s.UpdateTaskState(id, models.StateInProgress)
}

// CompleteTask marks a task as DONE
func (s *taskService) CompleteTask(id string) error {
	return s.UpdateTaskState(id, models.StateDone)
}

// CancelTask marks a task as CANCELLED
func (s *taskService) CancelTask(id string) error {
	return s.UpdateTaskState(id, models.StateCancelled)
}

// ReopenTask moves a cancelled task back to NEW
func (s *taskService) ReopenTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	if task.State != models.StateCancelled {
		return fmt.Errorf("task %s is not in CANCELLED state (current: %s)", task.ShortHash(), task.State)
	}

	return s.UpdateTaskState(id, models.StateNew)
}

// BlockTask marks a task as blocked by another task
func (s *taskService) BlockTask(taskID, blockingTaskID string) error {
	// Validate both tasks exist
	task, err := s.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("task to block not found: %w", err)
	}

	blockingTask, err := s.GetTask(blockingTaskID)
	if err != nil {
		return fmt.Errorf("blocking task not found: %w", err)
	}

	// Validate not blocking by itself
	if task.ID == blockingTask.ID {
		return fmt.Errorf("cannot block a task by itself")
	}

	return s.repo.Block(taskID, blockingTaskID)
}

// UnblockTask removes the blocking relationship from a task
func (s *taskService) UnblockTask(taskID string) error {
	return s.repo.Unblock(taskID)
}

// GetSubtasks retrieves all subtasks of a parent task
func (s *taskService) GetSubtasks(parentID string) ([]*models.Task, error) {
	return s.repo.GetChildren(parentID)
}

// ListTasks retrieves tasks based on the given options
func (s *taskService) ListTasks(opts models.ListOptions) ([]*models.Task, error) {
	return s.repo.List(opts)
}

// ListByState retrieves all tasks with a specific state
func (s *taskService) ListByState(state string) ([]*models.Task, error) {
	return s.repo.ListByState(state)
}

// SearchTasks searches for tasks by query
func (s *taskService) SearchTasks(query string) ([]*models.Task, error) {
	return s.repo.Search(query)
}

// getTransitionError returns a helpful error message for invalid state transitions
func (s *taskService) getTransitionError(currentState, newState string, children []*models.Task) error {
	// Check for parent task completion with incomplete children
	if newState == models.StateDone && len(children) > 0 {
		for _, child := range children {
			if child.State != models.StateDone && child.State != models.StateCancelled {
				return fmt.Errorf("cannot mark parent task as DONE: child task %s is in %s state", child.ShortHash(), child.State)
			}
		}
	}

	// Provide helpful guidance on valid transitions
	var helpMsg string
	switch currentState {
	case models.StateInbox:
		helpMsg = "use 'gtd accept' to accept the task or 'gtd reject' to mark as invalid"
	case models.StateNew:
		helpMsg = "use 'gtd in-progress' to start work, 'gtd done' to complete, or 'gtd cancel' to cancel"
	case models.StateInProgress:
		helpMsg = "use 'gtd done' to complete or 'gtd cancel' to cancel"
	case models.StateDone:
		helpMsg = "use 'gtd in-progress' to reopen the task"
	case models.StateCancelled:
		helpMsg = "use 'gtd reopen' to move back to NEW or 'gtd in-progress' to start work"
	case models.StateInvalid:
		helpMsg = "invalid tasks cannot be transitioned to other states"
	}

	return fmt.Errorf("cannot transition from %s to %s (%s)", currentState, newState, helpMsg)
}