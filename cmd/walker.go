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

	"github.com/id9051/got/internal/git"
	"github.com/pkg/errors"
)

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