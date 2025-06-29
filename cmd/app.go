package cmd

import (
	"fmt"

	"github.com/zw3rk/gtd/internal/config"
	"github.com/zw3rk/gtd/internal/database"
	"github.com/zw3rk/gtd/internal/git"
	"github.com/zw3rk/gtd/internal/models"
	"github.com/zw3rk/gtd/internal/services"
)

// App encapsulates all application dependencies
type App struct {
	config  *config.Config
	db      *database.Database
	repo    *models.TaskRepository
	service services.TaskService
}

// NewApp creates a new application instance
func NewApp() *App {
	return &App{
		config: config.NewConfig(),
	}
}

// Initialize sets up the application dependencies
func (a *App) Initialize() error {
	// Load configuration from environment
	if err := a.config.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find git root
	gitRoot, err := git.FindGitRoot(".")
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}
	a.config.GitRoot = gitRoot

	// Open database
	dbPath := a.config.GetDatabasePath()
	a.db, err = database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create schema if needed
	if err := a.db.CreateSchema(); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Create repository
	a.repo = models.NewTaskRepository(a.db)

	// Create service
	a.service = services.NewTaskService(a.repo)

	return nil
}

// Close cleans up application resources
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Repository returns the task repository
func (a *App) Repository() *models.TaskRepository {
	return a.repo
}

// Service returns the task service
func (a *App) Service() services.TaskService {
	return a.service
}

// Config returns the application configuration
func (a *App) Config() *config.Config {
	return a.config
}