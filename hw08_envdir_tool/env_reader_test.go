package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name        string
	setup       func(t *testing.T) string
	expected    Environment
	expectError bool
}

func createTestDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "envtest")
	require.NoError(t, err, "Failed to create temp dir")
	t.Cleanup(func() { os.RemoveAll(dir) })

	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
		require.NoErrorf(t, err, "Failed to create test file %s", name)
	}

	return dir
}

func assertEnvironmentsEqual(t *testing.T, expected, actual Environment) {
	t.Helper()

	assert.Equal(t, len(expected), len(actual), "Number of environment variables does not match")

	for key, expectedVal := range expected {
		actualVal, exists := actual[key]
		if assert.Truef(t, exists, "Expected key %s not found in result", key) {
			assert.Equalf(t, expectedVal.Value, actualVal.Value, "Value mismatch for key %s", key)
			assert.Equalf(t, expectedVal.NeedRemove, actualVal.NeedRemove,
				"NeedRemove mismatch for key %s", key)
		}
	}
}

func TestReadDir(t *testing.T) {
	testCases := []testCase{
		{
			name: "normal case",
			setup: func(t *testing.T) string {
				t.Helper()
				return createTestDir(t, map[string]string{
					"VAR1":  "value1",
					"VAR2":  "value2\nwith newline",
					"VAR3":  "value3\x00with null",
					"EMPTY": "",
				})
			},
			expected: Environment{
				"VAR1":  {Value: "value1", NeedRemove: false},
				"VAR2":  {Value: "value2", NeedRemove: false},
				"VAR3":  {Value: "value3\nwith null", NeedRemove: false},
				"EMPTY": {Value: "", NeedRemove: true},
			},
			expectError: false,
		},
		{
			name: "skip invalid filenames",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := createTestDir(t, map[string]string{
					"INVALID=VAR": "value",
					"VALID_VAR":   "valid",
				})
				return dir
			},
			expected:    Environment{},
			expectError: false,
		},
		{
			name: "skip directories",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := createTestDir(t, map[string]string{"VALID": "valid"})

				subDir := filepath.Join(dir, "subdir")
				err := os.Mkdir(subDir, 0o755)
				require.NoError(t, err, "Failed to create subdirectory")

				err = os.WriteFile(filepath.Join(subDir, "SHOULD_NOT_BE_READ"), []byte("test"), 0o644)
				require.NoError(t, err, "Failed to write test file")

				return dir
			},
			expected: Environment{
				"VALID": {Value: "valid", NeedRemove: false},
			},
			expectError: false,
		},
		{
			name: "non-existent directory",
			setup: func(_ *testing.T) string {
				return "/non/existent/directory/that/should/not/exist"
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testDir := tc.setup(t)
			result, err := ReadDir(testDir)

			if tc.expectError {
				assert.Error(t, err, "Expected an error but got none")
				return
			}

			require.NoError(t, err, "Unexpected error")
			assertEnvironmentsEqual(t, tc.expected, result)
		})
	}
}
