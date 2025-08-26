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

	"github.com/id9051/got/internal/git"
	"github.com/id9051/got/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDirectoryPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "current directory",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "directory path cannot be empty",
		},
		{
			name:    "non-existent directory",
			path:    "/path/that/does/not/exist",
			wantErr: true,
			errMsg:  "directory does not exist",
		},
		{
			name:    "file instead of directory",
			path:    "go.mod",
			wantErr: true,
			errMsg:  "directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDirectoryPath(tt.path)
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

func TestValidateDirectoryPath_WithTempDir(t *testing.T) {
	// Test with temporary directory
	tempDir := t.TempDir()
	err := validateDirectoryPath(tempDir)
	assert.NoError(t, err)

	// Test with temporary file
	tempFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(tempFile, []byte("test"), 0644))

	err = validateDirectoryPath(tempFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is not a directory")
}

func TestIsGitRepository(t *testing.T) {
	tests := []struct {
		name     string
		setupDir func(t *testing.T) string
		expected bool
	}{
		{
			name: "directory with .git folder",
			setupDir: func(t *testing.T) string {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, git.DirName)
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
			result := isGitRepository(dir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldSkipPath(t *testing.T) {
	// Save and restore viper state
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for key, value := range originalConfig {
			viper.Set(key, value)
		}
	}()

	tests := []struct {
		name     string
		path     string
		skipList []string
		expected bool
	}{
		{
			name:     "path not in skip list",
			path:     "/some/random/path",
			skipList: []string{"node_modules", ".git"},
			expected: false,
		},
		{
			name:     "path segment matches skip item",
			path:     "/path/to/node_modules/package",
			skipList: []string{"node_modules"},
			expected: true,
		},
		{
			name:     "empty skip list",
			path:     "/any/path",
			skipList: []string{},
			expected: false,
		},
		{
			name:     "false positive prevention - substring in directory name",
			path:     "/project/vendor-tools/repo",
			skipList: []string{"vendor"},
			expected: false, // Should NOT be skipped because "vendor" is not a complete segment
		},
		{
			name:     "false positive prevention - substring in file name",
			path:     "/project/my-vendor-app/src",
			skipList: []string{"vendor"},
			expected: false, // Should NOT be skipped
		},
		{
			name:     "exact segment match",
			path:     "/project/vendor/package",
			skipList: []string{"vendor"},
			expected: true, // Should be skipped because "vendor" is an exact segment
		},
		{
			name:     "basename match",
			path:     "/project/tools/vendor",
			skipList: []string{"vendor"},
			expected: true, // Should be skipped because basename is "vendor"
		},
		{
			name:     "multiple segments with match",
			path:     "/home/user/.git/hooks/pre-commit",
			skipList: []string{".git"},
			expected: true,
		},
		{
			name:     "nested path with skip pattern",
			path:     "/project/src/node_modules/package/lib",
			skipList: []string{"node_modules"},
			expected: true,
		},
		{
			name:     "case sensitive matching",
			path:     "/project/Node_Modules/package",
			skipList: []string{"node_modules"},
			expected: false, // Case sensitive - should not match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and set up test configuration
			viper.Reset()
			viper.Set("skipList", tt.skipList)
			viper.Set("useDefaultSkips", false) // Disable defaults for predictable testing

			result := shouldSkipPath(tt.path)
			assert.Equal(t, tt.expected, result, "Skip check failed for path: %s with skipList: %v", tt.path, tt.skipList)
		})
	}
}

func TestMatchesSkipPattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "exact segment match",
			path:     "/project/vendor/package",
			pattern:  "vendor",
			expected: true,
		},
		{
			name:     "substring in segment - should not match",
			path:     "/project/vendor-tools/package",
			pattern:  "vendor",
			expected: false,
		},
		{
			name:     "basename match",
			path:     "/project/tools/vendor",
			pattern:  "vendor",
			expected: true,
		},
		{
			name:     "empty pattern",
			path:     "/any/path",
			pattern:  "",
			expected: false,
		},
		{
			name:     "pattern matches entire path",
			path:     "/vendor",
			pattern:  "/vendor",
			expected: true,
		},
		{
			name:     "pattern in middle of path",
			path:     "/home/user/node_modules/package",
			pattern:  "node_modules",
			expected: true,
		},
		{
			name:     "case sensitive - no match",
			path:     "/project/Vendor/package",
			pattern:  "vendor",
			expected: false,
		},
		{
			name:     "dotfile pattern",
			path:     "/project/.git/hooks",
			pattern:  ".git",
			expected: true,
		},
		{
			name:     "complex path with multiple separators",
			path:     "/home/user//project/./vendor/../vendor/lib",
			pattern:  "vendor",
			expected: true, // filepath.Clean should normalize this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSkipPattern(tt.path, tt.pattern)
			assert.Equal(t, tt.expected, result, "Pattern match failed for path: %s, pattern: %s", tt.path, tt.pattern)
		})
	}
}

