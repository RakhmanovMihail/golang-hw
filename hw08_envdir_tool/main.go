package main

import (
	"fmt"
	"os"
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

	// Запускаем команду и выводим её результат
	os.Exit(RunCmd(command, env))
}
