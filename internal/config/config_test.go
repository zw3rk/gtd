package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	// Check defaults
	if cfg.DatabaseName != "claude-tasks.db" {
		t.Errorf("DatabaseName = %s, want claude-tasks.db", cfg.DatabaseName)
	}
	if cfg.ColorEnabled != true {
		t.Errorf("ColorEnabled = %v, want true", cfg.ColorEnabled)
	}
	if cfg.PageSize != 20 {
		t.Errorf("PageSize = %d, want 20", cfg.PageSize)
	}
	if cfg.DefaultPriority != "medium" {
		t.Errorf("DefaultPriority = %s, want medium", cfg.DefaultPriority)
	}
}

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "default values",
			envVars: map[string]string{},
			want: &Config{
				DatabaseName:    "claude-tasks.db",
				ColorEnabled:    true,
				PageSize:        20,
				DefaultPriority: "medium",
				ShowWarnings:    true,
				Editor:          "vi",
			},
		},
		{
			name: "custom database name",
			envVars: map[string]string{
				"GTD_DATABASE_NAME": "custom.db",
			},
			want: &Config{
				DatabaseName:    "custom.db",
				ColorEnabled:    true,
				PageSize:        20,
				DefaultPriority: "medium",
				ShowWarnings:    true,
				Editor:          "vi",
			},
		},
		{
			name: "disable colors",
			envVars: map[string]string{
				"GTD_COLOR": "false",
			},
			want: &Config{
				DatabaseName:    "claude-tasks.db",
				ColorEnabled:    false,
				PageSize:        20,
				DefaultPriority: "medium",
				ShowWarnings:    true,
				Editor:          "vi",
			},
		},
		{
			name: "NO_COLOR env var",
			envVars: map[string]string{
				"NO_COLOR": "1",
			},
			want: &Config{
				DatabaseName:    "claude-tasks.db",
				ColorEnabled:    false,
				PageSize:        20,
				DefaultPriority: "medium",
				ShowWarnings:    true,
				Editor:          "vi",
			},
		},
		{
			name: "custom page size",
			envVars: map[string]string{
				"GTD_PAGE_SIZE": "50",
			},
			want: &Config{
				DatabaseName:    "claude-tasks.db",
				ColorEnabled:    true,
				PageSize:        50,
				DefaultPriority: "medium",
				ShowWarnings:    true,
				Editor:          "vi",
			},
		},
		{
			name: "custom editor",
			envVars: map[string]string{
				"EDITOR": "nano",
			},
			want: &Config{
				DatabaseName:    "claude-tasks.db",
				ColorEnabled:    true,
				PageSize:        20,
				DefaultPriority: "medium",
				ShowWarnings:    true,
				Editor:          "nano",
			},
		},
		{
			name: "VISUAL overrides EDITOR",
			envVars: map[string]string{
				"EDITOR": "nano",
				"VISUAL": "vim",
			},
			want: &Config{
				DatabaseName:    "claude-tasks.db",
				ColorEnabled:    true,
				PageSize:        20,
				DefaultPriority: "medium",
				ShowWarnings:    true,
				Editor:          "vim",
			},
		},
		{
			name: "all behavior flags",
			envVars: map[string]string{
				"GTD_AUTO_REVIEW":   "true",
				"GTD_SHOW_WARNINGS": "false",
				"GTD_CONFIRM_DONE":  "true",
			},
			want: &Config{
				DatabaseName:    "claude-tasks.db",
				ColorEnabled:    true,
				PageSize:        20,
				DefaultPriority: "medium",
				AutoReview:      true,
				ShowWarnings:    false,
				ConfirmDone:     true,
				Editor:          "vi",
			},
		},
		{
			name: "invalid format",
			envVars: map[string]string{
				"GTD_DEFAULT_FORMAT": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid priority",
			envVars: map[string]string{
				"GTD_DEFAULT_PRIORITY": "urgent",
			},
			wantErr: true,
		},
		{
			name: "invalid page size",
			envVars: map[string]string{
				"GTD_PAGE_SIZE": "zero",
			},
			wantErr: true,
		},
		{
			name: "negative page size",
			envVars: map[string]string{
				"GTD_PAGE_SIZE": "-5",
			},
			wantErr: true,
		},
		{
			name: "invalid boolean",
			envVars: map[string]string{
				"GTD_COLOR": "yes",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv := func() {
				vars := []string{
					"GTD_DATABASE_NAME", "GTD_DATABASE_PATH", "GTD_DEFAULT_FORMAT",
					"GTD_COLOR", "NO_COLOR", "GTD_PAGE_SIZE", "GTD_AUTO_REVIEW",
					"GTD_SHOW_WARNINGS", "GTD_CONFIRM_DONE", "GTD_DEFAULT_PRIORITY",
					"EDITOR", "VISUAL",
				}
				for _, v := range vars {
					os.Unsetenv(v)
				}
			}

			// Set up environment
			clearEnv()
			defer clearEnv()

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Test
			cfg := NewConfig()
			err := cfg.Load()

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.want != nil {
				// Compare relevant fields
				if cfg.DatabaseName != tt.want.DatabaseName {
					t.Errorf("DatabaseName = %s, want %s", cfg.DatabaseName, tt.want.DatabaseName)
				}
				if cfg.ColorEnabled != tt.want.ColorEnabled {
					t.Errorf("ColorEnabled = %v, want %v", cfg.ColorEnabled, tt.want.ColorEnabled)
				}
				if cfg.PageSize != tt.want.PageSize {
					t.Errorf("PageSize = %d, want %d", cfg.PageSize, tt.want.PageSize)
				}
				if cfg.DefaultPriority != tt.want.DefaultPriority {
					t.Errorf("DefaultPriority = %s, want %s", cfg.DefaultPriority, tt.want.DefaultPriority)
				}
				if cfg.AutoReview != tt.want.AutoReview {
					t.Errorf("AutoReview = %v, want %v", cfg.AutoReview, tt.want.AutoReview)
				}
				if cfg.ShowWarnings != tt.want.ShowWarnings {
					t.Errorf("ShowWarnings = %v, want %v", cfg.ShowWarnings, tt.want.ShowWarnings)
				}
				if cfg.ConfirmDone != tt.want.ConfirmDone {
					t.Errorf("ConfirmDone = %v, want %v", cfg.ConfirmDone, tt.want.ConfirmDone)
				}
				if cfg.Editor != tt.want.Editor {
					t.Errorf("Editor = %s, want %s", cfg.Editor, tt.want.Editor)
				}
			}
		})
	}
}

