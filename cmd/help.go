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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Custom help function with styled output
func styledHelp(cmd *cobra.Command, args []string) {
	// Header
	fmt.Println()
	fmt.Println(styleHeader("Got - Git Repository Management Tool"))
	fmt.Println()
	
	// Description
	if cmd.Long != "" {
		fmt.Println(cmd.Long)
	} else if cmd.Short != "" {
		fmt.Println(cmd.Short)
	}
	fmt.Println()
	
	// Usage
	if cmd.HasSubCommands() {
		fmt.Println(headerStyle.Render("Usage:"))
		fmt.Println("  " + pathStyle.Render(cmd.Use) + " [command]")
		fmt.Println()
		
		// Available Commands
		fmt.Println(headerStyle.Render("Available Commands:"))
		for _, subcmd := range cmd.Commands() {
			if !subcmd.Hidden {
				name := pathStyle.Render(fmt.Sprintf("  %-12s", subcmd.Name()))
				desc := mutedStyle.Render(subcmd.Short)
				fmt.Printf("%s %s\n", name, desc)
			}
		}
		fmt.Println()
	} else {
		fmt.Println(headerStyle.Render("Usage:"))
		fmt.Println("  " + pathStyle.Render(cmd.Use))
		fmt.Println()
	}
	
	// Flags
	if cmd.HasAvailableFlags() {
		fmt.Println(headerStyle.Render("Flags:"))
		printStyledFlags(cmd.Flags())
		fmt.Println()
	}
	
	// Examples
	if cmd.Example != "" {
		fmt.Println(headerStyle.Render("Examples:"))
		examples := strings.Split(cmd.Example, "\n")
		for _, example := range examples {
			if strings.TrimSpace(example) != "" {
				fmt.Println("  " + mutedStyle.Render(dotIcon) + " " + pathStyle.Render(example))
			}
		}
		fmt.Println()
	}
	
	// Footer
	if cmd.HasSubCommands() {
		fmt.Println(mutedStyle.Render("Use \"" + cmd.Name() + " [command] --help\" for more information about a command."))
		fmt.Println()
	}
}

func printStyledFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		
		// Format flag
		flagStr := "  "
		if flag.Shorthand != "" {
			flagStr += infoStyle.Render("-"+flag.Shorthand+", ")
		} else {
			flagStr += "    "
		}
		flagStr += infoStyle.Render("--" + flag.Name)
		
		// Add type for non-bool flags
		if flag.Value.Type() != "bool" {
			flagStr += " " + mutedStyle.Render(flag.Value.Type())
		}
		
		// Add description
		if flag.Usage != "" {
			flagStr += "   " + flag.Usage
		}
		
		// Add default value if set
		if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "[]" {
			flagStr += " " + mutedStyle.Render("(default: "+flag.DefValue+")")
		}
		
		fmt.Println(flagStr)
	})
}