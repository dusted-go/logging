# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

```bash
# Build the project
go build ./...

# Run tests
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run ./...

# Complete build pipeline (as defined in build.sh)
go mod tidy && go build ./... && go test ./... && go fmt ./... && golangci-lint run ./...
```

## Architecture Overview

This is a Go module providing a `log/slog` handler for development-time logging.

### slogging Package
- A `slog.Handler` implementation for human-readable console output during development
- Supports JSON and YAML encoding formats via the `encoder` type
- Uses ANSI color codes for level-based colorization
- Thread-safe implementation using mutex for buffer operations
- Core type is `Handler` struct that wraps an inner `slog.JSONHandler`

## Key Design Patterns

The handler follows the decorator pattern - it wraps the standard `slog.JSONHandler` and transforms its output rather than implementing logging from scratch.

## Releasing

- 
