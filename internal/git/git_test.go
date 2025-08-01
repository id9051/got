// Copyright © 2025 Jeff Durham <jeffrey.durham@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRepository(t *testing.T) {
	tests := []struct {
		name     string
		setupDir func(t *testing.T) string
		expected bool
	}{
		{
			name: "directory with .git folder",
			setupDir: func(t *testing.T) string {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, DirName)
				require.NoError(t, os.Mkdir(gitDir, 0755))
				return tempDir
			},
			expected: true,
		},
		{
			name: "directory without .git folder",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			expected: false,
		},
		{
			name: "non-existent directory",
			setupDir: func(t *testing.T) string {
				return "/path/that/does/not/exist"
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			result := IsRepository(dir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetCommandRunner(t *testing.T) {
	// Create a mock runner
	mockRunner := &MockCommandRunner{}
	
	// Set the mock runner and get the original
	originalRunner := SetCommandRunner(mockRunner)
	
	// Verify the mock is now set
	assert.Equal(t, mockRunner, runner)
	
	// Restore the original runner
	SetCommandRunner(originalRunner)
	
	// Verify the original is restored
	assert.Equal(t, originalRunner, runner)
}

// MockCommandRunner is a simple mock for testing
type MockCommandRunner struct {
	Commands [][]string
}

func (m *MockCommandRunner) RunGitCommand(ctx context.Context, path string, args []string) ([]byte, error) {
	m.Commands = append(m.Commands, args)
	return []byte("mock output"), nil
}

func TestRunCommand(t *testing.T) {
	// Create and install mock runner
	mockRunner := &MockCommandRunner{}
	originalRunner := SetCommandRunner(mockRunner)
	defer SetCommandRunner(originalRunner)
	
	// Create a temporary git repository
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, DirName)
	require.NoError(t, os.Mkdir(gitDir, 0755))
	
	// Run a command
	output, err := RunCommand(context.Background(), tempDir, []string{"status"})
	
	// Verify the command was executed
	assert.NoError(t, err)
	assert.Equal(t, []byte("mock output"), output)
	
	// Verify the correct arguments were passed
	require.Len(t, mockRunner.Commands, 1)
	args := mockRunner.Commands[0]
	assert.Equal(t, "--work-tree="+tempDir, args[0])
	assert.Equal(t, "--git-dir="+filepath.Join(tempDir, DirName), args[1])
	assert.Equal(t, "status", args[2])
}