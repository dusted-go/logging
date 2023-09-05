package stackdriver

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

type MiddlewareOptions struct {
	GCPProjectID   string
	AddTrace       bool
	AddHTTPRequest bool
}

func Logging(
	hOpts *HandlerOptions,
	mOpts *MiddlewareOptions,
) func(http.Handler) http.Handler {
	var handler slog.Handler
	handler = NewHandler(hOpts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				if mOpts.AddTrace {
					span := trace.SpanFromContext(ctx).SpanContext()
					if span.IsValid() {
						traceID, spanID, sampled := getTraceAttrs(mOpts.GCPProjectID, span)
						handler = handler.WithAttrs([]slog.Attr{traceID, spanID, sampled})
					}
				}
				if mOpts.AddHTTPRequest {
					handler = handler.WithAttrs([]slog.Attr{
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
				logger := slog.New(handler)
				ctx = WithLogger(ctx, logger)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			},
		)
	}
}
