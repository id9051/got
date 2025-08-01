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

func TestRecursiveOperations_Integration(t *testing.T) {
	// Save and restore viper state
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for key, value := range originalConfig {
			viper.Set(key, value)
		}
	}()

	// Create test structure
	rootPath, repos := testutil.CreateAdvancedTestStructure(t)

	// Configure skip list to match our test structure
	skipList := []string{"node_modules", ".git", "vendor", "skip-me"}
	viper.Reset()
	viper.Set("skipList", skipList)

	t.Run("walkDirectories processes all directories", func(t *testing.T) {
		var processedPaths []string
		testOperation := func(ctx context.Context, path string) error {
			processedPaths = append(processedPaths, path)
			return nil
		}

		ctx := context.Background()
		err := walkDirectories(ctx, rootPath, testOperation)
		assert.NoError(t, err)

		// Count expected git repositories that should not be skipped
		expectedRepos := 0
		for _, repo := range repos {
			if repo.IsGit && !repo.Skipped {
				expectedRepos++
			}
		}

		// Should have processed only git repositories that aren't skipped
		assert.GreaterOrEqual(t, len(processedPaths), 1)
		assert.LessOrEqual(t, len(processedPaths), expectedRepos)
	})

	t.Run("git repositories are identified correctly", func(t *testing.T) {
		gitRepos := testutil.FilterGitRepos(repos)
		nonGitRepos := make([]testutil.GitRepoInfo, 0)

		for _, repo := range repos {
			if !repo.IsGit {
				nonGitRepos = append(nonGitRepos, repo)
			}
		}

		// Test git repository detection
		for _, repo := range gitRepos {
			assert.True(t, isGitRepository(repo.Path), "Should detect git repo: %s", repo.Path)
		}

		// Test non-git directory detection
		for _, repo := range nonGitRepos {
			assert.False(t, isGitRepository(repo.Path), "Should not detect as git repo: %s", repo.Path)
		}
	})

	t.Run("skip list works correctly", func(t *testing.T) {
		testCases := []struct {
			path     string
			expected bool
		}{
			{filepath.Join(rootPath, "repo1"), false},
			{filepath.Join(rootPath, "node_modules/package"), true},
			{filepath.Join(rootPath, "vendor/github.com"), true},
			{filepath.Join(rootPath, "skip-me"), true},
			{filepath.Join(rootPath, "normal/repo"), false},
		}

		for _, tc := range testCases {
			result := shouldSkipPath(tc.path)
			assert.Equal(t, tc.expected, result, "Skip check failed for path: %s", tc.path)
		}
	})
}

func TestFullCommandExecution_Integration(t *testing.T) {
	// Install mock git runner for integration tests
	mockGit, cleanup := testutil.InstallMockGitRunner(t, func(runner git.CommandRunner) git.CommandRunner {
		return SetGitCommandRunner(runner)
	})
	defer cleanup()

	// Configure mock to return success for all git commands
	mockGit.SetOutput("pull", "mock pull output")
	mockGit.SetOutput("fetch", "mock fetch output")
	mockGit.SetOutput("status", "mock status output")

	// Create a simpler test structure for command execution
	_, dirs := testutil.CreateTestDirStructure(t)

	t.Run("single functions work correctly", func(t *testing.T) {
		ctx := context.Background()
		repo1Path := dirs["repo1"]
		testutil.AssertIsGitRepo(t, repo1Path)

		// Test single functions directly
		err := pullSingle(ctx, repo1Path)
		assert.NoError(t, err) // Should not error with mocked git commands

		err = fetchSingle(ctx, repo1Path)
		assert.NoError(t, err) // Should not error with mocked git commands

		err = statusSingle(ctx, repo1Path)
		assert.NoError(t, err) // Should not error with mocked git commands
	})

	t.Run("single functions fail with non-git repo", func(t *testing.T) {
		ctx := context.Background()
		nonRepoPath := dirs["nonrepo"]

		// All single functions should fail with non-git directory
		err := pullSingle(ctx, nonRepoPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a git repository")

		err = fetchSingle(ctx, nonRepoPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a git repository")

		err = statusSingle(ctx, nonRepoPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a git repository")
	})
}

func TestRecursiveFlag_Integration(t *testing.T) {
	// Save and restore viper state
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for key, value := range originalConfig {
			viper.Set(key, value)
		}
	}()

	rootPath, _ := testutil.CreateTestDirStructure(t)

	// Set empty skip list for predictable behavior
	viper.Reset()
	viper.Set("skipList", []string{})

	t.Run("recursive flag is available", func(t *testing.T) {
		// Test that the recursive flag exists on the root command
		recursiveFlag := RootCmd.PersistentFlags().Lookup("recursive")
		assert.NotNil(t, recursiveFlag)
		assert.Equal(t, "r", recursiveFlag.Shorthand)
	})

	t.Run("walk operations work", func(t *testing.T) {
		// Test that walkDirectories function works correctly
		var processedPaths []string
		testOperation := func(ctx context.Context, path string) error {
			processedPaths = append(processedPaths, path)
			return nil
		}

		ctx := context.Background()
		err := walkDirectories(ctx, rootPath, testOperation)
		assert.NoError(t, err)
		assert.Greater(t, len(processedPaths), 0)
		assert.Contains(t, processedPaths, rootPath)
	})
}

func TestConfigurationIntegration(t *testing.T) {
	// Save and restore viper state
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for key, value := range originalConfig {
			viper.Set(key, value)
		}
	}()

	t.Run("configuration file loading affects skip behavior", func(t *testing.T) {
		// Save original HOME and create temporary HOME
		originalHome := os.Getenv("HOME")
		tempHome := t.TempDir()
		os.Setenv("HOME", tempHome)
		defer os.Setenv("HOME", originalHome)

		// Create test structure
		rootPath, _ := testutil.CreateAdvancedTestStructure(t)

		// Create config with specific skip list in temp HOME
		configContent := `skipList:
  - custom-skip1
  - custom-skip2
useDefaultSkips: false`
		configFile := filepath.Join(tempHome, ".got.yaml")
		require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

		// Reset viper and initialize config
		viper.Reset()
		cfgFile = ""
		initConfig()

		// Test that skip list is loaded correctly (only custom values, no defaults)
		skipList := getSkipList()
		expected := []string{"custom-skip1", "custom-skip2"}
		assert.ElementsMatch(t, expected, skipList)

		// Test that paths are skipped correctly
		assert.True(t, shouldSkipPath(filepath.Join(rootPath, "custom-skip1/package")))
		assert.True(t, shouldSkipPath(filepath.Join(rootPath, "custom-skip2/lib")))
		assert.False(t, shouldSkipPath(filepath.Join(rootPath, "normal/repo")))
	})
}

func TestErrorHandling_Integration(t *testing.T) {
	t.Run("invalid directory paths are handled by validation", func(t *testing.T) {
		invalidPaths := []string{
			"/path/that/does/not/exist",
			"",
			"go.mod", // file instead of directory
		}

		for _, path := range invalidPaths {
			err := validateDirectoryPath(path)
			assert.Error(t, err, "Should error for invalid path: %s", path)
		}
	})

	t.Run("validation works correctly", func(t *testing.T) {
		// Current directory should be valid
		err := validateDirectoryPath(".")
		assert.NoError(t, err)

		// Temp directory should be valid
		tempDir := testutil.CreateTempNonGitDir(t)
		err = validateDirectoryPath(tempDir)
		assert.NoError(t, err)
	})
}

