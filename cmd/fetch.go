// Copyright © 2025 Jeff Durham <jeffrey.durham@gmail.com>
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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch directory",
	Short: "Fetch changes from remote repositories",
	Long: `Fetch changes from remote repositories in the specified directory without merging.

If the --recursive flag is used, got will walk through all subdirectories
and fetch changes from any Git repositories found. Directories specified
in the skip list configuration will be ignored during recursive operations.`,
	Example: `got fetch .                    # Fetch changes in current directory
got fetch /path/to/repo        # Fetch changes in specific directory
got fetch -r /path/to/projects # Recursively fetch all repositories`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("directory argument is required")
		}
		
		// Validate directory path
		if err := validateDirectoryPath(args[0]); err != nil {
			return err
		}
		
		recursive, err := cmd.Flags().GetBool(RecursiveFlagName)
		if err != nil {
			return errors.Wrap(err, "failed to get recursive flag")
		}
		
		if recursive {
			return walkDirectories(args[0], func(path string) error {
				return executeGitCommand(path, "fetch")
			})
		}
		return fetchSingle(args[0])
	},
}

func init() {
	RootCmd.AddCommand(fetchCmd)
	fetchCmd.SetHelpFunc(styledHelp)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// fetchSingle performs git fetch on a single directory
func fetchSingle(path string) error {
	if shouldSkipPath(path) {
		logSkipped(path)
		return nil
	}
	return executeGitCommandSingle(path, "fetch")
}

// fetchWalk is deprecated - functionality moved to walkDirectories in utils.go
// Kept for backward compatibility but now just calls the generic walker
func fetchWalk(path string) error {
	return walkDirectories(path, func(path string) error {
		return executeGitCommand(path, "fetch")
	})
}
