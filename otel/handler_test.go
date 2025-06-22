package otel

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func Test_OtelHandler(t *testing.T) {
	t.Run("logs with valid span context should include trace_id and span_id", func(t *testing.T) {
		// Create a tracer provider and tracer
		tp := sdktrace.NewTracerProvider()
		tracer := tp.Tracer("test-tracer")

		// Create capture stream and base handler
		buf := new(bytes.Buffer)
		baseHandler := slog.NewTextHandler(buf, nil)

		// Wrap with OtelHandler
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		// Start a span and log within its context
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// Log with the span context
		logger.InfoContext(ctx, "test message with tracing")

		// Get the output (trim trailing newline for splitting)
		output := strings.TrimSuffix(buf.String(), "\n")
		lines := strings.Split(output, "\n")

		// Verify log output
		if len(lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(lines))
		}

		// Extract span context to get expected values
		spanContext := span.SpanContext()
		expectedTraceID := spanContext.TraceID().String()
		expectedSpanID := spanContext.SpanID().String()

		// Check that otel.trace_id and otel.span_id are present in the output
		if !strings.Contains(lines[0], `otel.trace_id=`+expectedTraceID) {
			t.Errorf("expected otel.trace_id=%s in output, got: %s", expectedTraceID, lines[0])
		}
		if !strings.Contains(lines[0], `otel.span_id=`+expectedSpanID) {
			t.Errorf("expected otel.span_id=%s in output, got: %s", expectedSpanID, lines[0])
		}
	})

	t.Run("logs without span context should not include trace_id and span_id", func(t *testing.T) {
		// Create capture stream and base handler
		buf := new(bytes.Buffer)
		baseHandler := slog.NewTextHandler(buf, nil)

		// Wrap with OtelHandler
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		// Log without span context
		logger.Info("test message without tracing")

		// Get the output (trim trailing newline for splitting)
		output := strings.TrimSuffix(buf.String(), "\n")
		lines := strings.Split(output, "\n")

		// Verify log output
		if len(lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(lines))
		}

		// Check that trace_id and span_id are NOT present
		if strings.Contains(lines[0], `otel.trace_id=`) {
			t.Errorf("unexpected otel.trace_id in output: %s", lines[0])
		}
		if strings.Contains(lines[0], `otel.span_id=`) {
			t.Errorf("unexpected otel.span_id in output: %s", lines[0])
		}
	})

	t.Run("logs with invalid span context should not include trace_id and span_id", func(t *testing.T) {
		// Create capture stream and base handler
		buf := new(bytes.Buffer)
		baseHandler := slog.NewTextHandler(buf, nil)

		// Wrap with OtelHandler
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		// Create a context with an invalid span (no-op span)
		ctx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(context.Background()))

		// Log with invalid span context
		logger.InfoContext(ctx, "test message with invalid span")

		// Get the output (trim trailing newline for splitting)
		output := strings.TrimSuffix(buf.String(), "\n")
		lines := strings.Split(output, "\n")

		// Verify log output
		if len(lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(lines))
		}

		// Check that trace_id and span_id are NOT present
		if strings.Contains(lines[0], `otel.trace_id=`) {
			t.Errorf("unexpected otel.trace_id in output: %s", lines[0])
		}
		if strings.Contains(lines[0], `otel.span_id=`) {
			t.Errorf("unexpected otel.span_id in output: %s", lines[0])
		}
	})

	t.Run("trace attributes should be at root level with groups", func(t *testing.T) {
		// Create a tracer provider and tracer
		tp := sdktrace.NewTracerProvider()
		tracer := tp.Tracer("test-tracer")

		// Create capture stream and base handler
		buf := new(bytes.Buffer)
		baseHandler := slog.NewTextHandler(buf, nil)

		// Wrap with OtelHandler
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		// Create logger with group
		groupedLogger := logger.WithGroup("mygroup")

		// Start a span and log within its context
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// Log with the span context using grouped logger
		groupedLogger.InfoContext(ctx, "test message", "key", "value")

		// Get the output (trim trailing newline for splitting)
		output := strings.TrimSuffix(buf.String(), "\n")
		lines := strings.Split(output, "\n")

		// Verify log output
		if len(lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(lines))
		}

		// Verify trace_id and span_id are present at root level (otel.trace_id format)
		if !strings.Contains(lines[0], `otel.trace_id=`) {
			t.Errorf("otel.trace_id missing from output: %s", lines[0])
		}
		if !strings.Contains(lines[0], `otel.span_id=`) {
			t.Errorf("otel.span_id missing from output: %s", lines[0])
		}

		// Verify the grouped attribute is present with dot notation
		if !strings.Contains(lines[0], `mygroup.key=value`) {
			t.Errorf("mygroup.key=value missing from output: %s", lines[0])
		}

		// Verify that trace attributes are at root level (not inside mygroup)
		// This is confirmed by the fact that they use otel.trace_id format, not mygroup.otel.trace_id
		// If they were inside mygroup, they would appear as mygroup.otel.trace_id
		if strings.Contains(lines[0], `mygroup.otel.trace_id=`) {
			t.Errorf("trace_id should NOT be inside mygroup, but found mygroup.otel.trace_id in output: %s", lines[0])
		}
		if strings.Contains(lines[0], `mygroup.otel.span_id=`) {
			t.Errorf("span_id should NOT be inside mygroup, but found mygroup.otel.span_id in output: %s", lines[0])
		}
	})
}
