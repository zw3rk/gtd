// Package git provides utilities for working with git repositories
package git

import (
	"fmt"
	"os"
	"path/filepath"
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