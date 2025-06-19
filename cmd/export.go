package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/models"
)

// newExportCommand creates the export command
func newExportCommand() *cobra.Command {
	var (
		format         string
		outputFile     string
		activeOnly     bool
		stateFilter    string
		priorityFilter string
		kindFilter     string
		tagFilter      string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export tasks to various formats",
		Long: `Export tasks to JSON, CSV, or Markdown format.
Tasks can be filtered by state, priority, kind, or tags before export.`,
		Example: `  claude-gtd export --format json
  claude-gtd export --format csv --output tasks.csv
  claude-gtd export --format markdown --active
  claude-gtd export --format json --state done --kind bug`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate format
			format = strings.ToLower(format)
			if format != "json" && format != "csv" && format != "markdown" {
				return fmt.Errorf("unsupported format: %s", format)
			}

			// Build list options
			opts := models.ListOptions{
				All:           true,
				ShowDone:      true,
				ShowCancelled: true,
			}

			if activeOnly {
				opts.ShowDone = false
				opts.ShowCancelled = false
			}

			if stateFilter != "" {
				state := strings.ToUpper(stateFilter)
				if state == "IN_PROGRESS" || state == "IN-PROGRESS" {
					state = models.StateInProgress
				}
				if state != models.StateNew && state != models.StateInProgress &&
					state != models.StateDone && state != models.StateCancelled {
					return fmt.Errorf("invalid state: %s", stateFilter)
				}
				opts.State = state
			}

			if priorityFilter != "" {
				priority := strings.ToLower(priorityFilter)
				if priority != models.PriorityHigh && priority != models.PriorityMedium &&
					priority != models.PriorityLow {
					return fmt.Errorf("invalid priority: %s", priorityFilter)
				}
				opts.Priority = priority
			}

			if kindFilter != "" {
				kind := strings.ToUpper(kindFilter)
				if kind != models.KindBug && kind != models.KindFeature &&
					kind != models.KindRegression {
					return fmt.Errorf("invalid kind: %s", kindFilter)
				}
				opts.Kind = kind
			}

			if tagFilter != "" {
				opts.Tag = tagFilter
			}

			// Get tasks
			tasks, err := repo.List(opts)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			// Determine output writer
			var writer io.Writer
			if outputFile != "" {
				file, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer file.Close()
				writer = file
			} else {
				writer = cmd.OutOrStdout()
			}

			// Export based on format
			switch format {
			case "json":
				if err := exportJSON(writer, tasks); err != nil {
					return fmt.Errorf("failed to export JSON: %w", err)
				}
			case "csv":
				if err := exportCSV(writer, tasks); err != nil {
					return fmt.Errorf("failed to export CSV: %w", err)
				}
			case "markdown":
				if err := exportMarkdown(writer, tasks); err != nil {
					return fmt.Errorf("failed to export Markdown: %w", err)
				}
			}

			// Show success message if writing to file
			if outputFile != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Exported %d tasks to %s\n", len(tasks), outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "json", "Export format (json, csv, markdown)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolVar(&activeOnly, "active", false, "Export only active tasks (exclude DONE and CANCELLED)")
	cmd.Flags().StringVar(&stateFilter, "state", "", "Filter by state (new, in_progress, done, cancelled)")
	cmd.Flags().StringVar(&priorityFilter, "priority", "", "Filter by priority (high, medium, low)")
	cmd.Flags().StringVar(&kindFilter, "kind", "", "Filter by kind (bug, feature, regression)")
	cmd.Flags().StringVar(&tagFilter, "tag", "", "Filter by tag")

	return cmd
}

// exportJSON exports tasks as JSON
func exportJSON(w io.Writer, tasks []*models.Task) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	// Convert tasks to a format that includes all fields
	type exportTask struct {
		ID          string  `json:"id"`
		Kind        string  `json:"kind"`
		State       string  `json:"state"`
		Priority    string  `json:"priority"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Tags        string  `json:"tags"`
		Source      string  `json:"source"`
		Parent      *string `json:"parent,omitempty"`
		BlockedBy   *string `json:"blocked_by,omitempty"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
	}

	exportTasks := make([]exportTask, len(tasks))
	for i, task := range tasks {
		exportTasks[i] = exportTask{
			ID:          task.ID,
			Kind:        task.Kind,
			State:       task.State,
			Priority:    task.Priority,
			Title:       task.Title,
			Description: task.Description,
			Tags:        task.Tags,
			Source:      task.Source,
			Parent:      task.Parent,
			BlockedBy:   task.BlockedBy,
			CreatedAt:   task.Created.Format("2006-01-02 15:04:05"),
			UpdatedAt:   task.Updated.Format("2006-01-02 15:04:05"),
		}
	}

	return encoder.Encode(exportTasks)
}

