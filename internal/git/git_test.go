package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		startPath string
		want      string
		wantErr   bool
	}{
		{
			name: "finds .git in current directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			startPath: ".",
			wantErr:   false,
		},
		{
			name: "finds .git in parent directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				subDir := filepath.Join(tmpDir, "subdir")
				if err := os.Mkdir(subDir, 0755); err != nil {
					t.Fatal(err)
				}
				return subDir
			},
			startPath: ".",
			wantErr:   false,
		},
		{
			name: "finds .git multiple levels up",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				deepDir := filepath.Join(tmpDir, "a", "b", "c")
				if err := os.MkdirAll(deepDir, 0755); err != nil {
					t.Fatal(err)
				}
				return deepDir
			},
			startPath: ".",
			wantErr:   false,
		},
		{
			name: "returns error when no .git found",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				subDir := filepath.Join(tmpDir, "subdir")
				if err := os.Mkdir(subDir, 0755); err != nil {
					t.Fatal(err)
				}
				return subDir
			},
			startPath: ".",
			wantErr:   true,
		},
		{
			name: "handles absolute path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			startPath: "", // will be set to absolute path in test
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := tt.setup(t)
			
			// Change to test directory
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(oldDir)
			
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}
			
			// Use absolute path for the absolute path test
			startPath := tt.startPath
			if tt.name == "handles absolute path" {
				startPath = testDir
			}
			
			got, err := FindGitRoot(startPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindGitRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Verify the returned path contains .git
				gitPath := filepath.Join(got, ".git")
				if _, err := os.Stat(gitPath); os.IsNotExist(err) {
					t.Errorf("FindGitRoot() returned %v, but .git not found there", got)
				}
			}
		})
	}
}

func TestFindGitRootWithSymlink(t *testing.T) {
	// Test that FindGitRoot works correctly with symlinks
	tmpDir := t.TempDir()
	realGitDir := filepath.Join(tmpDir, "real-repo")
	if err := os.Mkdir(realGitDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	gitDir := filepath.Join(realGitDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create a symlink to the repo
	linkDir := filepath.Join(tmpDir, "link-to-repo")
	if err := os.Symlink(realGitDir, linkDir); err != nil {
		t.Skip("Symlinks not supported on this platform")
	}
	
	// Change to symlinked directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)
	
	if err := os.Chdir(linkDir); err != nil {
		t.Fatal(err)
	}
	
	got, err := FindGitRoot(".")
	if err != nil {
		t.Errorf("FindGitRoot() unexpected error = %v", err)
		return
	}
	
	// Should find the git root through the symlink
	if _, err := os.Stat(filepath.Join(got, ".git")); err != nil {
		t.Errorf("FindGitRoot() = %v, but .git not found there", got)
	}
}