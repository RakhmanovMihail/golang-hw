package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	env := make(Environment)

	for _, entry := range dirEntries {
		if strings.Contains(entry.Name(), "=") {
			continue // пропускаем файлы с '=' в имени
		}

		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
		}

		scanner := bufio.NewScanner(file)
		var value string

		if scanner.Scan() {
			value = strings.TrimRight(scanner.Text(), " \t")
			value = strings.ReplaceAll(value, "\x00", "\n")
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Закрываем файл сразу после чтения
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("failed to close file %s: %w", filePath, err)
		}

		if value == "" {
			env[entry.Name()] = EnvValue{NeedRemove: true}
		} else {
			env[entry.Name()] = EnvValue{
				Value:      value,
				NeedRemove: false,
			}
		}
	}

	return env, nil
}
