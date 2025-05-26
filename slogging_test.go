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
