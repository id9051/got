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
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// ProgressTracker manages progress display for operations
type ProgressTracker struct {
	mu             sync.Mutex
	total          int
	current        int
	currentPath    string
	gitRepoCount   int
	prog           progress.Model
	lastUpdate     time.Time
	updateInterval time.Duration
	showProgress   bool
	startTime      time.Time
	lastETAUpdate  time.Time
	etaUpdateInterval time.Duration
	cachedETA      string
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	// Create a styled progress bar
	prog := progress.New(progress.WithDefaultGradient())
	prog.ShowPercentage = false // We'll show our own percentage
	prog.Width = 50             // Make it wider

	// Style the progress bar with our colors - optimized for dark backgrounds
	prog.Full = '█'
	prog.Empty = '░'
	prog.FullColor = string(primaryColor)
	prog.EmptyColor = "#444444" // Darker gray for empty sections to contrast with text

	return &ProgressTracker{
		prog:           prog,
		updateInterval: 50 * time.Millisecond, // Faster updates for better visibility
		showProgress:   true,
		etaUpdateInterval: 1 * time.Second,   // Update ETA every second to avoid flickering
		cachedETA:      "calculating...",
	}
}

// SetTotal sets the total number of items to process
func (pt *ProgressTracker) SetTotal(total int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.total = total
}

// Start begins progress tracking
func (pt *ProgressTracker) Start() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	now := time.Now()
	pt.startTime = now
	pt.lastUpdate = now
	pt.lastETAUpdate = now
	if pt.showProgress {
		// Hide cursor
		fmt.Print("\033[?25l") // Hide cursor
		pt.render()
	}
}

// Update updates the progress with current path
func (pt *ProgressTracker) Update(path string, isGitRepo bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.current++
	pt.currentPath = path
	if isGitRepo {
		pt.gitRepoCount++
	}

	// Only update display if enough time has passed
	if time.Since(pt.lastUpdate) >= pt.updateInterval {
		pt.lastUpdate = time.Now()
		if pt.showProgress {
			pt.render()
		}
	}
}

// Finish completes the progress tracking
func (pt *ProgressTracker) Finish() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.showProgress {
		// Clear the progress line and show cursor
		fmt.Print("\r\033[K")  // Clear current line
		fmt.Print("\033[?25h") // Show cursor again
	}
}

// calculateETA calculates estimated time remaining
func (pt *ProgressTracker) calculateETA() string {
	if pt.current == 0 || pt.total == 0 {
		return "calculating..."
	}

	elapsed := time.Since(pt.startTime)
	if elapsed < time.Second {
		return "calculating..."
	}

	// Calculate rate (items per second)
	rate := float64(pt.current) / elapsed.Seconds()
	if rate == 0 {
		return "calculating..."
	}

	// Calculate remaining items and time
	remaining := pt.total - pt.current
	if remaining <= 0 {
		return "0s"
	}

	etaSeconds := float64(remaining) / rate
	etaDuration := time.Duration(etaSeconds * float64(time.Second))

	// Format duration nicely
	if etaDuration < time.Minute {
		return fmt.Sprintf("%ds", int(etaDuration.Seconds()))
	} else if etaDuration < time.Hour {
		return fmt.Sprintf("%dm%ds", int(etaDuration.Minutes()), int(etaDuration.Seconds())%60)
	} else {
		return fmt.Sprintf("%dh%dm", int(etaDuration.Hours()), int(etaDuration.Minutes())%60)
	}
}

// render displays the current progress
func (pt *ProgressTracker) render() {
	if pt.total == 0 {
		return
	}

	percent := float64(pt.current) / float64(pt.total)
	if percent > 1.0 {
		percent = 1.0
	}

	// Create the progress bar view
	bar := pt.prog.ViewAs(percent)

	// Update ETA periodically to avoid flickering
	now := time.Now()
	if now.Sub(pt.lastETAUpdate) >= pt.etaUpdateInterval {
		pt.cachedETA = pt.calculateETA()
		pt.lastETAUpdate = now
	}

	// Build status line with ETA
	status := fmt.Sprintf("Progress: %s %3.0f%% [%d/%d dirs, %d git repos found] ETA: %s",
		bar,
		percent*100,
		pt.current,
		pt.total,
		pt.gitRepoCount,
		pt.cachedETA,
	)

	// Simple overwrite - just print with carriage return
	fmt.Printf("\r%s", infoStyle.Render(status))
}

// GetGitRepoCount returns the number of git repositories found
func (pt *ProgressTracker) GetGitRepoCount() int {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	return pt.gitRepoCount
}

// GetProcessedCount returns the number of directories processed
func (pt *ProgressTracker) GetProcessedCount() int {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	return pt.current
}

// ShowMessage temporarily displays a message without disrupting progress
func (pt *ProgressTracker) ShowMessage(message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.showProgress {
		// Clear current line and show message
		fmt.Print("\r\033[K" + message + "\n")
		// Redraw progress on next line
		pt.render()
	} else {
		// If not showing progress, just print the message
		fmt.Println(message)
	}
}

// SimpleProgressBar creates a simple inline progress bar for single operations
func SimpleProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(percent * float64(width))
	empty := width - filled

	filledPart := strings.Repeat("█", filled)
	emptyPart := strings.Repeat("░", empty)

	// Create a style with primary color
	barStyle := lipgloss.NewStyle().Foreground(primaryColor)

	return fmt.Sprintf("%s%s %s %s",
		barStyle.Render(filledPart),
		mutedStyle.Render(emptyPart),
		infoStyle.Render(fmt.Sprintf("%3.0f%%", percent*100)),
		mutedStyle.Render(fmt.Sprintf("(%d/%d)", current, total)),
	)
}

// SpinnerFrames provides spinner animation frames
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner manages a simple spinner animation
type Spinner struct {
	frames  []string
	current int
	mu      sync.Mutex
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	return &Spinner{
		frames: SpinnerFrames,
	}
}

// Next returns the next spinner frame
func (s *Spinner) Next() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	frame := s.frames[s.current]
	s.current = (s.current + 1) % len(s.frames)
	spinnerStyle := lipgloss.NewStyle().Foreground(primaryColor)
	return spinnerStyle.Render(frame)
}
