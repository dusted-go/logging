package otel

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func BenchmarkHandler(b *testing.B) {
	// Prepare test data
	testMessage := "test log message"
	testAttrs := []any{"key1", "value1", "key2", "value2", "key3", 123}

	// Create tracer for span context tests
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("bench-tracer")

	b.Run("BaselineTextHandler", func(b *testing.B) {
		buf := &bytes.Buffer{}
		handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		logger := slog.New(handler)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("OtelHandler_NoSpan", func(b *testing.B) {
		buf := &bytes.Buffer{}
		baseHandler := slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("OtelHandler_WithSpan", func(b *testing.B) {
		buf := &bytes.Buffer{}
		baseHandler := slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		// Create span context once
		ctx, span := tracer.Start(context.Background(), "bench-span")
		defer span.End()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.InfoContext(ctx, testMessage, testAttrs...)
		}
	})

	b.Run("OtelHandler_WithSpanAndGroups", func(b *testing.B) {
		buf := &bytes.Buffer{}
		baseHandler := slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler).WithGroup("service").WithGroup("component")

		// Create span context once
		ctx, span := tracer.Start(context.Background(), "bench-span")
		defer span.End()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.InfoContext(ctx, testMessage, testAttrs...)
		}
	})

	b.Run("OtelHandler_DiscardOutput", func(b *testing.B) {
		baseHandler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		// Create span context once
		ctx, span := tracer.Start(context.Background(), "bench-span")
		defer span.End()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.InfoContext(ctx, testMessage, testAttrs...)
		}
	})

	b.Run("OtelHandler_InvalidSpan", func(b *testing.B) {
		buf := &bytes.Buffer{}
		baseHandler := slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		// Create invalid span context
		ctx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(context.Background()))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.InfoContext(ctx, testMessage, testAttrs...)
		}
	})
}

func BenchmarkHandlerMemory(b *testing.B) {
	testMessage := "test log message"
	testAttrs := []any{"key1", "value1", "key2", "value2", "key3", 123}

	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("bench-tracer")

	b.Run("OtelHandler_WithSpan_Memory", func(b *testing.B) {
		baseHandler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		ctx, span := tracer.Start(context.Background(), "bench-span")
		defer span.End()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.InfoContext(ctx, testMessage, testAttrs...)
		}
	})

	b.Run("OtelHandler_NoSpan_Memory", func(b *testing.B) {
		baseHandler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("BaselineTextHandler_Memory", func(b *testing.B) {
		handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		logger := slog.New(handler)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(testMessage, testAttrs...)
		}
	})
}

func BenchmarkHandlerWithAttrs(b *testing.B) {
	testMessage := "test log message"
	testAttrs := []any{"key1", "value1", "key2", "value2"}

	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("bench-tracer")

	b.Run("OtelHandler_WithAttrs", func(b *testing.B) {
		baseHandler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		handler := Wrap(baseHandler)
		logger := slog.New(handler).With("service", "test", "version", "1.0")

		ctx, span := tracer.Start(context.Background(), "bench-span")
		defer span.End()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.InfoContext(ctx, testMessage, testAttrs...)
		}
	})

	b.Run("BaselineTextHandler_WithAttrs", func(b *testing.B) {
		handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		logger := slog.New(handler).With("service", "test", "version", "1.0")

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(testMessage, testAttrs...)
		}
	})
}
