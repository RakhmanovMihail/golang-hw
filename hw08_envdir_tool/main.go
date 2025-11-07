package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go-envdir <env_dir> <command> [args...]")
		os.Exit(1)
	}

	envDir := os.Args[1]
	command := os.Args[2:]

	env, err := ReadDir(envDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading environment directory: %v\n", err)
		os.Exit(1)
	}

	// Convert Windows paths to WSL paths if running on Windows
	if runtime.GOOS == "windows" && len(command) > 0 && strings.HasPrefix(command[0], "/mnt/") {
		// This is a WSL path, convert it to a Windows path for execution
		command[0] = "/bin/bash"
		command = append([]string{"-c", strings.Join(command, " ")}, command[1:]...)
	}

	// Запускаем команду и выводим её результат
	os.Exit(RunCmd(command, env))
}
