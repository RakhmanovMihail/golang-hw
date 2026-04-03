package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	release   = "UNKNOWN" //nolint:unused // Заполняется при сборке через -ldflags
	buildDate = "UNKNOWN" //nolint:unused // Заполняется при сборке через -ldflags
	gitHash   = "UNKNOWN" //nolint:unused // Заполняется при сборке через -ldflags
)

func printVersion() { //nolint:unused // Вызывается через команду version
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Release   string
		BuildDate string
		GitHash   string
	}{
		Release:   release,
		BuildDate: buildDate,
		GitHash:   gitHash,
	}); err != nil {
		fmt.Printf("error while decode version info: %v\n", err)
	}
}
