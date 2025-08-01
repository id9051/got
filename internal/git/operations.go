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
	"fmt"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// Output stores the output of a git command for later display
type Output struct {
	Path   string
	Output string
	Error  error
}

// OperationConfig contains configuration for git operations
type OperationConfig struct {
	ProgressMode      bool
	OutputBufferPtr   *[]Output
	LogSkipped        func(string)
	LogSuccess        func(string)
	LogError          func(string, error)
	ShowSpinner       func(string, string) (chan bool, error)
	StyleProgress     func(string) string
	StylePath         func(string) string
}

// ExecuteCommand executes a git command in the specified directory with context
// For recursive operations - silently skips non-git directories
func ExecuteCommand(ctx context.Context, path string, config *OperationConfig, gitArgs ...string) error {
	// Skip non-git directories silently during recursive operations
	if !IsRepository(path) {
		return nil
	}

	return runCommand(ctx, path, config, gitArgs...)
}

// ExecuteCommandSingle executes a git command on a single directory with context
// For single directory operations - returns error if not a git repository
func ExecuteCommandSingle(ctx context.Context, path string, config *OperationConfig, gitArgs ...string) error {
	if !IsRepository(path) {
		return errors.Errorf("[%s] is not a git repository", path)
	}

	return runCommand(ctx, path, config, gitArgs...)
}

// runCommand is the shared implementation for running git commands with context
func runCommand(ctx context.Context, path string, config *OperationConfig, gitArgs ...string) error {
	// Build git command with explicit work-tree and git-dir
	args := []string{
		fmt.Sprintf("--work-tree=%s", path),
		fmt.Sprintf("--git-dir=%s", filepath.Join(path, DirName)),
	}
	args = append(args, gitArgs...)

	// For status command, we want to capture output
	var output []byte
	var err error
	if len(gitArgs) > 0 && gitArgs[0] == "status" {
		// Capture output instead of sending directly to stdout to avoid interfering with progress bar
		output, err = runner.RunGitCommand(ctx, path, args)
	}

	// Show operation in progress
	operation := "operation"
	if len(gitArgs) > 0 {
		operation = gitArgs[0]
	}

	// Start spinner for non-status commands
	var done chan bool
	if operation != "status" && config != nil && config.ShowSpinner != nil && !config.ProgressMode {
		// Only show spinner when not in progress mode
		var spinnerErr error
		done, spinnerErr = config.ShowSpinner(operation, path)
		if spinnerErr != nil {
			return spinnerErr
		}
	}

	if operation != "status" {
		// Run the command for non-status operations
		_, err = runner.RunGitCommand(ctx, path, args)
		if done != nil {
			close(done)
			time.Sleep(50 * time.Millisecond) // Brief pause to ensure spinner cleanup
		}
	}
	// For status commands, output was already captured above

	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err != nil {
		if config != nil && config.LogError != nil {
			config.LogError(path, err)
		}
		return nil // Don't stop processing other repositories
	}

	// Handle output display based on mode
	if operation == "status" && len(output) > 0 && config != nil {
		if config.ProgressMode {
			// Buffer the output for later display - we need to modify the slice in place
			*config.OutputBufferPtr = append(*config.OutputBufferPtr, Output{
				Path:   path,
				Output: string(output),
				Error:  nil,
			})
		} else {
			// Display immediately for single operations
			fmt.Print(string(output))
		}
	}

	// Always log success
	if config != nil && config.LogSuccess != nil {
		if config.ProgressMode {
			// In progress mode, print success immediately after clearing progress line
			fmt.Print("\r\033[K") // Clear progress line
			config.LogSuccess(path)
		} else {
			config.LogSuccess(path)
		}
	}
	return nil
}