Release Notes
=============

# 2.0.0-alpha-01

- Separated the log context logic from the `stackdriver` package.
- Added a new middleware `httplogger` which decorates the `slog.Handler` with OTel attributes for traces and the request ID and optionally emits a request log.

# 1.3.0

- Added a new option called `WithOutputEmptyAttrs` to enable the writing of empty `Attrs`. This is enabled by default for backwards compatibility.

# 1.2.2

- Fixed bug where calling `.With` on an existing prettylog logger would panic (see: https://github.com/dusted-go/logging/issues/6)

# 1.2.1

- Append newline to log by @jsumners in https://github.com/dusted-go/logging/pull/5

# 1.2.0

- `prettylog` has a new `New` function which lets one specify additional options to provide an `io.Writer` and/or disable colouring

# 1.1.3

- Fixed bug in `stackdriver.Middleware` which meant that a single log handler was shared across the entire lifespan of the application instead of creating request scoped log handlers

# 1.1.2

- Added `nil` checks around the `ReplaceAttr` function to prevent panics (see: #2)

## 1.1.1

- Fixed panic when logging after WithGroup or WithAttrs using prettylog (see: #1)

## 1.1.0

- Removed logging folder

## 1.0.0

- **prettylog**: Pretty console log handler
- **stackdriver**: Google Cloud Logging handler