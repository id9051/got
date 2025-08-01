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

// Package git provides git command operations and repository detection
package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// CommandRunner defines the interface for executing git commands
// This allows for mocking during tests
type CommandRunner interface {
	RunGitCommand(ctx context.Context, path string, args []string) ([]byte, error)
}

// RealCommandRunner implements CommandRunner using actual git commands
type RealCommandRunner struct{}

// RunGitCommand executes a real git command
func (r *RealCommandRunner) RunGitCommand(ctx context.Context, path string, args []string) ([]byte, error) {
	gitCmd := exec.CommandContext(ctx, "git", args...)
	return gitCmd.CombinedOutput()
}

// runner is the global git command runner, can be replaced for testing
var runner CommandRunner = &RealCommandRunner{}

// SetCommandRunner sets the git command runner (for testing)
func SetCommandRunner(cmdRunner CommandRunner) CommandRunner {
	oldRunner := runner
	runner = cmdRunner
	return oldRunner
}

// Constants for commonly used git-related strings
const (
	DirName = ".git"
)

// IsRepository checks if the given path contains a git repository
func IsRepository(path string) bool {
	_, err := os.Stat(filepath.Join(path, DirName))
	return err == nil
}

// RunCommand executes a git command with the given context and arguments
func RunCommand(ctx context.Context, path string, args []string) ([]byte, error) {
	// Build git command with explicit work-tree and git-dir
	gitArgs := []string{
		"--work-tree=" + path,
		"--git-dir=" + filepath.Join(path, DirName),
	}
	gitArgs = append(gitArgs, args...)

	return runner.RunGitCommand(ctx, path, gitArgs)
}