# Slogging

A `log/slog` handler for pretty console output, designed to be used during development time only.

Example:

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