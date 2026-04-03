package logger

import "fmt"

// Logger implements a simple logger.
type Logger struct {
	Level Level
}

// ILogger is the interface for logger.
type ILogger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// New creates a new Logger instance.
func New(levelStr Level) *Logger {
	level := levelStr

	switch level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		level = LevelInfo
	}

	return &Logger{Level: level}
}

// Error logs an error message.
func (l Logger) Error(msg string) {
	if l.shouldLog(LevelError) {
		fmt.Printf("[ERROR] %s\n", msg)
	}
}

// Warn logs a warning message.
func (l Logger) Warn(msg string) {
	if l.shouldLog(LevelWarn) {
		fmt.Printf("[WARN] %s\n", msg)
	}
}

// Info logs an info message.
func (l Logger) Info(msg string) {
	if l.shouldLog(LevelInfo) {
		fmt.Printf("[INFO] %s\n", msg)
	}
}

// Debug logs a debug message.
func (l Logger) Debug(msg string) {
	if l.shouldLog(LevelDebug) {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

func (l Logger) shouldLog(required Level) bool {
	levels := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	return levels[l.Level] <= levels[required]
}
