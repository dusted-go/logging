# Logging

A collection of `log/slog` integrations.

## prettylog

Prettylog is a `log/slog` handler for pretty console output, designed to be used during development time only.

Example:

```go
package demo

import (
	"log/slog"
	
	"github.com/dusted-go/logging/prettylog"
)

func demo() {
	prettyHandler := prettylog.NewHandler(&slog.HandlerOptions{
		Level:       slog.LevelInfo,
		AddSource:   false,
		ReplaceAttr: nil,
	})
	logger := slog.New(prettyHandler)
}
```

## context

The library provides storing and retrieving a logger from a context. If no logger is found in the context, then `slog.Default()` is returned.

```go
package demo

import (
	"context"
	"log/slog"

	slogContext "github.com/dusted-go/logging/prettylog/context"
)

func demo() {
	ctx := context.Background()

	// add logger to context
	newCtx := slogContext.WithLogger(ctx, slog.Default())
	
	// getting logger from context
	logger := slogContext.GetLogger(newCtx)
}
```

## stackdriver

The `stackdriver.Handler` is a `log/slog` handler which will output log messages in the Google Cloud Logging format. It has support for traces, a http middleware which can be used to issue request scoped loggers (with attached request information) and has a `WithLogger` and `GetLogger` helper function to store/retrieve a logger from a `context.Context` object (this is used inside the middleware).
