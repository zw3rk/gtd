// Package cmd implements the CLI commands for gtd
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zw3rk/gtd/internal/database"
	"github.com/zw3rk/gtd/internal/git"
	"github.com/zw3rk/gtd/internal/models"
)

var (
	// Version is set at build time
	Version = "dev"

	// Global database and repository instances
	db   *database.Database
	repo *models.TaskRepository
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
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

			// Find git root
			gitRoot, err := git.FindGitRoot(".")
			if err != nil {
				return fmt.Errorf("not in a git repository: %w", err)
			}

			// Open database
			dbPath := filepath.Join(gitRoot, "claude-tasks.db")
			db, err = database.New(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}

			// Create schema if needed
			if err := db.CreateSchema(); err != nil {
				return fmt.Errorf("failed to create schema: %w", err)
			}

			// Create repository
			repo = models.NewTaskRepository(db)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Close database if it was opened
			if db != nil {
				return db.Close()
			}
			return nil
		},
	}

	// Add commands
	rootCmd.AddCommand(
		newAddBugCommand(),
		newAddFeatureCommand(),
		newAddRegressionCommand(),
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
	)

	return rootCmd
}

// Execute runs the root command
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
