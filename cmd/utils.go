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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/id9051/got/internal/git"
	"github.com/pkg/errors"
)

// SetGitCommandRunner sets the git command runner (for testing)
func SetGitCommandRunner(runner git.CommandRunner) git.CommandRunner {
	return git.SetCommandRunner(runner)
}

// Constants for commonly used strings
const (
	RecursiveFlagName   = "recursive"
	ErrorWalkingMessage = "error walking filepath [%s]"
	SkippingMessage     = "Skipping [%s]"
	SuccessMessage      = "[%s]:  Success"
	ErrorMessage        = "[%s]: ERROR %v"
)

// validateDirectoryPath validates that the given path exists and is accessible
func validateDirectoryPath(path string) error {
	if path == "" {
		return errors.New("directory path cannot be empty")
	}

	// Convert to absolute path for better error messages
	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve absolute path for '%s'", path)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("directory does not exist: '%s'", absPath)
		}
		return errors.Wrapf(err, "failed to access directory: '%s'", absPath)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return errors.Errorf("path is not a directory: '%s'", absPath)
	}

	// Check if directory is readable
	if _, err := os.Open(absPath); err != nil {
		return errors.Wrapf(err, "directory is not accessible: '%s'", absPath)
	}

	return nil
}

// isGitRepository checks if the given path contains a git repository
func isGitRepository(path string) bool {
	return git.IsRepository(path)
}

// shouldSkipPath checks if a path should be skipped based on the skip list
func shouldSkipPath(path string) bool {
	skipList := getSkipList()
	return slices.ContainsFunc(skipList, func(skip string) bool {
		return strings.Contains(path, skip)
	})
}

// logSkipped logs that a path was skipped
func logSkipped(path string) {
	fmt.Println(styleSkipped(path))
}

// logSuccess logs successful operation
func logSuccess(path string) {
	fmt.Println(styleSuccess(path))
}

// logError logs error from operation
func logError(path string, err error) {
	fmt.Println(styleError(path, err))
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

// walkDirectories is a generic function for walking directories and applying git operations
func walkDirectories(ctx context.Context, rootPath string, gitOperation func(context.Context, string) error) error {
	// Enable progress mode and clear output buffer
	inProgressMode = true
	gitOutputBuffer = []git.Output{}

	// First, count total directories for progress bar (applying same optimization logic)
	totalDirs := 0
	var skipCount int
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}

		// Skip .git directories
		if filepath.Base(path) == git.DirName {
			return filepath.SkipDir
		}

		// Skip paths in skip list during counting too
		if shouldSkipPath(path) {
			return filepath.SkipDir
		}

		// Count this directory
		totalDirs++

		// If this is a git repository, skip its subdirectories in counting
		if isGitRepository(path) {
			return filepath.SkipDir
		}

		return nil
	})

	// Show initial progress message
	fmt.Println(styleProgress("Recursively scanning directories under " + stylePath(rootPath) + "..."))
	fmt.Printf(styleInfo("Found %s directories to process"), numberStyle.Render(fmt.Sprintf("%d", totalDirs)))
	fmt.Println()
	fmt.Println()

	// Create progress tracker
	progress := NewProgressTracker()
	progress.SetTotal(totalDirs)
	progress.Start()

	dirCount := 0
	gitRepoCount := 0

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// Check for context cancellation first
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Handle walking errors (e.g., deleted directories during walk)
		if err != nil {
			fmt.Println(styleError(path, errors.Wrapf(err, "error walking filepath")))
			return nil // Continue processing
		}

		// Skip non-directories
		if !info.IsDir() {
			return nil
		}

		dirCount++

		// Skip .git directories
		if filepath.Base(path) == git.DirName {
			return filepath.SkipDir
		}

		// Update progress
		isGit := isGitRepository(path)
		progress.Update(path, isGit)

		// Skip paths in skip list
		if shouldSkipPath(path) {
			// Show skip message through progress tracker
			progress.ShowMessage(styleSkipped(path))
			skipCount++
			return filepath.SkipDir
		}

		// Check if this is a git repository before applying operation
		if isGit {
			gitRepoCount++
			// Apply git operation with context
			if err := gitOperation(ctx, path); err != nil && err == context.Canceled {
				return err // Propagate cancellation
			}
			// Skip subdirectories of git repositories since we only operate on repo roots
			return filepath.SkipDir
		}

		// Apply git operation to non-git directories (will be skipped silently)
		if err := gitOperation(ctx, path); err != nil && err == context.Canceled {
			return err // Propagate cancellation
		}
		return nil
	})

	// Finish progress display
	progress.Finish()

	// Disable progress mode
	inProgressMode = false

	// Display buffered git outputs
	if len(gitOutputBuffer) > 0 {
		fmt.Println() // Add space after progress
		for _, output := range gitOutputBuffer {
			if output.Error != nil {
				logError(output.Path, output.Error)
			} else {
				fmt.Print(output.Output)
				logSuccess(output.Path)
			}
		}
	}

	// Show completion summary
	fmt.Println() // Add space after progress

	summaryMsg := ""
	if gitRepoCount > 0 {
		summaryMsg = fmt.Sprintf("Completed recursive operation on %s git repositories (scanned %s directories",
			numberStyle.Render(fmt.Sprintf("%d", gitRepoCount)),
			numberStyle.Render(fmt.Sprintf("%d", dirCount)))
	} else {
		summaryMsg = fmt.Sprintf("No git repositories found (scanned %s directories",
			numberStyle.Render(fmt.Sprintf("%d", dirCount)))
	}

	// Add skip count if any
	if skipCount > 0 {
		summaryMsg += fmt.Sprintf(", skipped %s", numberStyle.Render(fmt.Sprintf("%d", skipCount)))
	}
	summaryMsg += ")"

	if gitRepoCount > 0 {
		fmt.Println(styleSummary(summaryMsg))
	} else {
		fmt.Println(styleInfo(summaryMsg))
	}

	return err
}
