package stackdriver

import (
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

type HandlerOptions struct {
	ServiceName    string
	ServiceVersion string
	MinLevel       slog.Leveler
	AddSource      bool
}

// Official Google Cloud Logging docs for structured logs:
// - https://cloud.google.com/run/docs/logging#writing_structured_logs
// Documentation on JSON payloads and special fields:
// - https://cloud.google.com/logging/docs/agent/logging/configuration#process-payload
func stackdriverAttrs(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.MessageKey {
		a.Key = "message"
		return a
	}
	if a.Key == slog.SourceKey {
		a.Key = "logging.googleapis.com/sourceLocation"
		return a
	}
	if err, ok := a.Value.Any().(error); ok {
		return slog.Group("error",
			slog.String("message", err.Error()),
			slog.Any("stack", CaptureStack().Slice()),
		)
	}
	return ReplaceLogLevel(groups, a)
}

func NewHandler(opts *HandlerOptions) *Handler {
	handlerOpts := &slog.HandlerOptions{
		Level:       opts.MinLevel,
		AddSource:   opts.AddSource,
		ReplaceAttr: stackdriverAttrs,
	}
	handler := slog.
		NewJSONHandler(os.Stdout, handlerOpts).
		WithAttrs([]slog.Attr{
			slog.Group("serviceContext",
				slog.String("service", opts.ServiceName),
				slog.String("version", opts.ServiceVersion),
			),
		})
	return &Handler{h: handler}
}

func getTraceAttrs(googleProjectID string, span trace.SpanContext) (slog.Attr, slog.Attr, slog.Attr) {
	googleTraceID := fmt.Sprintf(
		"projects/%s/traces/%s",
		googleProjectID,
		span.TraceID().String())
	return slog.String("logging.googleapis.com/trace", googleTraceID),
		slog.String("logging.googleapis.com/spanId", span.SpanID().String()),
		slog.Bool("logging.googleapis.com/trace_sampled", span.IsSampled())
}

func WithTrace(
	logger *slog.Logger,
	span trace.Span,
	googleProjectID string,
) *slog.Logger {
	spanCtx := span.SpanContext()
	if spanCtx.IsValid() {
		traceID, spanID, sampled := getTraceAttrs(googleProjectID, spanCtx)
		return logger.With(traceID, spanID, sampled)
	}
	return logger
}
