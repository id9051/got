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
	"fmt"
	"time"

	"github.com/id9051/got/internal/git"
)

// Constants for commonly used strings
const (
	RecursiveFlagName = "recursive"
)

// SetGitCommandRunner sets the git command runner (for testing)
func SetGitCommandRunner(runner git.CommandRunner) git.CommandRunner {
	return git.SetCommandRunner(runner)
}

var gitOutputBuffer []git.Output
var inProgressMode bool

// executeGitCommand executes a git command in the specified directory with context
// For recursive operations - silently skips non-git directories
func executeGitCommand(ctx context.Context, path string, gitArgs ...string) error {
	config := &git.OperationConfig{
		ProgressMode:    inProgressMode,
		OutputBufferPtr: &gitOutputBuffer,
		LogSkipped:      logSkipped,
		LogSuccess:      logSuccess,
		LogError:        logError,
		ShowSpinner:     showSpinner,
	}

	return git.ExecuteCommand(ctx, path, config, gitArgs...)
}

// executeGitCommandSingle executes a git command on a single directory with context
// For single directory operations - returns error if not a git repository
func executeGitCommandSingle(ctx context.Context, path string, gitArgs ...string) error {
	config := &git.OperationConfig{
		ProgressMode:    inProgressMode,
		OutputBufferPtr: &gitOutputBuffer,
		LogSkipped:      logSkipped,
		LogSuccess:      logSuccess,
		LogError:        logError,
		ShowSpinner:     showSpinner,
	}

	return git.ExecuteCommandSingle(ctx, path, config, gitArgs...)
}

// showSpinner creates and manages a spinner for git operations
func showSpinner(operation, path string) (chan bool, error) {
	spinner := NewSpinner()
	done := make(chan bool)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				fmt.Print("\r\033[K") // Clear line
				return
			case <-ticker.C:
				fmt.Printf("\r%s %s %s",
					spinner.Next(),
					infoStyle.Render("Running git "+operation+" on"),
					pathStyle.Render(path))
			}
		}
	}()

	return done, nil
}