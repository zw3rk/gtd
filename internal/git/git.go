// Package git provides utilities for working with git repositories
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindGitRoot searches for the nearest .git directory starting from the given path
// and traversing up the directory tree. It returns the absolute path to the
// directory containing .git, or an error if no git repository is found.
func FindGitRoot(startPath string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Start from the given path
	current := absPath
	
	// Keep going up until we find .git or reach the root
	for {
		// Check if .git exists in current directory
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return current, nil
		}
		
		// Get parent directory
		parent := filepath.Dir(current)
		
		// If we've reached the root, stop
		if parent == current {
			break
		}
		
		current = parent
	}
	
	return "", fmt.Errorf("not in a git repository (or any of the parent directories)")
}

// GetAuthor retrieves the git author name and email from git config
func GetAuthor() (string, error) {
	// Try to get user.name
	nameCmd := exec.Command("git", "config", "user.name")
	nameOut, err := nameCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user.name: %w", err)
	}
	name := strings.TrimSpace(string(nameOut))
	
	// Try to get user.email
	emailCmd := exec.Command("git", "config", "user.email")
	emailOut, err := emailCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user.email: %w", err)
	}
	email := strings.TrimSpace(string(emailOut))
	
	if name == "" || email == "" {
		return "", fmt.Errorf("git user.name and user.email must be configured")
	}
	
	// Format like git does: Name <email>
	return fmt.Sprintf("%s <%s>", name, email), nil
}