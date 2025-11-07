package main

import (
	"context"
	"errors"
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

	// For the first command, try to find it in the system path
	var path string
	var err error

	// Special handling for commands with paths
	if filepath.IsAbs(cmd[0]) || strings.Contains(cmd[0], string(filepath.Separator)) {
		// If it's an absolute path or contains path separators, use it as is
		path = cmd[0]
	} else {
		// Otherwise, try to find it in the system path
		path, err = exec.LookPath(cmd[0])
		if err != nil {
			return 1
		}
	}

	// Clean the path to prevent path traversal
	path = filepath.Clean(path)

	// Create the command with context
	ctx := context.Background()
	var command *exec.Cmd

	// Prepare the command with arguments
	if len(cmd) > 1 {
		command = exec.CommandContext(ctx, path, cmd[1:]...)
	} else {
		command = exec.CommandContext(ctx, path)
	}

	// Set up the command's standard I/O
	command.Stdin = os.Stdin
	// Capture both stdout and stderr
	var outputBuf strings.Builder
	command.Stdout = &outputBuf
	command.Stderr = &outputBuf

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

	// Run the command and capture its output
	err = command.Run()
	// Handle command execution errors
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Return the captured output and the exit code
			return exitErr.ExitCode()
		}
		// For non-ExitError cases, include the error in the output
		return 1
	}

	return 0
}
