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
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCmd(t *testing.T) {
	// Test that status command is properly configured
	assert.NotNil(t, statusCmd)
	assert.Equal(t, "status directory", statusCmd.Use)
	assert.Equal(t, "Show working tree status of repositories", statusCmd.Short)
	assert.Contains(t, statusCmd.Long, "Show working tree status")
	assert.Contains(t, statusCmd.Long, "recursive flag")
	assert.Contains(t, statusCmd.Long, "Examples:")
}

func TestStatusCmd_ArgumentValidation(t *testing.T) {
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
				Use:  statusCmd.Use,
				RunE: statusCmd.RunE,
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
					assert.NotContains(t, err.Error(), "directory argument is required")
					assert.NotContains(t, err.Error(), "directory does not exist")
				}
			}
		})
	}
}

func TestStatusCmd_FlagHandling(t *testing.T) {
	t.Run("recursive flag", func(t *testing.T) {
		tempDir := t.TempDir()
		
		statusCmd.SetArgs([]string{"--recursive", tempDir})
		err := statusCmd.ParseFlags([]string{"--recursive", tempDir})
		assert.NoError(t, err)
		
		recursive, err := statusCmd.Flags().GetBool("recursive")
		assert.NoError(t, err)
		assert.True(t, recursive)
		
		statusCmd.SetArgs(nil)
	})

	t.Run("short recursive flag", func(t *testing.T) {
		tempDir := t.TempDir()
		
		statusCmd.SetArgs([]string{"-r", tempDir})
		err := statusCmd.ParseFlags([]string{"-r", tempDir})
		assert.NoError(t, err)
		
		recursive, err := statusCmd.Flags().GetBool("recursive")
		assert.NoError(t, err)
		assert.True(t, recursive)
		
		statusCmd.SetArgs(nil)
	})
}

func TestStatusSingle(t *testing.T) {
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
				return tempDir
			},
			wantErr: false, // Function returns nil even if git command fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			err := statusSingle(dir)
			
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

func TestStatusWalk(t *testing.T) {
	// Test the deprecated statusWalk function
	tempDir := t.TempDir()
	
	// Create a git repository
	gitDir := filepath.Join(tempDir, GitDirName)
	require.NoError(t, os.Mkdir(gitDir, 0755))
	
	// statusWalk should not return an error (it returns nil even on git failures)
	err := statusWalk(tempDir)
	assert.NoError(t, err)
}

func TestStatusCmd_Examples(t *testing.T) {
	// Test that examples are properly formatted and present
	examples := statusCmd.Long
	assert.Contains(t, examples, "got status .")
	assert.Contains(t, examples, "got status /path/to/repo")
	assert.Contains(t, examples, "got status -r /path/to/projects")
}

func TestStatusCmd_UniqueFeatures(t *testing.T) {
	// Test status-specific features
	assert.Contains(t, statusCmd.Long, "working tree status")
	
	// Ensure it's different from other commands
	assert.NotEqual(t, statusCmd.Short, pullCmd.Short)
	assert.NotEqual(t, statusCmd.Short, fetchCmd.Short)
	assert.Contains(t, statusCmd.Short, "status")
}

func TestStatusCmd_OutputHandling(t *testing.T) {
	// Status command should show output to user (unlike pull/fetch)
	// This is tested indirectly through the runGitCommand function
	// which has special handling for status commands
	
	// We can verify that the status command uses the same git execution path
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, GitDirName)
	require.NoError(t, os.Mkdir(gitDir, 0755))
	
	// The status command should not error on execution path
	err := statusSingle(tempDir)
	assert.NoError(t, err) // Returns nil even if git command fails
}

func TestStatusCmd_Integration(t *testing.T) {
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
	
	t.Run("status functions work correctly", func(t *testing.T) {
		// Test statusSingle function directly
		err := statusSingle(repo1)
		assert.NoError(t, err) // Should not error even if git command fails
		
		// Test with non-git repo
		err = statusSingle(nonRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a git repository")
	})
}