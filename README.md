# Logging

A collection of `log/slog` integrations and helpers.

## Modules

### handlers/prettylog

`prettylog` is a `log/slog` handler for pretty console output, designed to be used during development time only. It supports colorized output and human-readable formatting.

**Example:**

```go
import (
	"log/slog"
	"os"

	"github.com/dusted-go/logging/v2/handlers/prettylog"
)

func main() {
	// Create a pretty handler
	handler := prettylog.NewHandler(&slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	logger.Info("Application started", "env", "dev")
}
```

### handlers/stackdriver

The `stackdriver` package provides a `log/slog` handler that outputs logs in the Google Cloud Logging structured format. It also includes an HTTP middleware for request logging and trace integration.

**Handler Usage:**

```go
import (
	"log/slog"

	"github.com/dusted-go/logging/v2/handlers/stackdriver"
)

func main() {
	opts := &stackdriver.HandlerOptions{
		ServiceName:    "my-service",
		ServiceVersion: "1.0.0",
		MinLevel:       slog.LevelInfo,
	}

	handler := stackdriver.NewHandler(opts)
	logger := slog.New(handler)
}
```

**Middleware Usage:**

The middleware automatically adds a request-scoped logger to the context, handles trace ID propagation, and logs HTTP request details.

```go
import (
	"net/http"

	"github.com/dusted-go/logging/v2/handlers/stackdriver"
)

func main() {
	hOpts := &stackdriver.HandlerOptions{
		ServiceName: "my-service",
	}
	mOpts := &stackdriver.MiddlewareOptions{
		GCPProjectID:   "my-project-id",
		AddTrace:       true,
		AddHTTPRequest: true,
	}

	// Create the middleware
	mw := stackdriver.Logging(hOpts, mOpts)

	http.Handle("/", mw(myHandler))
}
```

### middlewares/httplogger

`httplogger` provides a generic HTTP middleware that creates a request-scoped logger with attributes based on [OpenTelemetry HTTP Server Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#http-server).

It handles:
- `X-Request-ID` generation/propagation
- Trace context extraction
- Request attribute extraction (method, path, user agent, etc.)
- `Forwarded` header parsing (RFC 7239)

**Example:**

```go
import (
	"net/http"
	"log/slog"

	"github.com/dusted-go/logging/v2/middlewares/httplogger"
	"github.com/dusted-go/logging/v2/slogctx"
)

func main() {
	cfg := httplogger.Config{
		LogRequest: true,
		// BaseHandler: jsonHandler, // Optional: use a specific handler
	}

	mw := httplogger.RequestScoped(cfg)

	http.Handle("/", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the request-scoped logger
		logger := slogctx.GetLogger(r.Context())
		logger.Info("Handling request")
	})))
}
```

### slogctx

`slogctx` is a helper package for storing and retrieving `*slog.Logger` from `context.Context`. It is used by the middlewares to propagate the request-scoped logger.

**Example:**

```go
import (
	"context"
	"log/slog"

	"github.com/dusted-go/logging/v2/slogctx"
)

func main() {
	ctx := context.Background()
	
	// Add logger to context
	ctx = slogctx.WithLogger(ctx, slog.Default())

	// Retrieve logger from context (returns slog.Default() if not found)
	logger := slogctx.GetLogger(ctx)
	logger.Info("Log from context")
}
```