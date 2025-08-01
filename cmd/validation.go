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
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/id9051/got/internal/git"
	"github.com/pkg/errors"
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