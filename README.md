# Got - Git Repository Management Tool

A CLI tool for managing multiple Git repositories using Cobra and Viper.

## Configuration

The tool supports configuration via a YAML file. By default, it looks for `.got.yaml` in your home directory, but you can specify a custom config file using the `--config` flag.

### Skip List Configuration

You can configure which directories to skip during recursive operations by setting the `skipList` in your configuration file:

```yaml
# .got.yaml
skipList:
  - "princetontmx.com/mobile/tmx-shipper-app"
  - "go/src/github.com/ardanlabs/service-training"
  - "vendor"
  - "node_modules"
  - "tmp"
```

If no configuration file is found or the `skipList` is not specified, the tool will use these default skip patterns:
- `princetontmx.com/mobile/tmx-shipper-app`
- `go/src/github.com/ardanlabs/service-training`

## Usage

### Commands

- `got pull [directory]` - Pull changes in a Git repository
- `got fetch [directory]` - Fetch changes in a Git repository  
- `got status [directory]` - Check status of a Git repository

### Flags

- `-r, --recursive` - Recursively operate on subdirectories
- `--config [file]` - Specify custom configuration file

### Examples

```bash
# Pull changes in current directory
got pull .

# Recursively pull all git repositories in a directory
got pull -r /path/to/projects

# Use custom config file
got --config /path/to/custom.yaml pull -r /path/to/projects
```
