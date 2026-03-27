package logger_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/RakhmanovMihail/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger_New(t *testing.T) {
	tests := []struct {
		name  string
		input logger.Level
		want  logger.Level
	}{
		{"debug", logger.LevelDebug, logger.LevelDebug},
		{"info", logger.LevelInfo, logger.LevelInfo},
		{"warn", logger.LevelWarn, logger.LevelWarn},
		{"error", logger.LevelError, logger.LevelError},
		{"invalid", logger.Level("invalid"), logger.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.New(tt.input)
			assert.Equal(t, tt.want, l.Level)
		})
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		setLevel  logger.Level
		testLevel logger.Level
		shouldLog bool
	}{
		{"debug→debug", logger.LevelDebug, logger.LevelDebug, true},
		{"debug→info", logger.LevelDebug, logger.LevelInfo, true},
		{"info→debug", logger.LevelInfo, logger.LevelDebug, false},
		{"info→info", logger.LevelInfo, logger.LevelInfo, true},
		{"warn→warn", logger.LevelWarn, logger.LevelWarn, true},
		{"error→error", logger.LevelError, logger.LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.New(tt.setLevel)
			output := captureOutput(t, func() {
				switch tt.testLevel {
				case logger.LevelDebug:
					l.Debug("test")
				case logger.LevelInfo:
					l.Info("test")
				case logger.LevelWarn:
					l.Warn("test")
				case logger.LevelError:
					l.Error("test")
				}
			})

			expected := "[" + strings.ToUpper(string(tt.testLevel)) + "] test\n"
			if tt.shouldLog {
				assert.Contains(t, output, expected)
			} else {
				assert.NotContains(t, output, expected)
			}
		})
	}
}

func captureOutput(t *testing.T, f func()) string {
	r, w, err := os.Pipe()
	assert.NoError(t, err)

	oldStdout := os.Stdout
	os.Stdout = w

	done := make(chan bool, 1)
	buf := new(bytes.Buffer)

	go func() {
		_, err = buf.ReadFrom(r)
		assert.NoError(t, err)
		done <- true
	}()

	f() // Вызываем logger

	w.Close()
	os.Stdout = oldStdout

	<-done // Ждём завершения чтения
	r.Close()

	return buf.String()
}
