# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"Got" is a CLI tool for managing multiple Git repositories built with Go using the Cobra CLI framework and Viper for configuration. It allows users to perform git operations (pull, fetch, status) across single repositories or recursively across directory trees containing multiple git repositories.

## Architecture

### Core Structure
- **Entry Point**: `main.go` - Simple entry point that calls `cmd.Execute()`
- **Command Layer**: `cmd/` directory contains all CLI commands
  - `root.go` - Root command setup, configuration initialization, shared utilities, and global flags
  - `utils.go` - Shared utilities for validation, git operations, logging, and directory walking
  - `pull.go` - Git pull command implementation (simplified, uses shared utilities)
  - `fetch.go` - Git fetch command implementation (simplified, uses shared utilities)  
  - `status.go` - Git status command implementation (simplified, uses shared utilities)

### Key Design Patterns
- **Shared Utilities Architecture**: Common functionality centralized in `utils.go`
- **Cobra Commands**: Each git operation is implemented as a separate Cobra command
- **Generic Directory Walker**: Single `walkDirectories()` function handles all recursive operations
- **Dual Operation Modes**: Separate functions for single directory vs recursive operations
- **Input Validation**: Comprehensive directory path validation with clear error messages
- **Configuration Management**: Uses Viper to load `.got.yaml` config files with skip list validation
- **Error Handling**: Extensive use of `github.com/pkg/errors` for error wrapping and context

### Configuration System
- Config file: `.got.yaml` (default location: `$HOME/.got.yaml`)
- Override with `--config` flag
- Key configuration: `skipList` array for directories to skip during recursive operations
- Configuration validation: Empty and whitespace-only entries are automatically filtered
- Default skip list is now empty (configurable via `getSkipList()` in `cmd/root.go:28`)

## Development Commands

### Building and Running
```bash
# Build the application
go build -o got .

# Run directly with go
go run . [command] [flags] [args]

# Install locally
go install .
```

### Testing
The project has comprehensive test coverage (78.7%) including unit tests, integration tests, and test utilities.

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific test package
go test ./cmd

# Run tests with verbose output
go test ./cmd -v

# Generate detailed coverage report
go test ./cmd -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test ./cmd -run TestValidateDirectoryPath

# Run integration tests only
go test ./cmd -run Integration
```

#### Test Structure
- **Unit Tests**: `cmd/*_test.go` - Test individual functions and components
- **Integration Tests**: `cmd/integration_test.go` - Test component interactions  
- **Test Utilities**: `testutil/helpers.go` - Reusable test infrastructure

#### Test Categories
1. **Utility Function Tests** (`utils_test.go`):
   - Directory validation (`validateDirectoryPath`)
   - Git repository detection (`isGitRepository`) 
   - Skip list functionality (`shouldSkipPath`)
   - Command execution logic
   - Directory walking operations

2. **Configuration Tests** (`root_test.go`):
   - Viper configuration loading
   - Skip list parsing and validation
   - Command structure and flags
   - Shell completion setup

3. **Command Tests** (`pull_test.go`, `fetch_test.go`, `status_test.go`):
   - Argument validation
   - Flag handling (recursive, etc.)
   - Error message verification
   - Command-specific functionality

4. **Integration Tests** (`integration_test.go`):
   - End-to-end workflow testing
   - Complex directory structures
   - Skip list behavior verification
   - Configuration file interactions

#### Test Utilities
The `testutil` package provides:
- `CreateTempGitRepo()` - Creates temporary git repositories
- `CreateTestDirStructure()` - Creates complex directory hierarchies
- `CreateConfigFile()` - Generates test configuration files
- Helper functions for assertions and validation

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run Go modules maintenance
go mod tidy
go mod verify
```

## Usage Patterns

### Command Structure
All commands follow the pattern: `got [command] [directory] [flags]`

### Dual Operation Modes
- **Single Directory Mode**: `executeGitCommandSingle()` - Requires target to be a git repository
- **Recursive Mode**: `executeGitCommand()` - Silently skips non-git directories during tree walking
- Input validation ensures directory exists and is accessible before operations begin

### Recursive Operations  
- Generic `walkDirectories()` function in `cmd/utils.go:150` handles all recursive operations
- Progress indicators show scanning progress and completion summary
- Skip patterns are checked using `strings.Contains()` for flexible matching
- Robust error handling continues processing even when individual operations fail

### Git Operations
- All git commands use `runGitCommand()` with explicit `--work-tree` and `--git-dir` flags
- Status command output is directed to stdout/stderr for user visibility
- Error handling logs failures but continues processing other repositories
- Constants defined for all magic strings to prevent typos

### Error Recovery and Validation
- Comprehensive input validation with clear, user-friendly error messages
- Race condition handling for directories deleted during recursive operations
- Configuration validation automatically filters invalid skip list entries
- Defensive programming throughout with proper error wrapping and context