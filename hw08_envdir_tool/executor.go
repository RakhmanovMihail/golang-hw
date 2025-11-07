package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	// Validate the command path
	if cmd[0] == "" {
		return 1
	}

	// Get the absolute path of the command to prevent path traversal
	path, err := exec.LookPath(cmd[0])
	if err != nil {
		return 1
	}

	// Additional security check: ensure the path is clean and doesn't contain any path traversal
	if path != filepath.Clean(path) {
		return 1
	}

	// Create the command with context
	ctx := context.Background()
	var command *exec.Cmd

	switch len(cmd) {
	case 1:
		command = exec.CommandContext(ctx, path)
	default:
		// Use the resolved path but keep the original command name for display
		args := make([]string, len(cmd))
		copy(args, cmd[1:])
		command = exec.CommandContext(ctx, path, args...)
	}

	// Set up the command's standard I/O
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Get current environment
	command.Env = os.Environ()

	// Apply environment variables from env
	for key, value := range env {
		if value.NeedRemove {
			// Remove the variable if it exists
			for i, envVar := range command.Env {
				if strings.HasPrefix(envVar, key+"=") {
					command.Env = append(command.Env[:i], command.Env[i+1:]...)
					break
				}
			}
		} else {
			// Set or update the variable
			found := false
			for i, envVar := range command.Env {
				if strings.HasPrefix(envVar, key+"=") {
					command.Env[i] = key + "=" + value.Value
					found = true
					break
				}
			}
			if !found {
				command.Env = append(command.Env, key+"="+value.Value)
			}
		}
	}

	// Start the command and wait for it to finish
	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if len(exitErr.Stderr) > 0 {
				fmt.Fprintln(os.Stderr, string(exitErr.Stderr))
			}
			return exitErr.ExitCode()
		}
		// For non-ExitError cases, print the error and return error code 1
		fmt.Fprintln(os.Stderr, "Error executing command:", err)
		return 1
	}

	return 0
}
