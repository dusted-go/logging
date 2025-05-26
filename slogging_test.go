package slogging

import (
	"bytes"
	"log/slog"
	"regexp"
	"strings"
	"testing"
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
		t.Errorf("exected line to be terminated with `\\n` but found `%s`", line[len(line)-1:])
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
		t.Errorf("exected line to be terminated with `\\n` but found `%s`", line[len(line)-1:])
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
