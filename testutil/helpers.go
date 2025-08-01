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

// Package testutil provides utilities for testing the Got CLI tool
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateTempGitRepo creates a temporary directory with a .git folder
// Returns the path to the temporary directory
func CreateTempGitRepo(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	return tempDir
}

// CreateTempNonGitDir creates a temporary directory without a .git folder
// Returns the path to the temporary directory
func CreateTempNonGitDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// CreateTestDirStructure creates a complex directory structure for testing
// Returns the root path and a map of created directories
func CreateTestDirStructure(t *testing.T) (rootPath string, dirs map[string]string) {
	t.Helper()

	rootPath = t.TempDir()
	dirs = make(map[string]string)

	// Create various directory types
	structures := []struct {
		name   string
		hasGit bool
		subDir string
	}{
		{"repo1", true, ""},
		{"repo2", true, ""},
		{"nonrepo", false, ""},
		{"nested", false, ""},
		{"nested/repo3", true, ""},
		{"nested/nonrepo2", false, ""},
		{"deep", false, ""},
		{"deep/nested", false, ""},
		{"deep/nested/repo4", true, ""},
	}

	for _, s := range structures {
		dirPath := filepath.Join(rootPath, s.name)
		require.NoError(t, os.MkdirAll(dirPath, 0755))
		dirs[s.name] = dirPath

		if s.hasGit {
			gitDir := filepath.Join(dirPath, ".git")
			require.NoError(t, os.Mkdir(gitDir, 0755))
		}
	}

	return rootPath, dirs
}

// CreateConfigFile creates a temporary configuration file with given content
// Returns the path to the configuration file
func CreateConfigFile(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".got.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	return configFile
}

// CreateSkipListConfig creates a configuration file with a skip list
func CreateSkipListConfig(t *testing.T, skipList []string) string {
	t.Helper()

	content := "skipList:\n"
	for _, item := range skipList {
		content += "  - " + item + "\n"
	}

	return CreateConfigFile(t, content)
}

// GitRepoInfo represents information about a git repository in test structure
type GitRepoInfo struct {
	Path    string
	IsGit   bool
	Skipped bool
}

// CreateAdvancedTestStructure creates a more complex test structure with various scenarios
func CreateAdvancedTestStructure(t *testing.T) (rootPath string, repos []GitRepoInfo) {
	t.Helper()

	rootPath = t.TempDir()

	// Define test structure with various scenarios
	testRepos := []struct {
		path     string
		isGit    bool
		skip     bool
		skipName string // What would cause it to be skipped
	}{
		{"repo1", true, false, ""},
		{"repo2", true, false, ""},
		{"nonrepo", false, false, ""},
		{"node_modules/some-package", false, true, "node_modules"},
		{"node_modules/package/nested", true, true, "node_modules"},
		{".git/hooks", false, true, ".git"},
		{"vendor/github.com/pkg", true, true, "vendor"},
		{"normal/repo", true, false, ""},
		{"skip-me", true, true, "skip-me"},
		{"deeply/nested/structure/repo", true, false, ""},
	}

	for _, tr := range testRepos {
		fullPath := filepath.Join(rootPath, tr.path)
		require.NoError(t, os.MkdirAll(fullPath, 0755))

		if tr.isGit {
			gitDir := filepath.Join(fullPath, ".git")
			require.NoError(t, os.Mkdir(gitDir, 0755))
		}

		repos = append(repos, GitRepoInfo{
			Path:    fullPath,
			IsGit:   tr.isGit,
			Skipped: tr.skip,
		})
	}

	return rootPath, repos
}

// CreateFileInDir creates a file with given content in the specified directory
func CreateFileInDir(t *testing.T, dir, filename, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, filename)
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	return filePath
}

// AssertDirectoryExists verifies that a directory exists
func AssertDirectoryExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "Directory should exist: %s", path)
	require.True(t, info.IsDir(), "Path should be a directory: %s", path)
}

// AssertFileExists verifies that a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "File should exist: %s", path)
	require.False(t, info.IsDir(), "Path should be a file: %s", path)
}

// AssertIsGitRepo verifies that a directory contains a .git folder
func AssertIsGitRepo(t *testing.T, path string) {
	t.Helper()

	gitPath := filepath.Join(path, ".git")
	AssertDirectoryExists(t, gitPath)
}

