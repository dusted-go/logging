package httplogger

import (
	"crypto/tls"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestAttributes(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		tls            bool
		excludeHeaders []string
		want           map[string]any
	}{
		{
			name:       "Basic Request",
			headers:    map[string]string{"User-Agent": "Go-Test"},
			remoteAddr: "127.0.0.1:12345",
			want: map[string]any{
				"network.protocol.name":    "http",
				"network.protocol.version": "1.1",
				"http.request.method":      "GET",
				"url.path":                 "/",
				"url.scheme":               "http",
				"user_agent.original":      "Go-Test",
				"client.address":           "127.0.0.1",
				"client.port":              int64(12345),
				"server.port":              int64(80),
			},
		},
		{
			name: "HTTPS Request",
			headers: map[string]string{
				"User-Agent": "Go-Test",
			},
			remoteAddr: "127.0.0.1:12345",
			tls:        true,
			want: map[string]any{
				"url.scheme":  "https",
				"server.port": int64(443),
			},
		},
		{
			name: "X-Forwarded Headers",
			headers: map[string]string{
				"X-Forwarded-For":   "10.0.0.1, 10.0.0.2",
				"X-Forwarded-Host":  "example.com:8080",
				"X-Forwarded-Proto": "https",
			},
			remoteAddr: "127.0.0.1:12345",
			want: map[string]any{
				"client.address": "10.0.0.1",
				"server.address": "example.com",
				"server.port":    int64(8080),
				"url.scheme":     "https",
			},
		},
		{
			name: "Forwarded Header (RFC 7239)",
			headers: map[string]string{
				"Forwarded": "for=192.0.2.60;proto=http;by=203.0.113.43, for=198.51.100.17",
			},
			remoteAddr: "127.0.0.1:12345",
			want: map[string]any{
				"client.address": "192.0.2.60",
				"url.scheme":     "http",
				"server.port":    int64(80),
			},
		},
		{
			name: "Forwarded Header with Host and Port",
			headers: map[string]string{
				"Forwarded": "host=api.example.com:8443;proto=https",
			},
			remoteAddr: "127.0.0.1:12345",
			want: map[string]any{
				"server.address": "api.example.com",
				"server.port":    int64(8443),
				"url.scheme":     "https",
			},
		},
		{
			name: "Excluded Headers",
			headers: map[string]string{
				"Authorization": "Bearer secret",
				"X-API-Key":     "12345",
				"User-Agent":    "Go-Test",
			},
			excludeHeaders: []string{"Authorization", "x-api-key"},
			want: map[string]any{
				"http.request.header.user-agent": "Go-Test",
			},
			// We need to check that excluded headers are NOT present
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remoteAddr
			if tt.tls {
				req.TLS = &tls.ConnectionState{} // Dummy TLS state
			}

			// We need to mock TLS if needed, but httptest.NewRequest doesn't set it easily for non-TLS
			// For the HTTPS test case, we can manually set req.TLS
			if tt.tls {
				// Just setting it to non-nil is enough for the logic
				req.TLS = &tls.ConnectionState{} // Use correct type
			}

			attrs := requestAttributes(req, tt.excludeHeaders)
			attrMap := make(map[string]any)
			for _, a := range attrs {
				if kv, ok := a.(slog.Attr); ok {
					attrMap[kv.Key] = kv.Value.Any()
				}
			}

			for k, v := range tt.want {
				assert.Equal(t, v, attrMap[k], "Attribute %s mismatch", k)
			}

			// Check excluded headers
			for _, h := range tt.excludeHeaders {
				key := "http.request.header." + strings.ToLower(h)
				_, exists := attrMap[key]
				assert.False(t, exists, "Header %s should be excluded", h)
			}
		})
	}
}

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		input    string
		wantHost string
		wantPort int
	}{
		{"example.com:80", "example.com", 80},
		{"example.com", "example.com", -1},
		{"127.0.0.1:8080", "127.0.0.1", 8080},
		{"[::1]:80", "::1", 80},
		{":80", "", 80},
		{"invalid", "invalid", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			host, port := splitHostPort(tt.input)
			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantPort, port)
		})
	}
}
