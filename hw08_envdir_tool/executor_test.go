package main

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

const windowsOS = "windows"

func TestRunCmd(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		env      Environment
		expected int
	}{
		{
			name: "simple command with success",
			cmd:  createEchoCommand("test"),
			env: Environment{
				"TEST_VAR": {Value: "test_value", NeedRemove: false},
			},
			expected: 0,
		},
		{
			name: "command with environment variables",
			cmd:  createPrintEnvCommand("TEST_VAR"),
			env: Environment{
				"TEST_VAR": {Value: "test_value", NeedRemove: false},
			},
			expected: 0,
		},
		{
			name: "remove existing environment variable",
			cmd:  createPrintEnvCommand("SHOULD_NOT_EXIST"),
			env: Environment{
				"SHOULD_NOT_EXIST": {Value: "", NeedRemove: true},
			},
			expected: 1, // Printenv fails with code 1 when variable doesn't exist
		},
		{
			name:     "command with non-zero exit code",
			cmd:      createExitCodeCommand(42),
			env:      Environment{},
			expected: 42,
		},
		{
			name:     "empty command",
			cmd:      []string{},
			env:      Environment{},
			expected: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			if val, exists := os.LookupEnv("SHOULD_NOT_EXIST"); exists {
				t.Cleanup(func() { os.Setenv("SHOULD_NOT_EXIST", val) })
			} else {
				t.Cleanup(func() { os.Unsetenv("SHOULD_NOT_EXIST") })
			}
			os.Setenv("SHOULD_NOT_EXIST", "should be removed")

			// Run the command
			got := RunCmd(tc.cmd, tc.env)

			// Verify the result
			require.Equal(t, tc.expected, got, "Unexpected exit code")
		})
	}
}

// createEchoCommand creates a command that prints the given message to stdout.
func createEchoCommand(message string) []string {
	if runtime.GOOS == windowsOS {
		return []string{"cmd", "/C", "echo", message}
	}
	return []string{"echo", message}
}

// createPrintEnvCommand creates a command that prints the value of the given environment variable.
func createPrintEnvCommand(varName string) []string {
	if runtime.GOOS == windowsOS {
		return []string{"cmd", "/C", "if not defined " + varName + " exit 1"}
	}
	return []string{"sh", "-c", "printenv " + varName + " || exit 1"}
}

// createExitCodeCommand creates a command that exits with the given status code.
func createExitCodeCommand(code int) []string {
	if runtime.GOOS == windowsOS {
		return []string{"cmd", "/C", "exit", fmt.Sprint(code)}
	}
	return []string{"sh", "-c", "exit " + fmt.Sprint(code)}
}
