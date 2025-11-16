package httplogger

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dusted-go/logging/v2/slogctx"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type HandlerFactory func() slog.Handler

func RequestScoped(
	baseHandler slog.Handler,
	addTrace bool,
	logRequest bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				// Always parse an existing X-Request-ID header or generate a new one.
				// More info: https://http.dev/x-request-id
				requestID := r.Header.Get("X-Request-ID")
				if requestID == "" {
					requestID = uuid.NewString()
				}

				// Create a request-scoped handler with request ID.
				reqHandler := baseHandler.WithAttrs(
					[]slog.Attr{slog.String("request.id", requestID)})

				// Add trace IDs if requested and available.
				if addTrace {
					span := trace.SpanFromContext(ctx).SpanContext()
					if span.IsValid() {
						reqHandler = reqHandler.WithAttrs([]slog.Attr{
							slog.String("trace_id", span.TraceID().String()),
							slog.String("span_id", span.SpanID().String()),
						})
					}
				}

				// Create a request-scoped logger and add it to the request context.
				logger := slog.New(reqHandler)
				ctx = slogctx.WithLogger(ctx, logger)
				r = r.WithContext(ctx)

				// Optionally log HTTP request metadata.
				if logRequest {
					logger.Info("Processing HTTP request",
						// See: https://opentelemetry.io/docs/specs/semconv/http/http-spans/
						slog.String("http.request.method", r.Method),
						slog.String("http.scheme", r.URL.Scheme),
						slog.String("http.flavor", fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor)),
						slog.String("http.url", r.URL.String()),
						slog.String("http.target", r.URL.RawPath+"?"+r.URL.RawQuery),
						slog.String("http.route", r.Pattern),
						slog.String("http.user_agent", r.UserAgent()),
						slog.String("http.referer", r.Referer()),
						slog.Int64("http.request.size", r.ContentLength),
						slog.String("client.address", r.RemoteAddr),
					)
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
