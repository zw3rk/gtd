// Package models defines the data structures for claude-gtd
package models

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"strings"
	"time"
	
	"github.com/zw3rk/gtd/internal/git"
)

// Task kinds
const (
	KindBug        = "BUG"
	KindFeature    = "FEATURE"
	KindRegression = "REGRESSION"
)

// Task priorities
const (
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)

// Task states
const (
	StateNew        = "NEW"
	StateInProgress = "IN_PROGRESS"
	StateDone       = "DONE"
	StateCancelled  = "CANCELLED"
)

// Task represents a task in the system
type Task struct {
	ID          string    `json:"id"`
	Parent      *string   `json:"parent,omitempty"`
	Priority    string    `json:"priority"`
	State       string    `json:"state"`
	Kind        string    `json:"kind"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Author      string    `json:"author"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Source      string    `json:"source,omitempty"`
	BlockedBy   *string   `json:"blocked_by,omitempty"`
	Tags        string    `json:"tags,omitempty"`
}

// NewTask creates a new task with default values
func NewTask(kind, title, description string) *Task {
	now := time.Now()
	
	// Get author from git config
	author, err := git.GetAuthor()
	if err != nil {
		// Fallback to a default if git config is not available
		author = "Unknown <unknown@example.com>"
	}
	
	task := &Task{
		Kind:        kind,
		Title:       title,
		Description: description,
		Author:      author,
		Priority:    PriorityMedium,
		State:       StateNew,
		Created:     now,
		Updated:     now,
	}
	// Generate hash ID based on content and timestamp
	task.ID = generateTaskHash(kind, title, description, now)
	return task
}

// Validate checks if the task has valid field values
func (t *Task) Validate() error {
	// Title is required
	if strings.TrimSpace(t.Title) == "" {
		return fmt.Errorf("title is required")
	}

	// Description is required
	if strings.TrimSpace(t.Description) == "" {
		return fmt.Errorf("description is required - tasks must have a body explaining the work")
	}

	// Validate kind
	switch t.Kind {
	case KindBug, KindFeature, KindRegression:
		// valid
	default:
		return fmt.Errorf("invalid kind: %s", t.Kind)
	}

	// Validate priority
	switch t.Priority {
	case PriorityHigh, PriorityMedium, PriorityLow:
		// valid
	default:
		return fmt.Errorf("invalid priority: %s", t.Priority)
	}

	// Validate state
	switch t.State {
	case StateNew, StateInProgress, StateDone, StateCancelled:
		// valid
	default:
		return fmt.Errorf("invalid state: %s", t.State)
	}

	return nil
}

// CanTransitionTo checks if the task can transition to the given state
func (t *Task) CanTransitionTo(newState string, children []*Task) bool {
	// First check if parent task can be marked as DONE
	if newState == StateDone && len(children) > 0 {
		// Parent can only be DONE if all children are DONE or CANCELLED
		for _, child := range children {
			if child.State != StateDone && child.State != StateCancelled {
				return false
			}
		}
	}

	// Check basic state transitions
	switch t.State {
	case StateNew:
		// Can transition to any state from NEW (after parent check above)
		return true
	case StateInProgress:
		// Can transition to DONE or CANCELLED
		if newState == StateNew {
			return false
		}
	case StateDone:
		// Can only transition back to IN_PROGRESS
		if newState != StateInProgress {
			return false
		}
	case StateCancelled:
		// Can transition to NEW or IN_PROGRESS
		if newState == StateDone {
			return false
		}
	}

	return true
}

// IsBlocked returns true if the task is blocked by another task
func (t *Task) IsBlocked() bool {
	return t.BlockedBy != nil
}

// ParseTags returns the tags as a slice of strings
func (t *Task) ParseTags() []string {
	if t.Tags == "" {
		return []string{}
	}

	parts := strings.Split(t.Tags, ",")
	tags := make([]string, 0, len(parts))
	for _, tag := range parts {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// SetTags sets the tags from a slice of strings
func (t *Task) SetTags(tags []string) {
	t.Tags = strings.Join(tags, ",")
}

// generateTaskHash creates a unique hash for a task based on its content
func generateTaskHash(kind, title, description string, created time.Time) string {
	// Create a hash based on content and timestamp to ensure uniqueness
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%s%s%s%d%d", kind, title, description, created.Unix(), rand.Int63())))
	// Return full 40-character SHA-1 hash like git
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ShortHash returns the first 7 characters of the hash (like git)
func (t *Task) ShortHash() string {
	if len(t.ID) >= 7 {
		return t.ID[:7]
	}
	return t.ID
}