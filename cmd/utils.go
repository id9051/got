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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

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
	log.Printf(SkippingMessage+"\n", path)
}

// logSuccess logs successful operation
func logSuccess(path string) {
	log.Printf(SuccessMessage+"\n", path)
}

// logError logs error from operation
func logError(path string, err error) {
	log.Printf(ErrorMessage+"\n", path, err)
}

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
	
	// For status command, we want to show output to user
	if len(gitArgs) > 0 && gitArgs[0] == "status" {
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr
	}

	if err := gitCmd.Run(); err != nil {
		logError(path, err)
		return nil // Don't stop processing other repositories
	}

	logSuccess(path)
	return nil
}

// walkDirectories is a generic function for walking directories and applying git operations
func walkDirectories(rootPath string, gitOperation func(string) error) error {
	// Show initial progress message for recursive operations
	log.Printf("Recursively scanning directories under %s...", rootPath)
	
	dirCount := 0
	gitRepoCount := 0
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// Handle walking errors (e.g., deleted directories during walk)
		if err != nil {
			log.Println(errors.Wrapf(err, ErrorWalkingMessage, path).Error())
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

		// Skip paths in skip list
		if shouldSkipPath(path) {
			logSkipped(path)
			return filepath.SkipDir
		}

		// Check if this is a git repository before applying operation
		if isGitRepository(path) {
			gitRepoCount++
		}

		// Apply git operation
		return gitOperation(path)
	})
	
	// Show completion summary
	if gitRepoCount > 0 {
		log.Printf("Completed recursive operation on %d git repositories (scanned %d directories)", gitRepoCount, dirCount)
	} else {
		log.Printf("No git repositories found (scanned %d directories)", dirCount)
	}
	
	return err
}