Release Notes
=============

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