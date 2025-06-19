package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/zw3rk/claude-gtd/internal/models"
)

// List command flags
type listFlags struct {
	oneline  bool
	all      bool
	state    string
	priority string
	kind     string
	tag      string
	blocked  bool
	limit    int
}

// newListCommand creates the list command
func newListCommand() *cobra.Command {
	var flags listFlags
	
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long: `List tasks with various filtering options.
By default, shows top 20 tasks (IN_PROGRESS first, then NEW), excluding DONE and CANCELLED tasks.`,
		Example: `  claude-gtd list
  claude-gtd list --oneline
  claude-gtd list --all
  claude-gtd list --state NEW --priority high
  claude-gtd list --kind bug --tag backend
  claude-gtd list --blocked`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate filters
			if err := validateListFlags(&flags); err != nil {
				return err
			}
			
			// Build list options
			opts := models.ListOptions{
				State:         flags.state,
				Priority:      flags.priority,
				Kind:          flags.kind,
				Tag:           flags.tag,
				Blocked:       flags.blocked,
				All:           flags.all,
				Limit:         flags.limit,
				ShowDone:      flags.all || flags.state == models.StateDone,
				ShowCancelled: flags.all || flags.state == models.StateCancelled,
			}
			
			// List tasks
			tasks, err := repo.List(opts)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}
			
			// Format and output
			formatTaskList(cmd.OutOrStdout(), tasks, flags.oneline)
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&flags.oneline, "oneline", false, "Show tasks in compact format")
	cmd.Flags().BoolVar(&flags.all, "all", false, "Show all tasks including DONE and CANCELLED")
	cmd.Flags().StringVar(&flags.state, "state", "", "Filter by state (NEW, IN_PROGRESS, DONE, CANCELLED)")
	cmd.Flags().StringVar(&flags.priority, "priority", "", "Filter by priority (high, medium, low)")
	cmd.Flags().StringVar(&flags.kind, "kind", "", "Filter by kind (bug, feature, regression)")
	cmd.Flags().StringVar(&flags.tag, "tag", "", "Filter by tag")
	cmd.Flags().BoolVar(&flags.blocked, "blocked", false, "Show only blocked tasks")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "Maximum number of tasks to show")
	
	return cmd
}

// newListDoneCommand creates the list-done command
func newListDoneCommand() *cobra.Command {
	var oneline bool
	
	cmd := &cobra.Command{
		Use:   "list-done",
		Short: "List completed tasks",
		Long:  `List completed tasks, most recent first.`,
		Example: `  claude-gtd list-done
  claude-gtd list-done --oneline`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := models.ListOptions{
				State:      models.StateDone,
				ShowDone:   true,
				All:        true,
			}
			
			tasks, err := repo.List(opts)
			if err != nil {
				return fmt.Errorf("failed to list done tasks: %w", err)
			}
			
			formatTaskList(cmd.OutOrStdout(), tasks, oneline)
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&oneline, "oneline", false, "Show tasks in compact format")
	
	return cmd
}

// newListCancelledCommand creates the list-cancelled command
func newListCancelledCommand() *cobra.Command {
	var oneline bool
	
	cmd := &cobra.Command{
		Use:   "list-cancelled",
		Short: "List cancelled tasks",
		Long:  `List cancelled tasks.`,
		Example: `  claude-gtd list-cancelled
  claude-gtd list-cancelled --oneline`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := models.ListOptions{
				State:         models.StateCancelled,
				ShowCancelled: true,
				All:           true,
			}
			
			tasks, err := repo.List(opts)
			if err != nil {
				return fmt.Errorf("failed to list cancelled tasks: %w", err)
			}
			
			formatTaskList(cmd.OutOrStdout(), tasks, oneline)
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&oneline, "oneline", false, "Show tasks in compact format")
	
	return cmd
}

// validateListFlags validates the list command flags
func validateListFlags(flags *listFlags) error {
	// Validate state
	if flags.state != "" {
		switch flags.state {
		case models.StateNew, models.StateInProgress, models.StateDone, models.StateCancelled:
			// valid
		default:
			return fmt.Errorf("invalid state: %s (must be NEW, IN_PROGRESS, DONE, or CANCELLED)", flags.state)
		}
	}
	
	// Validate priority
	if flags.priority != "" {
		switch flags.priority {
		case models.PriorityHigh, models.PriorityMedium, models.PriorityLow:
			// valid
		default:
			return fmt.Errorf("invalid priority: %s (must be high, medium, or low)", flags.priority)
		}
	}
	
	// Validate kind
	if flags.kind != "" {
		switch flags.kind {
		case "bug", "feature", "regression":
			// Convert to uppercase for the query
			if flags.kind == "bug" {
				flags.kind = models.KindBug
			} else if flags.kind == "feature" {
				flags.kind = models.KindFeature
			} else if flags.kind == "regression" {
				flags.kind = models.KindRegression
			}
		default:
			return fmt.Errorf("invalid kind: %s (must be bug, feature, or regression)", flags.kind)
		}
	}
	
	return nil
}

// formatTaskList formats and outputs a list of tasks
func formatTaskList(w io.Writer, tasks []*models.Task, oneline bool) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks found.")
		return
	}
	
	for i, task := range tasks {
		if oneline {
			fmt.Fprintln(w, formatTaskOneline(task))
		} else {
			// Get subtask stats for the task
			var stats *SubtaskStats
			if task.Parent == nil { // Only get subtasks for parent tasks
				subtasks, err := repo.GetChildren(task.ID)
				if err == nil && len(subtasks) > 0 {
					stats = &SubtaskStats{Total: len(subtasks)}
					for _, st := range subtasks {
						if st.State == models.StateDone {
							stats.Done++
						}
					}
				}
			}
			
			// Use git-style format
			fmt.Fprint(w, formatTaskGitStyle(task, stats))
			// Add blank line between tasks
			if i < len(tasks)-1 {
				fmt.Fprintln(w)
			}
		}
	}
	
	// Show count at the end
	fmt.Fprintf(w, "\n%s\n", formatTaskCount(len(tasks), "task"))
}