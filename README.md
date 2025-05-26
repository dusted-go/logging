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

## Performance

This handler is designed for development use. Run `go test -bench=.` to benchmark on your system.

```
goos: darwin
goarch: arm64
pkg: github.com/mikluko/slogging
cpu: Apple M4 Pro
BenchmarkHandlers/StandardJSONHandler-14         	 2933217	       395.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkHandlers/StandardTextHandler-14         	 2657721	       440.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkHandlers/SloggingHandler-14             	  668452	      1784 ns/op	    1420 B/op	      37 allocs/op
BenchmarkHandlers/SloggingHandlerWithColor-14    	  554329	      2124 ns/op	    1798 B/op	      49 allocs/op
BenchmarkHandlers/SloggingHandlerYAML-14         	  339201	      3351 ns/op	    8091 B/op	      63 allocs/op
```

**Key takeaways:**
- **~4-5x slower** than standard handlers due to JSON parsing and formatting
- **Significant memory allocations** for pretty formatting
- **YAML encoding** is considerably more expensive than JSON
- **Colorization** adds ~15% overhead

**Use this handler for development/debugging only.** For production, use standard `slog.JSONHandler` or `slog.TextHandler`.