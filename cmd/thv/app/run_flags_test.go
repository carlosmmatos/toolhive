package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		envFileContent string
		expectedVars   []string
		expectError    bool
	}{
		{
			name:           "empty env file",
			envFileContent: "",
			expectedVars:   nil,
			expectError:    false,
		},
		{
			name: "basic env vars",
			envFileContent: `KEY1=value1
KEY2=value2`,
			expectedVars: []string{"KEY1=value1", "KEY2=value2"},
			expectError:  false,
		},
		{
			name: "env vars with quotes and spaces",
			envFileContent: `API_KEY="test key with spaces"
DATABASE_URL=postgres://user:pass@localhost/db
DEBUG=true`,
			expectedVars: []string{
				"API_KEY=test key with spaces",
				"DATABASE_URL=postgres://user:pass@localhost/db",
				"DEBUG=true",
			},
			expectError: false,
		},
		{
			name: "env vars with comments and empty lines",
			envFileContent: `# This is a comment
KEY1=value1

# Another comment
KEY2=value2
`,
			expectedVars: []string{"KEY1=value1", "KEY2=value2"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary file
			tempDir := t.TempDir()
			envFile := filepath.Join(tempDir, ".env")
			err := os.WriteFile(envFile, []byte(tt.envFileContent), 0o644)
			require.NoError(t, err)

			// Test loading env vars
			result, err := loadEnvFromFile(envFile)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedVars, result)
		})
	}
}

func TestLoadEnvFromFile_FileNotFound(t *testing.T) {
	t.Parallel()

	result, err := loadEnvFromFile("/nonexistent/file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "env file not found")
	assert.Nil(t, result)
}

func TestLoadEnvFromFile_EmptyPath(t *testing.T) {
	t.Parallel()

	result, err := loadEnvFromFile("")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetEnvVarKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "simple key-value",
			envVar:   "KEY=value",
			expected: "KEY",
		},
		{
			name:     "empty value",
			envVar:   "KEY=",
			expected: "KEY",
		},
		{
			name:     "value with equals",
			envVar:   "DATABASE_URL=postgres://user:pass@localhost/db",
			expected: "DATABASE_URL",
		},
		{
			name:     "no equals sign",
			envVar:   "INVALID",
			expected: "",
		},
		{
			name:     "empty string",
			envVar:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := getEnvVarKey(tt.envVar)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeEnvVars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		envFileContent string
		cmdLineVars    []string
		expectedVars   []string
		expectError    bool
	}{
		{
			name:           "no env file, only command line vars",
			envFileContent: "",
			cmdLineVars:    []string{"CMD_VAR=cmd_value"},
			expectedVars:   []string{"CMD_VAR=cmd_value"},
			expectError:    false,
		},
		{
			name:           "env file only, no command line vars",
			envFileContent: `FILE_VAR=file_value`,
			cmdLineVars:    nil,
			expectedVars:   []string{"FILE_VAR=file_value"},
			expectError:    false,
		},
		{
			name:           "both file and command line vars, no overlap",
			envFileContent: `FILE_VAR=file_value`,
			cmdLineVars:    []string{"CMD_VAR=cmd_value"},
			expectedVars:   []string{"FILE_VAR=file_value", "CMD_VAR=cmd_value"},
			expectError:    false,
		},
		{
			name: "command line overrides file var",
			envFileContent: `SHARED_VAR=file_value
FILE_VAR=file_value`,
			cmdLineVars:  []string{"SHARED_VAR=cmd_value", "CMD_VAR=cmd_value"},
			expectedVars: []string{"FILE_VAR=file_value", "SHARED_VAR=cmd_value", "CMD_VAR=cmd_value"},
			expectError:  false,
		},
		{
			name: "multiple overlapping vars",
			envFileContent: `VAR1=file1
VAR2=file2
VAR3=file3`,
			cmdLineVars:  []string{"VAR1=cmd1", "VAR3=cmd3", "VAR4=cmd4"},
			expectedVars: []string{"VAR2=file2", "VAR1=cmd1", "VAR3=cmd3", "VAR4=cmd4"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary file if content is provided
			var envFile string
			if tt.envFileContent != "" {
				tempDir := t.TempDir()
				envFile = filepath.Join(tempDir, ".env")
				err := os.WriteFile(envFile, []byte(tt.envFileContent), 0o644)
				require.NoError(t, err)
			}

			// Test merging env vars
			result, err := mergeEnvVars(envFile, tt.cmdLineVars)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedVars, result)
		})
	}
}

func TestMergeEnvVars_FileNotFound(t *testing.T) {
	t.Parallel()

	result, err := mergeEnvVars("/nonexistent/file", []string{"CMD_VAR=value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "env file not found")
	assert.Nil(t, result)
}
