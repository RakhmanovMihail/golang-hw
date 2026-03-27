// Package logger provides logging functionality for the application.
package logger

// Level represents the log level.
type Level string

const (
	// LevelDebug is the debug log level.
	LevelDebug Level = "debug"
	// LevelInfo is the info log level.
	LevelInfo Level = "info"
	// LevelWarn is the warn log level.
	LevelWarn Level = "warn"
	// LevelError is the error log level.
	LevelError Level = "error"
)
