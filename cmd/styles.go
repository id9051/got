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

	"github.com/charmbracelet/lipgloss"
)

// Color palette for consistent theming - optimized for dark backgrounds
var (
	primaryColor   = lipgloss.Color("#00D787") // Brighter green for dark backgrounds
	secondaryColor = lipgloss.Color("#5FAFFF") // Brighter blue
	successColor   = lipgloss.Color("#00FF87") // Bright success green
	warningColor   = lipgloss.Color("#FFAF00") // Bright warning orange
	errorColor     = lipgloss.Color("#FF5F5F") // Bright error red
	mutedColor     = lipgloss.Color("#AAAAAA") // Lighter gray for better visibility
	accentColor    = lipgloss.Color("#FF5FAF") // Bright pink accent
)

// Base styles for different message types
var (
	// Success style with checkmark
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Error style with X mark
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Warning/skip style with warning icon
	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// Info style for general information
	infoStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Muted style for less important info
	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Path style for highlighting file paths
	pathStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Count/number style
	numberStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Progress style
	progressStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Header style for section headers
	headerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Underline(true)

	// Box style for important messages
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)

	// Summary box style
	summaryStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(successColor).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1).
			Bold(true)
)

// Icons for different message types
const (
	checkIcon   = "✓"
	crossIcon   = "✗"
	warningIcon = "⚠"
	infoIcon    = "ℹ"
	arrowIcon   = "→"
	dotIcon     = "•"
	searchIcon  = "🔍"
	gitIcon     = "📁"
	skipIcon    = "⏭"
	summaryIcon = "📊"
	rocketIcon  = "🚀"
)

// Styled message functions
func styleSuccess(path string) string {
	return successStyle.Render(checkIcon+" ") + pathStyle.Render(path) + successStyle.Render(" Success")
}

func styleError(path string, err error) string {
	return errorStyle.Render(crossIcon+" ") + pathStyle.Render(path) + errorStyle.Render(" ERROR ") + err.Error()
}

func styleSkipped(path string) string {
	return warningStyle.Render(skipIcon+" Skipping ") + pathStyle.Render(path)
}

func styleProgress(message string) string {
	return progressStyle.Render(searchIcon + " " + message)
}

func styleSummary(message string) string {
	return summaryStyle.Render(summaryIcon + " " + message)
}

func styleInfo(message string) string {
	return infoStyle.Render(infoIcon + " " + message)
}

func styleHeader(message string) string {
	return headerStyle.Render(rocketIcon + " " + message)
}

func stylePath(path string) string {
	return pathStyle.Render(path)
}

func styleNumber(num int) string {
	return numberStyle.Render(fmt.Sprintf("%d", num))
}

// Helper function to style git command descriptions
func styleGitCommand(command string) string {
	return infoStyle.Render(gitIcon + " git " + command)
}

// Box wrapper for important messages
func styleBox(content string) string {
	return boxStyle.Render(content)
}

// Enhanced command descriptions with styling
func getStyledDescription(baseDesc string, examples []string) string {
	styled := baseDesc + "\n\n"

	if len(examples) > 0 {
		styled += headerStyle.Render("Examples:") + "\n"
		for _, example := range examples {
			styled += mutedStyle.Render("  "+dotIcon+" ") + pathStyle.Render(example) + "\n"
		}
	}

	return styled
}
