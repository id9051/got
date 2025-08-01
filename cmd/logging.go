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
)

// Constants for commonly used logging messages
const (
	ErrorWalkingMessage = "error walking filepath [%s]"
	SkippingMessage     = "Skipping [%s]"
	SuccessMessage      = "[%s]:  Success"
	ErrorMessage        = "[%s]: ERROR %v"
)

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