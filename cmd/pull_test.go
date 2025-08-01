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

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/id9051/got/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullCmd(t *testing.T) {
	// Test that pull command is properly configured
	assert.NotNil(t, pullCmd)
	assert.Equal(t, "pull directory", pullCmd.Use)
	assert.Equal(t, "Pull changes from remote repositories", pullCmd.Short)
	assert.Contains(t, pullCmd.Long, "Pull changes from remote repositories")
	assert.Contains(t, pullCmd.Long, "recursive flag")
	assert.Contains(t, pullCmd.Long, "Examples:")
}

func TestPullCmd_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
			errMsg:  "directory argument is required",
		},
		{
			name:    "single valid directory",
			args:    []string{"."},
			wantErr: false,
		},
		{
			name:    "invalid directory",
			args:    []string{"/does/not/exist"},
			wantErr: true,
			errMsg:  "directory does not exist",
		},
		{
			name:    "file instead of directory",
			args:    []string{"go.mod"},
			wantErr: true,
			errMsg:  "directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the command to avoid state issues
			cmd := &cobra.Command{
				Use:  pullCmd.Use,
				RunE: pullCmd.RunE,
			}
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				// Note: This might still error due to git command execution
				// but it shouldn't error due to argument validation
				if err != nil {
					// If it errors, it should not be due to validation
					assert.NotContains(t, err.Error(), "directory argument is required")
					assert.NotContains(t, err.Error(), "directory does not exist")
				}
			}
		})
	}
}

func TestPullCmd_FlagHandling(t *testing.T) {
	t.Run("recursive flag", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test recursive flag parsing
		pullCmd.SetArgs([]string{"-r", tempDir})

		// We can't easily test the full execution, but we can test that
		// the command accepts the flag without argument validation errors
		err := pullCmd.ParseFlags([]string{"-r", tempDir})
		assert.NoError(t, err)

		// Check that the flag was parsed correctly
		recursive, err := pullCmd.Flags().GetBool("recursive")
		assert.NoError(t, err)
		assert.True(t, recursive)

		// Reset for next test
		pullCmd.SetArgs(nil)
	})

	t.Run("short recursive flag", func(t *testing.T) {
		tempDir := t.TempDir()

		pullCmd.SetArgs([]string{"-r", tempDir})
		err := pullCmd.ParseFlags([]string{"-r", tempDir})
		assert.NoError(t, err)

		recursive, err := pullCmd.Flags().GetBool("recursive")
		assert.NoError(t, err)
		assert.True(t, recursive)

		pullCmd.SetArgs(nil)
	})
}

func TestPullSingle(t *testing.T) {
	// Install mock git runner for all tests
	mockGit, cleanup := testutil.InstallMockGitRunner(t, func(runner testutil.GitCommandRunnerInterface) testutil.GitCommandRunnerInterface {
		return SetGitCommandRunner(runner)
	})
	defer cleanup()

	tests := []struct {
		name     string
		setupDir func(t *testing.T) string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "non-git directory",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			errMsg:  "is not a git repository",
		},
		{
			name: "git directory",
			setupDir: func(t *testing.T) string {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, GitDirName)
				require.NoError(t, os.Mkdir(gitDir, 0755))
				
				// Configure mock to return success for pull command
				mockGit.SetOutput("pull", "mock pull output")
				
				return tempDir
			},
			wantErr: false, // Function returns nil with mocked git command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			err := pullSingle(context.Background(), dir)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}


func TestPullCmd_Examples(t *testing.T) {
	// Test that examples are properly formatted and present
	examples := pullCmd.Long
	assert.Contains(t, examples, "got pull .")
	assert.Contains(t, examples, "got pull /path/to/repo")
	assert.Contains(t, examples, "got pull -r /path/to/projects")
}

func TestPullCmd_Integration(t *testing.T) {
	// Install mock git runner for integration tests
	mockGit, cleanup := testutil.InstallMockGitRunner(t, func(runner testutil.GitCommandRunnerInterface) testutil.GitCommandRunnerInterface {
		return SetGitCommandRunner(runner)
	})
	defer cleanup()

	// Configure mock to return success for pull command
	mockGit.SetOutput("pull", "mock pull output")

	// Create a complex directory structure for integration testing
	tempDir := t.TempDir()

	// Create multiple subdirectories, some with git repos
	repo1 := filepath.Join(tempDir, "repo1")
	repo2 := filepath.Join(tempDir, "repo2")
	nonRepo := filepath.Join(tempDir, "nonrepo")

	require.NoError(t, os.MkdirAll(repo1, 0755))
	require.NoError(t, os.MkdirAll(repo2, 0755))
	require.NoError(t, os.MkdirAll(nonRepo, 0755))

	// Make repo1 and repo2 git repositories
	require.NoError(t, os.Mkdir(filepath.Join(repo1, GitDirName), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(repo2, GitDirName), 0755))

	t.Run("git repository detection", func(t *testing.T) {
		// Test git repository detection directly
		assert.True(t, isGitRepository(repo1))
		assert.True(t, isGitRepository(repo2))
		assert.False(t, isGitRepository(nonRepo))
	})

	t.Run("pull functions work correctly", func(t *testing.T) {
		// Test pullSingle function directly
		err := pullSingle(context.Background(), repo1)
		assert.NoError(t, err) // Should not error with mocked git command

		// Test with non-git repo
		err = pullSingle(context.Background(), nonRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a git repository")
	})
}
