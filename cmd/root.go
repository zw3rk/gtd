// Package cmd implements the CLI commands for claude-gtd
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zw3rk/claude-gtd/internal/database"
	"github.com/zw3rk/claude-gtd/internal/git"
	"github.com/zw3rk/claude-gtd/internal/models"
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
		Use:   "claude-gtd",
		Short: "A SQLite-driven CLI task management tool",
		Long: `claude-gtd is a task management tool following GTD methodology.
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
	)
	
	return rootCmd
}

// Execute runs the root command
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

// Placeholder commands - to be implemented

func newAddSubtaskCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-subtask",
		Short: "Add a subtask to an existing task",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newInProgressCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "in-progress",
		Short: "Mark a task as in progress",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newDoneCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "done",
		Short: "Mark a task as done",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newCancelCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newBlockCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "block",
		Short: "Mark a task as blocked by another task",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newUnblockCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unblock",
		Short: "Remove blocking status from a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newListDoneCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-done",
		Short: "List completed tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newListCancelledCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-cancelled",
		Short: "List cancelled tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show task details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Search tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newSummaryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show task summary statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func newExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export tasks to various formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}