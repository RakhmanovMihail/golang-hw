package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	// Create the command
	command := exec.Command(cmd[0], cmd[1:]...)

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
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Stderr != nil {
				fmt.Fprintln(os.Stderr, string(exitErr.Stderr))
			}
			return exitErr.ExitCode()
		}
		return 1
	}

	return 0
}
