// Package cmd implements the CLI commands for gtd
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/database"
	"github.com/zw3rk/gtd/internal/models"
)

var (
	// Version is set at build time
	Version = "dev"

	// Global database and repository instances - DEPRECATED: use App instead
	db   *database.Database
	repo *models.TaskRepository
)

// NewRootCommand creates the root command with the provided app instance
func NewRootCommand(app *App) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gtd",
		Short: "A SQLite-driven CLI task management tool",
		Long: `gtd is a task management tool following GTD methodology.
It stores tasks per-project in a claude-tasks.db file at the git repository root.`,
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip DB initialization for help and version commands
			if cmd.Name() == "help" || cmd.Name() == "version" {
				return nil
			}
			if cmd.Parent() != nil && cmd.Parent().Name() == "help" {
				return nil
			}

			// Initialize the app
			if err := app.Initialize(); err != nil {
				return err
			}

			// Set global variables for backward compatibility
			// TODO: Remove these once all commands are refactored
			db = app.db
			repo = app.repo

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return app.Close()
		},
	}

	// Add commands
	rootCmd.AddCommand(
		newAddCommand(),
		newAddSubtaskCommand(),
		newInProgressCommand(),
		newDoneCommand(),
		newCancelCommand(),
		newBlockCommand(),
		newUnblockCommand(),
		newListCommand(),
		newListDoneCommand(),
		newListCancelledCommand(),
		newShowCommand(),
		newSearchCommand(),
		newSummaryCommand(),
		newExportCommand(),
		newReviewCommand(),
		newAcceptCommand(),
		newRejectCommand(),
		newReopenCommand(),
	)

	return rootCmd
}

// Execute runs the root command
func Execute() {
	app := NewApp()
	if err := NewRootCommand(app).Execute(); err != nil {
		os.Exit(1)
	}
}