// CountGitRepos counts the number of git repositories in a slice of GitRepoInfo
func CountGitRepos(repos []GitRepoInfo) int {
	count := 0
	for _, repo := range repos {
		if repo.IsGit {
			count++
		}
	}
	return count
}

// CountSkippedRepos counts the number of skipped repositories in a slice of GitRepoInfo
func CountSkippedRepos(repos []GitRepoInfo) int {
	count := 0
	for _, repo := range repos {
		if repo.Skipped {
			count++
		}
	}
	return count
}

// FilterGitRepos returns only the git repositories from a slice of GitRepoInfo
func FilterGitRepos(repos []GitRepoInfo) []GitRepoInfo {
	var gitRepos []GitRepoInfo
	for _, repo := range repos {
		if repo.IsGit {
			gitRepos = append(gitRepos, repo)
		}
	}
	return gitRepos
}

// FilterNonSkippedRepos returns only the non-skipped repositories from a slice of GitRepoInfo
func FilterNonSkippedRepos(repos []GitRepoInfo) []GitRepoInfo {
	var nonSkipped []GitRepoInfo
	for _, repo := range repos {
		if !repo.Skipped {
			nonSkipped = append(nonSkipped, repo)
		}
	}
	return nonSkipped
}

// MockGitCommandRunner is a mock implementation of GitCommandRunnerInterface for testing
type MockGitCommandRunner struct {
	// Commands stores the commands that were executed for verification
	Commands [][]string
	// Outputs maps command strings to output that should be returned
	Outputs map[string]string
	// Errors maps command strings to errors that should be returned
	Errors map[string]error
}

// NewMockGitCommandRunner creates a new mock git command runner
func NewMockGitCommandRunner() *MockGitCommandRunner {
	return &MockGitCommandRunner{
		Commands: make([][]string, 0),
		Outputs:  make(map[string]string),
		Errors:   make(map[string]error),
	}
}

// RunGitCommand mocks git command execution
func (m *MockGitCommandRunner) RunGitCommand(ctx context.Context, path string, args []string) ([]byte, error) {
	// Record the command that was executed
	m.Commands = append(m.Commands, args)
	
	// Create a key from the git command (excluding work-tree and git-dir args)
	var gitArgs []string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--work-tree=") && !strings.HasPrefix(arg, "--git-dir=") {
			gitArgs = append(gitArgs, arg)
		}
	}
	cmdKey := strings.Join(gitArgs, " ")
	
	// Return configured error if exists
	if err, exists := m.Errors[cmdKey]; exists {
		return nil, err
	}
	
	// Return configured output if exists
	if output, exists := m.Outputs[cmdKey]; exists {
		return []byte(output), nil
	}
	
	// Default successful response
	return []byte("mock git output"), nil
}

// SetOutput configures the mock to return specific output for a git command
func (m *MockGitCommandRunner) SetOutput(command, output string) {
	m.Outputs[command] = output
}

// SetError configures the mock to return an error for a git command
func (m *MockGitCommandRunner) SetError(command string, err error) {
	m.Errors[command] = err
}

// GetExecutedCommands returns all commands that were executed
func (m *MockGitCommandRunner) GetExecutedCommands() [][]string {
	return m.Commands
}

// GitCommandRunnerInterface defines the interface for git command execution
// This is defined here to avoid circular imports
type GitCommandRunnerInterface interface {
	RunGitCommand(ctx context.Context, path string, args []string) ([]byte, error)
}

// InstallMockGitRunner installs a mock git command runner for testing
// This function should be called from test files in the cmd package
// Returns the mock runner and a cleanup function
func InstallMockGitRunner(t *testing.T, setRunnerFunc func(GitCommandRunnerInterface) GitCommandRunnerInterface) (*MockGitCommandRunner, func()) {
	t.Helper()
	mock := NewMockGitCommandRunner()
	
	// Install the mock using the provided setter function
	originalRunner := setRunnerFunc(mock)
	
	cleanup := func() {
		t.Helper()
		// Restore the original runner
		setRunnerFunc(originalRunner)
	}
	
	return mock, cleanup
}