func TestSkipPathLogicFix(t *testing.T) {
	// This test demonstrates the fix for the false positive issue
	// where paths containing skip terms as substrings were incorrectly skipped

	// Save and restore viper state
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for key, value := range originalConfig {
			viper.Set(key, value)
		}
	}()

	// Set up test configuration
	viper.Reset()
	viper.Set("skipList", []string{"vendor", "node_modules"})
	viper.Set("useDefaultSkips", false)

	testCases := []struct {
		path        string
		shouldSkip  bool
		description string
	}{
		{
			path:        "/project/vendor/package",
			shouldSkip:  true,
			description: "Exact segment match - should be skipped",
		},
		{
			path:        "/project/vendor-tools/repo",
			shouldSkip:  false,
			description: "False positive prevention - substring in directory name should NOT be skipped",
		},
		{
			path:        "/project/my-vendor-app/src",
			shouldSkip:  false,
			description: "False positive prevention - substring in directory name should NOT be skipped",
		},
		{
			path:        "/project/node_modules/package",
			shouldSkip:  true,
			description: "Exact segment match - should be skipped",
		},
		{
			path:        "/project/my-node_modules-backup/src",
			shouldSkip:  false,
			description: "False positive prevention - substring should NOT be skipped",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := shouldSkipPath(tc.path)
			assert.Equal(t, tc.shouldSkip, result,
				"Path: %s\nExpected: %v, Got: %v\nDescription: %s",
				tc.path, tc.shouldSkip, result, tc.description)
		})
	}
}

func TestLogFunctions(t *testing.T) {
	// Test that log functions don't panic
	t.Run("logSkipped", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logSkipped("/test/path")
		})
	})

	t.Run("logSuccess", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logSuccess("/test/path")
		})
	})

	t.Run("logError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logError("/test/path", assert.AnError)
		})
	})
}

func TestExecuteGitCommand(t *testing.T) {
	tests := []struct {
		name     string
		setupDir func(t *testing.T) string
		gitArgs  []string
		wantErr  bool
	}{
		{
			name: "non-git directory returns nil (skipped)",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			gitArgs: []string{"status"},
			wantErr: false, // Should return nil for non-git dirs in recursive mode
		},
		{
			name: "git directory with invalid command",
			setupDir: func(t *testing.T) string {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, git.DirName)
				require.NoError(t, os.Mkdir(gitDir, 0755))
				return tempDir
			},
			gitArgs: []string{"invalid-command"},
			wantErr: false, // Function returns nil even on git command failure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			err := executeGitCommand(context.Background(), dir, tt.gitArgs...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteGitCommandSingle(t *testing.T) {
	// Install mock git runner for all tests
	mockGit, cleanup := testutil.InstallMockGitRunner(t, func(runner git.CommandRunner) git.CommandRunner {
		return SetGitCommandRunner(runner)
	})
	defer cleanup()

	tests := []struct {
		name     string
		setupDir func(t *testing.T) string
		gitArgs  []string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "non-git directory returns error",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			gitArgs: []string{"status"},
			wantErr: true,
			errMsg:  "is not a git repository",
		},
		{
			name: "git directory processes command",
			setupDir: func(t *testing.T) string {
				tempDir := t.TempDir()
				gitDir := filepath.Join(tempDir, git.DirName)
				require.NoError(t, os.Mkdir(gitDir, 0755))

				// Configure mock to return success for status command
				mockGit.SetOutput("status", "mock status output")

				return tempDir
			},
			gitArgs: []string{"status"},
			wantErr: false, // Function returns nil with mocked git command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			err := executeGitCommandSingle(context.Background(), dir, tt.gitArgs...)
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

func TestWalkDirectories(t *testing.T) {
	// Create a complex directory structure for testing
	tempDir := t.TempDir()

	// Create subdirectories
	subDir1 := filepath.Join(tempDir, "repo1")
	subDir2 := filepath.Join(tempDir, "repo2")
	nonGitDir := filepath.Join(tempDir, "notrepo")

	require.NoError(t, os.MkdirAll(subDir1, 0755))
	require.NoError(t, os.MkdirAll(subDir2, 0755))
	require.NoError(t, os.MkdirAll(nonGitDir, 0755))

	// Make repo1 and repo2 git repositories
	require.NoError(t, os.Mkdir(filepath.Join(subDir1, git.DirName), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(subDir2, git.DirName), 0755))

	// Track which directories the operation was called on
	var calledPaths []string
	testOperation := func(ctx context.Context, path string) error {
		calledPaths = append(calledPaths, path)
		return nil
	}

	err := walkDirectories(context.Background(), tempDir, testOperation)
	assert.NoError(t, err)

	// Should have been called on the root and all subdirectories
	assert.Contains(t, calledPaths, tempDir)
	assert.Contains(t, calledPaths, subDir1)
	assert.Contains(t, calledPaths, subDir2)
	assert.Contains(t, calledPaths, nonGitDir)
}
