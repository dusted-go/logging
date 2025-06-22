package slogging

import (
	"bytes"
	"context"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type captureStream struct {
	lines [][]byte
}

func (cs *captureStream) Write(bytes []byte) (int, error) {
	cs.lines = append(cs.lines, bytes)
	return len(bytes), nil
}

func Test_WritesToProvidedStream(t *testing.T) {
	cs := &captureStream{}
	handler := New(nil, WithDestinationWriter(cs), WithOutputEmptyAttrs())
	logger := slog.New(handler)

	logger.Info("testing logger")
	if len(cs.lines) != 1 {
		t.Errorf("expected 1 lines logged, got: %d", len(cs.lines))
	}

	lineMatcher := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\] INFO: testing logger {}`)
	line := string(cs.lines[0])
	if lineMatcher.MatchString(line) == false {
		t.Errorf("expected `testing logger` but found `%s`", line)
	}
	if !strings.HasSuffix(line, "\n") {
		t.Errorf("expected line to be terminated with `\\n` but found `%s`", line[len(line)-1:])
	}
}

func Test_SkipEmptyAttributes(t *testing.T) {
	cs := &captureStream{}
	handler := New(nil, WithDestinationWriter(cs))
	logger := slog.New(handler)

	logger.Info("testing logger")
	if len(cs.lines) != 1 {
		t.Errorf("expected 1 lines logged, got: %d", len(cs.lines))
	}

	lineMatcher := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\] INFO: testing logger`)
	line := string(cs.lines[0])
	if lineMatcher.MatchString(line) == false {
		t.Errorf("expected `testing logger` but found `%s`", line)
	}
	if !strings.HasSuffix(line, "\n") {
		t.Errorf("expected line to be terminated with `\\n` but found `%s`", line[len(line)-1:])
	}
}

func Test_WithAttrsPreservesOutputEmptyAttrs(t *testing.T) {
	cs := &captureStream{}
	handler := New(nil, WithDestinationWriter(cs), WithOutputEmptyAttrs())
	logger := slog.New(handler)

	// Create a new logger with additional attributes
	loggerWithAttrs := logger.With("key", "value")

	// Log a message without any inline attributes
	loggerWithAttrs.Info("test message")

	// The output should still include empty attrs {} because WithOutputEmptyAttrs was set
	if len(cs.lines) != 1 {
		t.Errorf("expected 1 line logged, got: %d", len(cs.lines))
	}

	line := string(cs.lines[0])
	// Should see the key:value from With() and empty attrs should still be shown
	if !strings.Contains(line, `"key": "value"`) {
		t.Errorf("expected to find key:value in output, got: %s", line)
	}
}

func Test_WithGroupPreservesOutputEmptyAttrs(t *testing.T) {
	// First test: handler without outputEmptyAttrs
	cs1 := &captureStream{}
	handler1 := New(nil, WithDestinationWriter(cs1)) // No WithOutputEmptyAttrs
	logger1 := slog.New(handler1)
	loggerWithGroup1 := logger1.WithGroup("mygroup")
	loggerWithGroup1.Info("test message")

	// Second test: handler with outputEmptyAttrs
	cs2 := &captureStream{}
	handler2 := New(nil, WithDestinationWriter(cs2), WithOutputEmptyAttrs())
	logger2 := slog.New(handler2)
	loggerWithGroup2 := logger2.WithGroup("mygroup")
	loggerWithGroup2.Info("test message")

	line1 := string(cs1.lines[0])
	line2 := string(cs2.lines[0])

	t.Logf("Without outputEmptyAttrs: %s", line1)
	t.Logf("With outputEmptyAttrs: %s", line2)

	// They should be different - one should have {} and one shouldn't
	if line1 == line2 {
		t.Errorf("expected different output with and without outputEmptyAttrs, but got same output")
	}
}

func Test_NilWriterHandling(t *testing.T) {
	t.Run("nil writer should not panic", func(t *testing.T) {
		// This might panic if not handled properly
		handler := New(nil, WithDestinationWriter(nil))
		logger := slog.New(handler)

		// Try to log something - this could panic with nil writer
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("logging with nil writer panicked: %v", r)
			}
		}()

		logger.Info("test message")
	})

	t.Run("default writer should be used if none provided", func(t *testing.T) {
		// Create handler without specifying a writer
		handler := New(nil)
		logger := slog.New(handler)

		// This should not panic - should use a default writer or handle gracefully
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("logging without writer option panicked: %v", r)
			}
		}()

		logger.Info("test message")
	})
}

