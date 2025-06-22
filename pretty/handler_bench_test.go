package pretty

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
)

// BenchmarkHandlers compares performance of different slog handlers
func BenchmarkHandlers(b *testing.B) {
	// Prepare test data
	testMessage := "test log message"
	testAttrs := []any{"key1", "value1", "key2", "value2", "key3", 123}

	b.Run("StandardJSONHandler", func(b *testing.B) {
		buf := &bytes.Buffer{}
		handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		logger := slog.New(handler)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("StandardTextHandler", func(b *testing.B) {
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

	b.Run("SloggingHandler", func(b *testing.B) {
		buf := &bytes.Buffer{}
		handler := NewHandler(WithWriter(buf), WithLevel(slog.LevelInfo))
		logger := slog.New(handler)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("SloggingHandlerWithColor", func(b *testing.B) {
		buf := &bytes.Buffer{}
		handler := NewHandler(WithWriter(buf), WithLevel(slog.LevelInfo), WithColor())
		logger := slog.New(handler)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("SloggingHandlerYAML", func(b *testing.B) {
		buf := &bytes.Buffer{}
		handler := NewHandler(WithWriter(buf), WithLevel(slog.LevelInfo), WithEncoder(YAML))
		logger := slog.New(handler)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("SloggingHandlerDiscard", func(b *testing.B) {
		handler := NewHandler(WithWriter(io.Discard), WithLevel(slog.LevelInfo))
		logger := slog.New(handler)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(testMessage, testAttrs...)
		}
	})
}

// BenchmarkLogLevels tests performance across different log levels
func BenchmarkLogLevels(b *testing.B) {
	buf := &bytes.Buffer{}
	handler := NewHandler(WithWriter(buf), WithLevel(slog.LevelDebug), WithColor())
	logger := slog.New(handler)

	testMessage := "benchmark message"
	testAttrs := []any{"key", "value"}

	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Debug(testMessage, testAttrs...)
		}
	})

	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, testAttrs...)
		}
	})

	b.Run("Warn", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Warn(testMessage, testAttrs...)
		}
	})

	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Error(testMessage, testAttrs...)
		}
	})
}

// BenchmarkAttributeCounts tests performance with different numbers of attributes
func BenchmarkAttributeCounts(b *testing.B) {
	buf := &bytes.Buffer{}
	handler := NewHandler(WithLevel(slog.LevelInfo), WithWriter(buf))
	logger := slog.New(handler)

	testMessage := "benchmark message"

	b.Run("NoAttrs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage)
		}
	})

	b.Run("2Attrs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, "key1", "value1")
		}
	})

	b.Run("4Attrs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, "key1", "value1", "key2", "value2")
		}
	})

	b.Run("8Attrs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			logger.Info(testMessage, "key1", "value1", "key2", "value2", "key3", "value3", "key4", "value4")
		}
	})
}
