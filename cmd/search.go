package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// newSearchCommand creates the search command
func newSearchCommand() *cobra.Command {
	var oneline bool

	cmd := &cobra.Command{
		Use:   "search QUERY",
		Short: "Search tasks",
		Long: `Search for tasks by looking in title and description fields.
The search is case-insensitive and matches partial words.`,
		Example: `  claude-gtd search "memory leak"
  claude-gtd search database
  claude-gtd search --oneline connection`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Join all args to form the search query
			query := strings.Join(args, " ")

			// Search tasks
			tasks, err := repo.Search(query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Format and output
			if len(tasks) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No tasks found.")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Search results for: %q\n", query)
				fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("=", 50))
				fmt.Fprintln(cmd.OutOrStdout())

				formatTaskList(cmd.OutOrStdout(), tasks, oneline)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&oneline, "oneline", false, "Show results in compact format")

	return cmd
}