func Test_Encoder(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		handler := New(nil, WithDestinationWriter(buf), WithEncoder(JSON))
		logger := slog.New(handler)

		logger.Info("testing logger", "key1", "value1", "key2", "value2")
		lines := strings.Split(buf.String(), "\n")

		line0Matcher := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\] INFO: testing logger \{`)
		if !line0Matcher.MatchString(lines[0]) {
			t.Errorf("expected `[...] INFO: testing logger {` but found `%s`", lines[0])
		}
		if lines[1] != `  "key1": "value1",` {
			t.Errorf("expected `\"key1\": \"value1\"` but found `%s`", lines[1])
		}
		if lines[2] != `  "key2": "value2"` {
			t.Errorf("expected `\"key2\": \"value2\"` but found `%s`", lines[2])
		}
	})

	t.Run("yaml", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		handler := New(nil, WithDestinationWriter(buf), WithEncoder(YAML))
		logger := slog.New(handler)

		logger.Info("testing logger", "key1", "value1", "key2", "value2")
		lines := strings.Split(buf.String(), "\n")

		line0Matcher := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\] INFO: testing logger`)
		if !line0Matcher.MatchString(lines[0]) {
			t.Errorf("expected `testing logger` but found `%s`", lines[0])
		}
		if lines[1] != "key1: value1" {
			t.Errorf("expected `key1: value1` but found `%s`", lines[1])
		}
		if lines[2] != "key2: value2" {
			t.Errorf("expected `key2: value2` but found `%s`", lines[2])
		}
	})
}

func Test_OpenTelemetrySpanContext(t *testing.T) {
	t.Run("otel span context", func(t *testing.T) {
		// Create a tracer provider and tracer
		tp := sdktrace.NewTracerProvider()
		tracer := tp.Tracer("test-tracer")

		// Create capture stream and handler
		cs := &captureStream{}
		handler := New(nil, WithDestinationWriter(cs), WithOutputEmptyAttrs())
		logger := slog.New(handler)

		// Start a span and log within its context
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// Log with the span context
		logger.InfoContext(ctx, "test message with tracing")

		// Verify log output
		if len(cs.lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(cs.lines))
		}

		line := string(cs.lines[0])
		t.Logf("Log output: %s", line)

		// Extract span context to get expected values
		spanContext := span.SpanContext()
		expectedTraceID := spanContext.TraceID().String()
		expectedSpanID := spanContext.SpanID().String()

		// Check that trace_id and span_id are present in the output
		if !strings.Contains(line, `"trace_id": "`+expectedTraceID+`"`) {
			t.Errorf("expected trace_id %s in output, got: %s", expectedTraceID, line)
		}
		if !strings.Contains(line, `"span_id": "`+expectedSpanID+`"`) {
			t.Errorf("expected span_id %s in output, got: %s", expectedSpanID, line)
		}
	})

	t.Run("logs without span context should not include trace_id and span_id", func(t *testing.T) {
		// Create capture stream and handler
		cs := &captureStream{}
		handler := New(nil, WithDestinationWriter(cs), WithOutputEmptyAttrs())
		logger := slog.New(handler)

		// Log without span context
		logger.Info("test message without tracing")

		// Verify log output
		if len(cs.lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(cs.lines))
		}

		line := string(cs.lines[0])
		t.Logf("Log output: %s", line)

		// Check that trace_id and span_id are NOT present
		if strings.Contains(line, `"trace_id"`) {
			t.Errorf("unexpected trace_id in output: %s", line)
		}
		if strings.Contains(line, `"span_id"`) {
			t.Errorf("unexpected span_id in output: %s", line)
		}
	})

	t.Run("logs with invalid span context should not include trace_id and span_id", func(t *testing.T) {
		// Create capture stream and handler
		cs := &captureStream{}
		handler := New(nil, WithDestinationWriter(cs), WithOutputEmptyAttrs())
		logger := slog.New(handler)

		// Create a context with an invalid span (no-op span)
		ctx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(context.Background()))

		// Log with invalid span context
		logger.InfoContext(ctx, "test message with invalid span")

		// Verify log output
		if len(cs.lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(cs.lines))
		}

		line := string(cs.lines[0])
		t.Logf("Log output: %s", line)

		// Check that trace_id and span_id are NOT present
		if strings.Contains(line, `"trace_id"`) {
			t.Errorf("unexpected trace_id in output: %s", line)
		}
		if strings.Contains(line, `"span_id"`) {
			t.Errorf("unexpected span_id in output: %s", line)
		}
	})

	t.Run("trace attributes should be at root level with groups", func(t *testing.T) {
		// Create a tracer provider and tracer
		tp := sdktrace.NewTracerProvider()
		tracer := tp.Tracer("test-tracer")

		// Create capture stream and handler
		cs := &captureStream{}
		handler := New(nil, WithDestinationWriter(cs))
		logger := slog.New(handler)

		// Create logger with group
		groupedLogger := logger.WithGroup("mygroup")

		// Start a span and log within its context
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// Log with the span context using grouped logger
		groupedLogger.InfoContext(ctx, "test message", "key", "value")

		// Verify log output
		if len(cs.lines) != 1 {
			t.Fatalf("expected 1 line logged, got: %d", len(cs.lines))
		}

		line := string(cs.lines[0])
		t.Logf("Log output: %s", line)

		// Verify trace_id and span_id are present
		if !strings.Contains(line, `"trace_id":`) {
			t.Errorf("trace_id missing from output")
		}
		if !strings.Contains(line, `"span_id":`) {
			t.Errorf("span_id missing from output")
		}

		// Verify they are at root level by checking the JSON structure
		// The output should have mygroup as a separate object at root level
		if !strings.Contains(line, `"mygroup": {`) {
			t.Errorf("mygroup missing from output")
		}

		// Extract just the mygroup content to verify trace attrs are NOT inside it
		groupStart := strings.Index(line, `"mygroup": {`)
		if groupStart != -1 {
			// Find the matching closing brace for mygroup
			braceCount := 0
			inGroup := false
			groupContent := ""
			for i := groupStart; i < len(line); i++ {
				if line[i] == '{' {
					braceCount++
					inGroup = true
				} else if line[i] == '}' {
					braceCount--
					if braceCount == 0 && inGroup {
						groupContent = line[groupStart : i+1]
						break
					}
				}
			}

			// Verify trace attributes are NOT in the group content
			if strings.Contains(groupContent, `"trace_id"`) {
				t.Errorf("trace_id should NOT be inside mygroup, but it is. Group content: %s", groupContent)
			}
			if strings.Contains(groupContent, `"span_id"`) {
				t.Errorf("span_id should NOT be inside mygroup, but it is. Group content: %s", groupContent)
			}

			// Verify the group only contains the expected key
			if !strings.Contains(groupContent, `"key": "value"`) {
				t.Errorf("mygroup should contain key:value")
			}
		}
	})
}
