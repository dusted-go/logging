package httplogger

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/dusted-go/logging/v2/slogctx"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type HandlerFactory func() slog.Handler

func RequestScoped(
	baseHandler slog.Handler,
	addTrace bool,
	logRequest bool,
	excludeHeaders []string,
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
					// Log according to OpenTelemetry HTTP Server Semantic Conventions:
					// https://opentelemetry.io/docs/specs/semconv/http/http-spans/#http-server
					host, port := firstHostPort(
						r.Header.Get("X-Forwarded-Host"),
						r.Host,
					)

					scheme := r.Header.Get("X-Forwarded-Proto")
					if scheme == "" && r.TLS != nil {
						scheme = "https"
					} else if scheme == "" {
						scheme = "http"
					}

					if port == -1 {
						switch scheme {
						case "https":
							port = 443
						case "http":
							port = 80
						}
					}

					clientAddress, clientPort := firstHostPort(
						r.Header.Get("X-Real-IP"),
						parseXForwardedFor(r.Header.Get("X-Forwarded-For")),
						r.RemoteAddr,
					)

					networkProtocol := "http"
					networkProtocolVersion := fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor)
					protoName, protoVersion, ok := strings.Cut(r.Proto, "/")
					if ok {
						networkProtocol = strings.ToLower(protoName)
						networkProtocolVersion = protoVersion
					}

					urlQuery := r.URL.Query().Encode()

					requestAttrs := []any{
						slog.String("server.address", host),
						slog.Int("server.port", port),
						slog.String("network.protocol.name", networkProtocol),
						slog.String("network.protocol.version", networkProtocolVersion),
						slog.String("http.request.method", r.Method),
						slog.Int64("http.request.size", r.ContentLength),
						slog.String("url.path", r.URL.Path),
						slog.String("url.scheme", scheme),
						slog.String("user_agent.original", r.UserAgent()),
						slog.String("client.address", clientAddress),
					}

					if clientPort != -1 {
						requestAttrs = append(
							requestAttrs,
							slog.Int("client.port", clientPort),
						)
					}

					if urlQuery != "" {
						requestAttrs = append(
							requestAttrs,
							slog.String("url.query", "?"+urlQuery),
						)
					}

					peer, peerPort := splitHostPort(r.RemoteAddr)
					if peer != "" {
						requestAttrs = append(
							requestAttrs,
							slog.String("network.peer.address", peer),
						)

						if peerPort != -1 {
							requestAttrs = append(
								requestAttrs,
								slog.Int("network.peer.port", peerPort),
							)
						}
					}

					logger.Info(
						"Processing HTTP request",
						requestAttrs...)
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}

// splitHostPort splits a network address hostport of the form "host",
// "host%zone", "[host]", "[host%zone], "host:port", "host%zone:port",
// "[host]:port", "[host%zone]:port", or ":port" into host or host%zone and
// port.
//
// An empty host is returned if it is not provided or unparsable. A negative
// port is returned if it is not provided or unparsable.
func splitHostPort(hostport string) (host string, port int) {
	port = -1

	if strings.HasPrefix(hostport, "[") {
		addrEnd := strings.LastIndex(hostport, "]")
		if addrEnd < 0 {
			// Invalid hostport.
			return host, port
		}
		if i := strings.LastIndex(hostport[addrEnd:], ":"); i < 0 {
			host = hostport[1:addrEnd]
			return host, port
		}
	} else {
		if i := strings.LastIndex(hostport, ":"); i < 0 {
			host = hostport
			return host, port
		}
	}

	host, pStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return host, port
	}

	p, err := strconv.ParseUint(pStr, 10, 16)
	if err != nil {
		return host, port
	}
	return host, int(p) // nolint: gosec  // Bit size of 16 checked above.
}

// Return the request host and port from the first non-empty source.
func firstHostPort(source ...string) (host string, port int) {
	for _, hostport := range source {
		host, port = splitHostPort(hostport)
		if host != "" || port > 0 {
			break
		}
	}
	return host, port
}

// parseXForwardedFor extracts the first IP address from a X-Forwarded-For header value.
func parseXForwardedFor(xForwardedFor string) string {
	if idx := strings.Index(xForwardedFor, ","); idx >= 0 {
		xForwardedFor = xForwardedFor[:idx]
	}
	return xForwardedFor
}
