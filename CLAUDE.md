# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"Got" is a CLI tool for managing multiple Git repositories built with Go using the Cobra CLI framework and Viper for configuration. It allows users to perform git operations (pull, fetch, status) across single repositories or recursively across directory trees containing multiple git repositories.

## Architecture

### Core Structure
- **Entry Point**: `main.go` - Simple entry point that calls `cmd.Execute()`
- **Command Layer**: `cmd/` directory contains all CLI commands
  - `root.go` - Root command setup, configuration initialization, and global flags
  - `pull.go` - Git pull operations with recursive directory walking
  - `fetch.go` - Git fetch operations with recursive directory walking  
  - `status.go` - Git status operations with recursive directory walking

### Key Design Patterns
- **Cobra Commands**: Each git operation is implemented as a separate Cobra command
- **Shared Recursive Logic**: All commands support `-r/--recursive` flag for directory tree operations
- **Configuration Management**: Uses Viper to load `.got.yaml` config files with skip list functionality
- **Directory Walking**: Uses `filepath.Walk` for recursive operations with configurable skip patterns
- **Error Handling**: Extensive use of `github.com/pkg/errors` for error wrapping and context

### Configuration System
- Config file: `.got.yaml` (default location: `$HOME/.got.yaml`)
- Override with `--config` flag
- Key configuration: `skipList` array for directories to skip during recursive operations
- Default skip list is now empty (configurable via `getSkipList()` in `cmd/pull.go:34`)

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
```bash
# Run tests (if any exist)
go test ./...

# Run tests with verbose output
go test -v ./...
```

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

### Recursive Operations
- The `recursive` variable in `cmd/pull.go:31` is shared across commands
- Recursive logic uses `filepath.Walk` with skip directory functionality
- Skip patterns are checked using `strings.Contains()` for flexible matching

### Git Operations
- All git commands use `exec.Command` with explicit `--work-tree` and `--git-dir` flags
- Error handling logs failures but continues processing in recursive mode
- Non-git directories are silently skipped in recursive mode

### Error Recovery
- `pullWalk()` in `cmd/pull.go:111` includes special handling for directories that are deleted during processing
- Implements defensive programming against race conditions during recursive operations