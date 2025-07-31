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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Constants for commonly used strings
const (
	GitDirName          = ".git"
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
	_, err := os.Stat(filepath.Join(path, GitDirName))
	return err == nil
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

// GitOutput stores the output of a git command for later display
type GitOutput struct {
	Path   string
	Output string
	Error  error
}

var gitOutputBuffer []GitOutput
var inProgressMode bool

// executeGitCommand executes a git command in the specified directory
// For recursive operations - silently skips non-git directories
func executeGitCommand(path string, gitArgs ...string) error {
	// Skip non-git directories silently during recursive operations
	if !isGitRepository(path) {
		return nil
	}

	return runGitCommand(path, gitArgs...)
}

// executeGitCommandSingle executes a git command on a single directory
// For single directory operations - returns error if not a git repository
func executeGitCommandSingle(path string, gitArgs ...string) error {
	if !isGitRepository(path) {
		return errors.Errorf("[%s] is not a git repository", path)
	}

	return runGitCommand(path, gitArgs...)
}

// runGitCommand is the shared implementation for running git commands
func runGitCommand(path string, gitArgs ...string) error {
	// Build git command with explicit work-tree and git-dir
	args := []string{
		fmt.Sprintf("--work-tree=%s", path),
		fmt.Sprintf("--git-dir=%s", filepath.Join(path, GitDirName)),
	}
	args = append(args, gitArgs...)

	gitCmd := exec.Command("git", args...)
	
	// For status command, we want to show output to user but need to handle progress bar
	var output []byte
	var err error
	if len(gitArgs) > 0 && gitArgs[0] == "status" {
		// Capture output instead of sending directly to stdout to avoid interfering with progress bar
		output, err = gitCmd.CombinedOutput()
	}

	// Show operation in progress
	operation := "operation"
	if len(gitArgs) > 0 {
		operation = gitArgs[0]
	}
	
	// Start spinner for non-status commands
	spinner := NewSpinner()
	done := make(chan bool)
	
	if operation != "status" {
		if !inProgressMode {
			// Only show spinner when not in progress mode
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
		}
		
		// Run the command for non-status operations
		err = gitCmd.Run()
		close(done)
		
		if !inProgressMode {
			time.Sleep(50 * time.Millisecond) // Brief pause to ensure spinner cleanup
			fmt.Print("\r\033[K") // Clear line
		}
	}
	// For status commands, output was already captured above

	if err != nil {
		logError(path, err)
		return nil // Don't stop processing other repositories
	}

	// Handle output display based on mode
	if operation == "status" && len(output) > 0 {
		if inProgressMode {
			// Buffer the output for later display
			gitOutputBuffer = append(gitOutputBuffer, GitOutput{
				Path:   path,
				Output: string(output),
				Error:  nil,
			})
		} else {
			// Display immediately for single operations
			fmt.Print(string(output))
		}
	}

	// Always log success, but handle it differently in progress mode
	if inProgressMode {
		// In progress mode, print success immediately after clearing progress line
		fmt.Print("\r\033[K") // Clear progress line
		logSuccess(path)
		// Let the progress bar re-render on next update
	} else {
		logSuccess(path)
	}
	return nil
}

// walkDirectories is a generic function for walking directories and applying git operations
func walkDirectories(rootPath string, gitOperation func(string) error) error {
	// Enable progress mode and clear output buffer
	inProgressMode = true
	gitOutputBuffer = []GitOutput{}
	
	// First, count total directories for progress bar (applying same optimization logic)
	totalDirs := 0
	var skipCount int
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		
		// Skip .git directories
		if filepath.Base(path) == GitDirName {
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
		if filepath.Base(path) == GitDirName {
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
			// Apply git operation
			gitOperation(path)
			// Skip subdirectories of git repositories since we only operate on repo roots
			return filepath.SkipDir
		}

		// Apply git operation to non-git directories (will be skipped silently)
		return gitOperation(path)
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