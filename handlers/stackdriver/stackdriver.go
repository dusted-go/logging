package stackdriver

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/dusted-go/logging/v2/slogctx"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type HandlerOptions struct {
	ServiceName    string
	ServiceVersion string
	MinLevel       slog.Leveler
	AddSource      bool
}

type MiddlewareOptions struct {
	GCPProjectID   string
	AddTrace       bool
	AddHTTPRequest bool
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

func Logging(
	hOpts *HandlerOptions,
	mOpts *MiddlewareOptions,
) func(http.Handler) http.Handler {
	handler := NewHandler(hOpts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				requestID := r.Header.Get("X-Request-ID")
				if requestID == "" {
					requestID = uuid.NewString()
				}
				reqHandler := handler.WithAttrs(
					[]slog.Attr{slog.String("requestId", requestID)})

				if mOpts.AddTrace {
					span := trace.SpanFromContext(ctx).SpanContext()
					if span.IsValid() {
						traceID, spanID, sampled := getTraceAttrs(mOpts.GCPProjectID, span)
						reqHandler = reqHandler.WithAttrs([]slog.Attr{traceID, spanID, sampled})
					}
				}
				if mOpts.AddHTTPRequest {
					reqHandler = reqHandler.WithAttrs([]slog.Attr{
						slog.Group("httpRequest",
							slog.String("requestMethod", r.Method),
							slog.String("requestUrl", r.URL.String()),
							slog.String("protocol", r.Proto),
							slog.String("remoteIp", r.RemoteAddr),
							slog.String("userAgent", r.UserAgent()),
							slog.String("referer", r.Referer()),
						),
					})
				}
				logger := slog.New(reqHandler)
				ctx = slogctx.WithLogger(ctx, logger)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			},
		)
	}
}
