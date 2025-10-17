package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name        string
		offset      int64
		limit       int64
		expectError bool
	}{
		{"offset 0, limit 0", 0, 0, false},
		{"offset 0, limit 10", 0, 10, false},
		{"offset 0, limit 1000", 0, 1000, false},
		{"offset 0, limit 10000", 0, 10000, false},
		{"offset 100, limit 1000", 100, 1000, false},
		{"offset 6000, limit 1000", 6000, 1000, false},
		{"offset beyond file size", 100000, 1000, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outputFile := "test_out_" + tc.name + ".txt"
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {

				}
			}(outputFile)

			err := Copy("testdata/input.txt", outputFile, tc.offset, tc.limit)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			_, err = os.Stat(outputFile)
			assert.NoError(t, err, "output file should exist")

		})
	}
}

func TestCopy_Validation(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		offset      int64
		limit       int64
		expectError bool
	}{
		{"non-existent source", "nonexistent.txt", 0, 10, true},
		{"directory as source", "testdata", 0, 10, true},
		{"negative offset", "testdata/input.txt", -1, 10, true},
		{"negative limit", "testdata/input.txt", 0, -1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outputFile := "test_validation_" + tc.name + ".txt"
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					t.Logf("failed to remove file: %v", err)
				}
			}(outputFile)

			err := Copy(tc.src, outputFile, tc.offset, tc.limit)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
