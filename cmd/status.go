// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("directory argument is required")
		}
		if recursive {
			return statusWalk(args[0])
		}
		return status(args[0])
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	statusCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively check status of subdirectories listed")
}

func status(path string) error {

	_, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		if recursive {
			return nil
		}
		return errors.Wrapf(err, "[%s] is not a git repository", path)
	}

	statusCmd := exec.Command("git", fmt.Sprintf("--work-tree=%s", path), fmt.Sprintf("--git-dir=%s", filepath.Join(path, ".git")), "status")
	statusCmd.Stdout = os.Stdout
	statusCmd.Stderr = os.Stderr

	if err := statusCmd.Run(); err != nil {
		log.Printf("[%s]: ERROR %v\n", path, err)
	} else {
		log.Printf("[%s]:  Success\n", path)
	}

	return nil
}

func statusWalk(path string) error {

	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() {
			return nil
		} else if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		return status(path)
	})
}
