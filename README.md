# Logging

A collection of `log/slog` integrations.

## prettylog

Prettylog is a `log/slog` handler for pretty console output, designed to be used during development time only.

Example:

```
prettyHandler = prettylog.NewHandler(&slog.HandlerOptions{
    Level:       slog.LevelInfo,
    AddSource:   false,
    ReplaceAttr: nil,
})
logger := slog.New(prettyHandler)
```

## stackdriver

The `stackdriver.Handler` is a `log/slog` handler which will output log messages in the Google Cloud Logging format. It has support for traces, a http middleware which can be used to issue request scoped loggers (with attached request information) and has a `WithLogger` and `GetLogger` helper function to store/retrieve a logger from a `context.Context` object (this is used inside the middleware).