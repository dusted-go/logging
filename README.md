# Slogging

A `log/slog` handler for pretty console output, designed to be used during development time only.

## Prior Work

This repository is a fork of [dusted-go/logging](https://github.com/dusted-go/logging), focusing
exclusively on the prettylog handler for development use. The original repository included
additional handlers for production logging (stackdriver) which have been removed in this fork to
maintain a single, focused purpose.

## Example

```go
import (
"log/slog"
"github.com/mikluko/slogging"
)

prettyHandler := slogging.NewHandler(&slog.HandlerOptions{
Level:       slog.LevelInfo,
AddSource:   false,
ReplaceAttr: nil,
})
logger := slog.New(prettyHandler)
```