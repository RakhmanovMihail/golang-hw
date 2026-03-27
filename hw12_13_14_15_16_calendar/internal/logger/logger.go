package logger

import "fmt"

type Logger struct {
	Level Level
}

type ILogger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

func New(levelStr Level) *Logger {
	level := Level(levelStr)

	switch level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
	default:
		level = LevelInfo
	}

	return &Logger{Level: level}
}

func (l Logger) Error(msg string) {
	if l.shouldLog(LevelError) {
		fmt.Printf("[ERROR] %s\n", msg)
	}
}

func (l Logger) Warn(msg string) {
	if l.shouldLog(LevelWarn) {
		fmt.Printf("[WARN] %s\n", msg)
	}
}

func (l Logger) Info(msg string) {
	if l.shouldLog(LevelInfo) {
		fmt.Printf("[INFO] %s\n", msg)
	}
}

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
