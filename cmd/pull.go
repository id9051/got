// Copyright © 2016 Jeff Durham <jeffrey.durham@gmail.com>
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
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull directory",
	Short: "Pull changes from remote repositories",
	Long: `Pull changes from remote repositories in the specified directory.

If the --recursive flag is used, got will walk through all subdirectories
and pull changes from any Git repositories found. Directories specified
in the skip list configuration will be ignored during recursive operations.

Examples:
  got pull .                    # Pull changes in current directory
  got pull /path/to/repo        # Pull changes in specific directory
  got pull -r /path/to/projects # Recursively pull all repositories`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("directory argument is required")
		}
		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return errors.Wrap(err, "failed to get recursive flag")
		}
		if recursive {
			return pullWalk(args[0])
		}
		return pull(args[0], recursive)
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func pull(path string, recursive bool) error {

	skipList := getSkipList()
	if slices.ContainsFunc(skipList, func(skip string) bool {
		return strings.Contains(path, skip)
	}) {
		log.Printf("Skipping [%s]\n", path)
		return nil
	}

	_, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {

		if recursive {
			return nil
		}

		return errors.Wrapf(err, "[%s] is not a git repository", path)
	}

	pullCmd := exec.Command("git", fmt.Sprintf("--work-tree=%s", path), fmt.Sprintf("--git-dir=%s", filepath.Join(path, ".git")), "pull")

	if err := pullCmd.Run(); err != nil {
		log.Printf("[%s]: ERROR %v\n", path, err)
	} else {
		log.Printf("[%s]:  Success\n", path)
	}

	return nil
}

func pullWalk(path string) error {

	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		// Usually usually happens when a director is deleted. If exists when filepath.Walk
		// is called but then the pull removes it. So we get a "No such file or directory"
		// error. We're returning nil so that processing continues.
		if err != nil {
			log.Println(errors.Wrapf(err, "error walking filepath [%s]", path).Error())
			return nil
		}

		if !info.IsDir() {
			return nil
		} else if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		} else if slices.ContainsFunc(getSkipList(), func(skip string) bool {
			return strings.Contains(path, skip)
		}) {
			log.Printf("Skipping [%s]\n", path)
			return filepath.SkipDir
		}

		return pull(path, true)
	})
}
