package errors

import (
	"fmt"
	"strings"
)

// TaskNotFoundError provides helpful suggestions when a task is not found
type TaskNotFoundError struct {
	ID          string
	Suggestions []string
}

func (e *TaskNotFoundError) Error() string {
	msg := fmt.Sprintf("task not found: %s", e.ID)
	
	if len(e.Suggestions) > 0 {
		if len(e.Suggestions) == 1 {
			msg += fmt.Sprintf("\n\nDid you mean: %s?", e.Suggestions[0])
		} else {
			msg += "\n\nDid you mean one of these?"
			for _, s := range e.Suggestions {
				msg += fmt.Sprintf("\n  - %s", s)
			}
		}
	}
	
	return msg
}

// Task interface to avoid circular imports
type Task interface {
	GetID() string
	GetTitle() string
	ShortHash() string
}

// NewTaskNotFoundError creates a new error with suggestions
func NewTaskNotFoundError(id string, allTasks []Task) error {
	suggestions := findSimilarTaskIDs(id, allTasks)
	
	if len(suggestions) == 0 {
		// No similar tasks found, provide general help
		return fmt.Errorf("task not found: %s\n\nHint: Use 'gtd list' to see available tasks", id)
	}
	
	return &TaskNotFoundError{
		ID:          id,
		Suggestions: suggestions,
	}
}

// findSimilarTaskIDs finds task IDs that are similar to the given ID
func findSimilarTaskIDs(id string, tasks []Task) []string {
	var suggestions []string
	idLower := strings.ToLower(id)
	
	// First, check for exact prefix matches
	for _, task := range tasks {
		taskID := task.GetID()
		if strings.HasPrefix(taskID, id) || strings.HasPrefix(strings.ToLower(taskID), idLower) {
			suggestions = append(suggestions, fmt.Sprintf("%s (%s)", task.ShortHash(), task.GetTitle()))
			if len(suggestions) >= 3 {
				return suggestions
			}
		}
	}
	
	// If we have exact prefix matches, return them
	if len(suggestions) > 0 {
		return suggestions
	}
	
	// Check for partial matches anywhere in the ID
	for _, task := range tasks {
		taskID := task.GetID()
		if strings.Contains(taskID, id) || strings.Contains(strings.ToLower(taskID), idLower) {
			suggestions = append(suggestions, fmt.Sprintf("%s (%s)", task.ShortHash(), task.GetTitle()))
			if len(suggestions) >= 3 {
				return suggestions
			}
		}
	}
	
	// Check for similar task titles
	for _, task := range tasks {
		if strings.Contains(strings.ToLower(task.GetTitle()), idLower) {
			suggestions = append(suggestions, fmt.Sprintf("%s (%s)", task.ShortHash(), task.GetTitle()))
			if len(suggestions) >= 3 {
				return suggestions
			}
		}
	}
	
	return suggestions
}

// InvalidStateTransitionError provides helpful guidance for state transitions
type InvalidStateTransitionError struct {
	CurrentState string
	TargetState  string
	ValidStates  []string
	Commands     map[string]string // Maps target states to commands
}

func (e *InvalidStateTransitionError) Error() string {
	msg := fmt.Sprintf("cannot transition from %s to %s", e.CurrentState, e.TargetState)
	
	if len(e.ValidStates) > 0 {
		msg += "\n\nValid transitions from " + e.CurrentState + ":"
		for _, state := range e.ValidStates {
			if cmd, ok := e.Commands[state]; ok {
				msg += fmt.Sprintf("\n  - %s (use 'gtd %s')", state, cmd)
			} else {
				msg += fmt.Sprintf("\n  - %s", state)
			}
		}
	}
	
	return msg
}

// State constants to avoid circular imports
const (
	StateInbox      = "INBOX"
	StateNew        = "NEW"
	StateInProgress = "IN_PROGRESS"
	StateDone       = "DONE"
	StateCancelled  = "CANCELLED"
	StateInvalid    = "INVALID"
)

// NewInvalidStateTransitionError creates a new error with helpful transition guidance
func NewInvalidStateTransitionError(currentState, targetState string) error {
	commands := map[string]map[string]string{
		StateInbox: {
			StateNew:     "accept",
			StateInvalid: "reject",
		},
		StateNew: {
			StateInProgress: "in-progress",
			StateDone:       "done",
			StateCancelled:  "cancel",
		},
		StateInProgress: {
			StateDone:      "done",
			StateCancelled: "cancel",
		},
		StateDone: {
			StateInProgress: "in-progress",
		},
		StateCancelled: {
			StateNew:        "reopen",
			StateInProgress: "reopen",
		},
	}
	
	validTransitions := map[string][]string{
		StateInbox:      {StateNew, StateInvalid},
		StateNew:        {StateInProgress, StateDone, StateCancelled},
		StateInProgress: {StateDone, StateCancelled},
		StateDone:       {StateInProgress},
		StateCancelled:  {StateNew, StateInProgress},
		StateInvalid:    {}, // No valid transitions from INVALID
	}
	
	validStates := validTransitions[currentState]
	stateCommands := commands[currentState]
	
	return &InvalidStateTransitionError{
		CurrentState: currentState,
		TargetState:  targetState,
		ValidStates:  validStates,
		Commands:     stateCommands,
	}
}

// InvalidCommandError provides suggestions for mistyped commands
type InvalidCommandError struct {
	Command     string
	Suggestions []string
}

func (e *InvalidCommandError) Error() string {
	msg := fmt.Sprintf("unknown command: %s", e.Command)
	
	if len(e.Suggestions) > 0 {
		if len(e.Suggestions) == 1 {
			msg += fmt.Sprintf("\n\nDid you mean: %s?", e.Suggestions[0])
		} else {
			msg += "\n\nDid you mean one of these?"
			for _, s := range e.Suggestions {
				msg += fmt.Sprintf("\n  - %s", s)
			}
		}
	}
	
	msg += "\n\nRun 'gtd --help' to see all available commands"
	
	return msg
}

// FindSimilarCommands finds commands similar to the given input
func FindSimilarCommands(input string, availableCommands []string) []string {
	var suggestions []string
	inputLower := strings.ToLower(input)
	
	// Check for prefix matches
	for _, cmd := range availableCommands {
		if strings.HasPrefix(cmd, inputLower) {
			suggestions = append(suggestions, cmd)
		}
	}
	
	// Check for partial matches
	if len(suggestions) == 0 {
		for _, cmd := range availableCommands {
			if strings.Contains(cmd, inputLower) || strings.Contains(inputLower, cmd) {
				suggestions = append(suggestions, cmd)
			}
		}
	}
	
	// Check for Levenshtein distance of 1 or 2
	if len(suggestions) == 0 {
		for _, cmd := range availableCommands {
			if levenshteinDistance(inputLower, cmd) <= 2 {
				suggestions = append(suggestions, cmd)
			}
		}
	}
	
	// Limit to top 3 suggestions
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}
	
	return suggestions
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}
	
	// Initialize first column and row
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}
	
	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len(s1)][len(s2)]
}

func min(nums ...int) int {
	result := nums[0]
	for _, n := range nums[1:] {
		if n < result {
			result = n
		}
	}
	return result
}