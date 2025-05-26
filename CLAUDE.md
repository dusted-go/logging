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

When instructed to release, follow these steps:

1. **Determine the version number using EffVer (Effort-based Versioning)**:
   - Format: MACRO.MESO.MICRO (e.g., 1.0.0, 1.0.1, 1.1.0, 2.0.0)
   - Review commits since last tag to assess upgrade effort:
     - **MICRO**: Minimal effort - bug fixes, docs, minor improvements that don't affect usage
     - **MESO**: Moderate effort - new features, changes that might require small adjustments
     - **MACRO**: Significant effort - breaking changes, major restructuring, API changes

2. **Generate changelog from git history**:
   ```bash
   # Get the last tag
   git describe --tags --abbrev=0
   
   # Generate changelog since last tag
   git log <last-tag>..HEAD --pretty=format:"- %s" --reverse
   ```

3. **Create and push the tag**:
   ```bash
   git tag v<VERSION>
   git push origin v<VERSION>
   ```

4. **Create GitHub release**:
   - Use the generated changelog for release notes
   - Tag version: v<VERSION>
   - Release title: v<VERSION>
   - Include upgrade effort note if MESO or MACRO bump

Example: If the last version was v1.2.5:
- Bug fixes only → v1.2.6
- New optional features → v1.3.0  
- Breaking API changes → v2.0.0
