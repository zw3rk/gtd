package models

import (
	"strings"
	"testing"
	"time"
)

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bug task",
			task: Task{
				Kind:     KindBug,
				Title:    "Fix memory leak",
				Priority: PriorityMedium,
				State:    StateNew,
			},
			wantErr: false,
		},
		{
			name: "valid feature task",
			task: Task{
				Kind:        KindFeature,
				Title:       "Add dark mode",
				Description: "Implement dark mode theme",
				Priority:    PriorityHigh,
				State:       StateNew,
			},
			wantErr: false,
		},
		{
			name: "missing title",
			task: Task{
				Kind:     KindBug,
				Priority: PriorityMedium,
				State:    StateNew,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "empty title",
			task: Task{
				Kind:     KindBug,
				Title:    "   ",
				Priority: PriorityMedium,
				State:    StateNew,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "invalid kind",
			task: Task{
				Kind:     "INVALID",
				Title:    "Test task",
				Priority: PriorityMedium,
				State:    StateNew,
			},
			wantErr: true,
			errMsg:  "invalid kind",
		},
		{
			name: "invalid priority",
			task: Task{
				Kind:     KindBug,
				Title:    "Test task",
				Priority: "urgent",
				State:    StateNew,
			},
			wantErr: true,
			errMsg:  "invalid priority",
		},
		{
			name: "invalid state",
			task: Task{
				Kind:     KindBug,
				Title:    "Test task",
				Priority: PriorityMedium,
				State:    "PENDING",
			},
			wantErr: true,
			errMsg:  "invalid state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Task.Validate() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestTaskCanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		canMove  bool
		hasChild bool
		childState string
	}{
		// NEW state transitions
		{
			name:    "NEW to IN_PROGRESS",
			from:    StateNew,
			to:      StateInProgress,
			canMove: true,
		},
		{
			name:    "NEW to DONE",
			from:    StateNew,
			to:      StateDone,
			canMove: true,
		},
		{
			name:    "NEW to CANCELLED",
			from:    StateNew,
			to:      StateCancelled,
			canMove: true,
		},
		// IN_PROGRESS state transitions
		{
			name:    "IN_PROGRESS to DONE",
			from:    StateInProgress,
			to:      StateDone,
			canMove: true,
		},
		{
			name:    "IN_PROGRESS to CANCELLED",
			from:    StateInProgress,
			to:      StateCancelled,
			canMove: true,
		},
		{
			name:    "IN_PROGRESS to NEW not allowed",
			from:    StateInProgress,
			to:      StateNew,
			canMove: false,
		},
		// DONE state transitions
		{
			name:    "DONE to IN_PROGRESS",
			from:    StateDone,
			to:      StateInProgress,
			canMove: true,
		},
		{
			name:    "DONE to CANCELLED not allowed",
			from:    StateDone,
			to:      StateCancelled,
			canMove: false,
		},
		// CANCELLED state transitions
		{
			name:    "CANCELLED to NEW",
			from:    StateCancelled,
			to:      StateNew,
			canMove: true,
		},
		{
			name:    "CANCELLED to IN_PROGRESS",
			from:    StateCancelled,
			to:      StateInProgress,
			canMove: true,
		},
		// Parent with children
		{
			name:       "Parent can't be DONE with NEW child",
			from:       StateInProgress,
			to:         StateDone,
			hasChild:   true,
			childState: StateNew,
			canMove:    false,
		},
		{
			name:       "Parent can be DONE with DONE child",
			from:       StateInProgress,
			to:         StateDone,
			hasChild:   true,
			childState: StateDone,
			canMove:    true,
		},
		{
			name:       "Parent can be DONE with CANCELLED child",
			from:       StateInProgress,
			to:         StateDone,
			hasChild:   true,
			childState: StateCancelled,
			canMove:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				State: tt.from,
			}
			
			var children []*Task
			if tt.hasChild {
				children = []*Task{{State: tt.childState}}
			}
			
			got := task.CanTransitionTo(tt.to, children)
			if got != tt.canMove {
				t.Errorf("Task.CanTransitionTo() = %v, want %v", got, tt.canMove)
			}
		})
	}
}

func TestTaskIsBlocked(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		blocked bool
	}{
		{
			name:    "not blocked",
			task:    Task{BlockedBy: nil},
			blocked: false,
		},
		{
			name:    "blocked",
			task:    Task{BlockedBy: intPtr(42)},
			blocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.IsBlocked(); got != tt.blocked {
				t.Errorf("Task.IsBlocked() = %v, want %v", got, tt.blocked)
			}
		})
	}
}

func TestTaskParseTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     string
		expected []string
	}{
		{
			name:     "empty tags",
			tags:     "",
			expected: []string{},
		},
		{
			name:     "single tag",
			tags:     "backend",
			expected: []string{"backend"},
		},
		{
			name:     "multiple tags",
			tags:     "backend,urgent,database",
			expected: []string{"backend", "urgent", "database"},
		},
		{
			name:     "tags with spaces",
			tags:     "backend, urgent , database",
			expected: []string{"backend", "urgent", "database"},
		},
		{
			name:     "empty elements",
			tags:     "backend,,database",
			expected: []string{"backend", "database"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Tags: tt.tags}
			got := task.ParseTags()
			
			if len(got) != len(tt.expected) {
				t.Errorf("Task.ParseTags() returned %d tags, want %d", len(got), len(tt.expected))
				return
			}
			
			for i, tag := range got {
				if tag != tt.expected[i] {
					t.Errorf("Task.ParseTags()[%d] = %v, want %v", i, tag, tt.expected[i])
				}
			}
		})
	}
}

func TestTaskSetTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "empty tags",
			tags:     []string{},
			expected: "",
		},
		{
			name:     "single tag",
			tags:     []string{"backend"},
			expected: "backend",
		},
		{
			name:     "multiple tags",
			tags:     []string{"backend", "urgent", "database"},
			expected: "backend,urgent,database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{}
			task.SetTags(tt.tags)
			if task.Tags != tt.expected {
				t.Errorf("Task.SetTags() resulted in Tags = %v, want %v", task.Tags, tt.expected)
			}
		})
	}
}

// Helper function for tests
func intPtr(i int) *int {
	return &i
}

func TestNewTask(t *testing.T) {
	now := time.Now()
	task := NewTask(KindBug, "Fix memory leak", "Memory usage grows over time")
	
	if task.Kind != KindBug {
		t.Errorf("NewTask() Kind = %v, want %v", task.Kind, KindBug)
	}
	if task.Title != "Fix memory leak" {
		t.Errorf("NewTask() Title = %v, want %v", task.Title, "Fix memory leak")
	}
	if task.Description != "Memory usage grows over time" {
		t.Errorf("NewTask() Description = %v, want %v", task.Description, "Memory usage grows over time")
	}
	if task.Priority != PriorityMedium {
		t.Errorf("NewTask() Priority = %v, want %v", task.Priority, PriorityMedium)
	}
	if task.State != StateNew {
		t.Errorf("NewTask() State = %v, want %v", task.State, StateNew)
	}
	if task.Created.Before(now) || task.Created.After(now.Add(time.Second)) {
		t.Errorf("NewTask() Created time not within expected range")
	}
	if task.Updated.Before(now) || task.Updated.After(now.Add(time.Second)) {
		t.Errorf("NewTask() Updated time not within expected range")
	}
}