// exportCSV exports tasks as CSV
func exportCSV(w io.Writer, tasks []*models.Task) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Type", "State", "Priority", "Title", "Tags", "Source", "Parent", "BlockedBy", "Created", "Updated"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, task := range tasks {
		parentStr := ""
		if task.Parent != nil {
			parentStr = *task.Parent
		}

		blockedByStr := ""
		if task.BlockedBy != nil {
			blockedByStr = *task.BlockedBy
		}

		row := []string{
			task.ID,
			task.Kind,
			task.State,
			task.Priority,
			task.Title,
			task.Tags,
			task.Source,
			parentStr,
			blockedByStr,
			task.Created.Format("2006-01-02 15:04:05"),
			task.Updated.Format("2006-01-02 15:04:05"),
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// exportMarkdown exports tasks as Markdown
func exportMarkdown(w io.Writer, tasks []*models.Task) error {
	fmt.Fprintln(w, "# Tasks Export")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Total tasks: %d\n", len(tasks))
	fmt.Fprintln(w)

	// Table header
	fmt.Fprintln(w, "| ID | Type | State | Priority | Title | Tags | Source | Parent | Blocked By |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|")

	// Table rows
	for _, task := range tasks {
		parentStr := "-"
		if task.Parent != nil {
			parentStr = fmt.Sprintf("#%s", *task.Parent)
		}

		blockedByStr := "-"
		if task.BlockedBy != nil {
			blockedByStr = fmt.Sprintf("#%s", *task.BlockedBy)
		}

		tagsStr := "-"
		if task.Tags != "" {
			tagsStr = task.Tags
		}

		sourceStr := "-"
		if task.Source != "" {
			sourceStr = task.Source
		}

		fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			task.ID,
			task.Kind,
			task.State,
			task.Priority,
			task.Title,
			tagsStr,
			sourceStr,
			parentStr,
			blockedByStr,
		)
	}

	// Add detailed task descriptions
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Task Details")
	fmt.Fprintln(w)

	for _, task := range tasks {
		fmt.Fprintf(w, "### #%s: %s\n", task.ID, task.Title)
		fmt.Fprintln(w)

		if task.Description != "" {
			fmt.Fprintln(w, task.Description)
			fmt.Fprintln(w)
		}

		fmt.Fprintf(w, "- **Type:** %s\n", formatKind(task.Kind))
		fmt.Fprintf(w, "- **State:** %s %s\n", task.State, getStateEmoji(task.State))
		fmt.Fprintf(w, "- **Priority:** %s %s\n", task.Priority, getPriorityEmoji(task.Priority))

		if task.Tags != "" {
			fmt.Fprintf(w, "- **Tags:** %s\n", task.Tags)
		}

		if task.Source != "" {
			fmt.Fprintf(w, "- **Source:** %s\n", task.Source)
		}

		if task.Parent != nil {
			fmt.Fprintf(w, "- **Parent:** #%s\n", *task.Parent)
		}

		if task.BlockedBy != nil {
			fmt.Fprintf(w, "- **Blocked by:** #%s\n", *task.BlockedBy)
		}

		fmt.Fprintf(w, "- **Created:** %s\n", task.Created.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(w, "- **Updated:** %s\n", task.Updated.Format("2006-01-02 15:04:05"))
		fmt.Fprintln(w)
	}

	return nil
}
