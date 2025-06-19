package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:     "no args shows help",
			args:     []string{},
			wantErr:  false,
			contains: []string{"gtd", "task management tool", "Usage"},
		},
		{
			name:     "help flag",
			args:     []string{"--help"},
			wantErr:  false,
			contains: []string{"claude-gtd", "task management tool", "Available Commands"},
		},
		{
			name:     "invalid command",
			args:     []string{"invalid-command"},
			wantErr:  true,
			contains: []string{"unknown command"},
		},
		{
			name:     "version flag",
			args:     []string{"--version"},
			wantErr:  false,
			contains: []string{"claude-gtd version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			rootCmd := NewRootCommand()
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := stdout.String() + stderr.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("Output does not contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	rootCmd := NewRootCommand()

	// Expected commands
	expectedCommands := []string{
		"add-bug",
		"add-feature",
		"add-regression",
		"add-subtask",
		"in-progress",
		"done",
		"cancel",
		"block",
		"unblock",
		"list",
		"list-done",
		"list-cancelled",
		"show",
		"search",
		"summary",
		"export",
	}

	// Get all subcommands
	commands := rootCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	// Check all expected commands exist
	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Expected command %q not found", expected)
		}
	}
}