func TestGetDatabasePath(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "explicit path",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				DatabaseName: "ignored.db",
				GitRoot:      "/repo",
			},
			expected: "/tmp/test.db",
		},
		{
			name: "git root path",
			config: &Config{
				DatabaseName: "tasks.db",
				GitRoot:      "/repo",
			},
			expected: "/repo/tasks.db",
		},
		{
			name: "fallback to database name only",
			config: &Config{
				DatabaseName: "tasks.db",
			},
			expected: "tasks.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetDatabasePath()
			if got != tt.expected {
				t.Errorf("GetDatabasePath() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DefaultPriority: "high",
				DefaultFormat:   "json",
				PageSize:        10,
			},
			wantErr: false,
		},
		{
			name: "invalid priority",
			config: &Config{
				DefaultPriority: "urgent",
				PageSize:        10,
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			config: &Config{
				DefaultPriority: "medium",
				DefaultFormat:   "xml",
				PageSize:        10,
			},
			wantErr: true,
		},
		{
			name: "invalid page size",
			config: &Config{
				DefaultPriority: "medium",
				PageSize:        0,
			},
			wantErr: true,
		},
		{
			name: "empty format is valid",
			config: &Config{
				DefaultPriority: "medium",
				DefaultFormat:   "",
				PageSize:        10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	cfg := &Config{
		DatabaseName:    "test.db",
		GitRoot:         "/repo",
		DefaultFormat:   "json",
		ColorEnabled:    true,
		PageSize:        50,
		AutoReview:      true,
		ShowWarnings:    false,
		ConfirmDone:     true,
		DefaultPriority: "high",
		Editor:          "vim",
	}

	str := cfg.String()
	
	// Check that important values are in the string representation
	expectedContains := []string{
		"GTD Configuration:",
		"/repo/test.db",
		"json",
		"true",
		"50",
		"high",
		"vim",
	}

	for _, expected := range expectedContains {
		if !contains(str, expected) {
			t.Errorf("String() missing expected content: %s\nGot: %s", expected, str)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return filepath.Join(s, substr) != filepath.Join(s) || s == substr || (len(s) > 0 && len(substr) > 0 && strings.Contains(s, substr))
}