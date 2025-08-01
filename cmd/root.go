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

// Package cmd provides the command-line interface for the Got CLI tool.
// It implements Cobra commands for git repository management operations
// including pull, fetch, and status commands that can operate on single
// repositories or recursively across directory trees.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// Global context for cancellation support
var globalCtx context.Context
var globalCancel context.CancelFunc

// getSkipList returns the skip list from configuration, with defaults if not configured
func getSkipList() []string {
	skipList := viper.GetStringSlice("skipList")

	// Check if user wants to disable default skips (defaults to true)
	useDefaults := true
	if viper.IsSet("useDefaultSkips") {
		useDefaults = viper.GetBool("useDefaultSkips")
	}

	// Default directories that are commonly skipped
	defaultSkips := []string{"node_modules", "vendor", ".git"}

	// Merge with configured skip list
	skipMap := make(map[string]bool)

	// Only add defaults if enabled
	if useDefaults {
		for _, skip := range defaultSkips {
			skipMap[skip] = true
		}
	}

	for _, skip := range skipList {
		skipMap[skip] = true
	}

	// Convert map back to slice
	validSkipList := make([]string, 0, len(skipMap))
	for skip := range skipMap {
		// Remove empty strings and whitespace-only entries
		skip = strings.TrimSpace(skip)
		if skip != "" {
			validSkipList = append(validSkipList, skip)
		}
	}

	return validSkipList
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "got",
	Short: "Git repository management tool",
	Long: `Got is a CLI tool for managing multiple Git repositories.

It allows you to perform git operations (pull, fetch, status) across single 
repositories or recursively across directory trees containing multiple git 
repositories. Use the --recursive flag to operate on all repositories found 
in subdirectories.

Configuration (.got.yaml in your home directory):
  skipList: ["custom_dir", "temp"]           # Custom directories to skip
  useDefaultSkips: true                      # Include defaults (node_modules, vendor, .git)

By default, common directories (node_modules, vendor, .git) are automatically 
skipped. Set useDefaultSkips: false to disable this behavior.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Create context that can be cancelled
	globalCtx, globalCancel = context.WithCancel(context.Background())
	defer globalCancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n" + styleInfo("Received interrupt signal, cancelling operations..."))
		globalCancel()
	}()

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(styleError("Error", err))
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set custom help function
	RootCmd.SetHelpFunc(styledHelp)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.got.yaml)")
	RootCmd.PersistentFlags().BoolP("recursive", "r", false, "Recursively operate on subdirectories")

	// Enable completion command
	RootCmd.CompletionOptions.DisableDefaultCmd = false
	RootCmd.CompletionOptions.DisableNoDescFlag = false
	RootCmd.CompletionOptions.DisableDescriptions = false

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".got")  // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println(styleInfo("Using config file: " + stylePath(viper.ConfigFileUsed())))
	}
}
