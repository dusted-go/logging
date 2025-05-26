package slogging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	timeFormat = "[15:04:05.000]"

	reset = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

type encoder string

const (
	JSON           = encoder("json")
	YAML           = encoder("yaml")
	defaultEncoder = JSON
)

func colorizer(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

// Handler is a slog.Handler implementation that outputs human-readable,
// colorized log messages for development use. It wraps the standard
// slog.JSONHandler and transforms its output into a pretty format.
type Handler struct {
	handler          slog.Handler
	replaceAttrFunc  func([]string, slog.Attr) slog.Attr
	buffer           *bytes.Buffer
	mutex            *sync.Mutex
	writer           io.Writer
	colorize         bool
	outputEmptyAttrs bool
	encoder          encoder
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		handler:          h.handler.WithAttrs(attrs),
		buffer:           h.buffer,
		encoder:          h.encoder,
		replaceAttrFunc:  h.replaceAttrFunc,
		mutex:            h.mutex,
		writer:           h.writer,
		colorize:         h.colorize,
		outputEmptyAttrs: h.outputEmptyAttrs,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		handler:          h.handler.WithGroup(name),
		buffer:           h.buffer,
		encoder:          h.encoder,
		replaceAttrFunc:  h.replaceAttrFunc,
		mutex:            h.mutex,
		writer:           h.writer,
		colorize:         h.colorize,
		outputEmptyAttrs: h.outputEmptyAttrs,
	}
}

func (h *Handler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.mutex.Lock()
	defer func() {
		h.buffer.Reset()
		h.mutex.Unlock()
	}()
	if err := h.handler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.buffer.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	colorize := func(code int, value string) string {
		return value
	}
	if h.colorize {
		colorize = colorizer
	}

	var level string
	levelAttr := slog.Attr{
		Key:   slog.LevelKey,
		Value: slog.AnyValue(r.Level),
	}
	if h.replaceAttrFunc != nil {
		levelAttr = h.replaceAttrFunc([]string{}, levelAttr)
	}

	if !levelAttr.Equal(slog.Attr{}) {
		level = levelAttr.Value.String() + ":"

		if r.Level <= slog.LevelDebug {
			level = colorize(lightGray, level)
		} else if r.Level <= slog.LevelInfo {
			level = colorize(cyan, level)
		} else if r.Level < slog.LevelWarn {
			level = colorize(lightBlue, level)
		} else if r.Level < slog.LevelError {
			level = colorize(lightYellow, level)
		} else if r.Level == slog.LevelError {
			level = colorize(lightRed, level)
		} else {
			level = colorize(lightMagenta, level)
		}
	}

	var timestamp string
	timeAttr := slog.Attr{
		Key:   slog.TimeKey,
		Value: slog.StringValue(r.Time.Format(timeFormat)),
	}
	if h.replaceAttrFunc != nil {
		timeAttr = h.replaceAttrFunc([]string{}, timeAttr)
	}
	if !timeAttr.Equal(slog.Attr{}) {
		timestamp = colorize(lightGray, timeAttr.Value.String())
	}

	var msg string
	msgAttr := slog.Attr{
		Key:   slog.MessageKey,
		Value: slog.StringValue(r.Message),
	}
	if h.replaceAttrFunc != nil {
		msgAttr = h.replaceAttrFunc([]string{}, msgAttr)
	}
	if !msgAttr.Equal(slog.Attr{}) {
		msg = colorize(white, msgAttr.Value.String())
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	var attrsAsBytes []byte
	if h.outputEmptyAttrs || len(attrs) > 0 {
		switch h.encoder {
		case JSON:
			attrsAsBytes, err = json.MarshalIndent(attrs, "", "  ")
		case YAML:
			attrsAsBytes, err = yaml.Marshal(attrs)
			attrsAsBytes = append([]byte{'\n'}, attrsAsBytes...)
		default:
			return fmt.Errorf("unsupported encoder %q", h.encoder)
		}
		if err != nil {
			return fmt.Errorf("error when marshaling attrs: %w", err)
		}
	}

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}
	if len(level) > 0 {
		out.WriteString(level)
		out.WriteString(" ")
	}
	if len(msg) > 0 {
		out.WriteString(msg)
		out.WriteString(" ")
	}
	if len(attrsAsBytes) > 0 {
		out.WriteString(colorize(darkGray, string(attrsAsBytes)))
	}

	if h.writer != nil {
		_, err = io.WriteString(h.writer, out.String()+"\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

// New creates a new Handler with the given options. If handlerOptions is nil,
// default options are used. Additional configuration can be applied using
// Option functions.
func New(handlerOptions *slog.HandlerOptions, options ...Option) *Handler {
	if handlerOptions == nil {
		handlerOptions = &slog.HandlerOptions{}
	}

	buf := &bytes.Buffer{}
	handler := &Handler{
		buffer:  buf,
		encoder: defaultEncoder,
		handler: slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level:       handlerOptions.Level,
			AddSource:   handlerOptions.AddSource,
			ReplaceAttr: suppressDefaults(handlerOptions.ReplaceAttr),
		}),
		replaceAttrFunc: handlerOptions.ReplaceAttr,
		mutex:           &sync.Mutex{},
	}

	for _, opt := range options {
		opt(handler)
	}

	return handler
}

// NewHandler creates a new Handler with sensible defaults for development:
// - Output to stdout
// - Colorized output
// - Empty attributes shown as {}
func NewHandler(opts *slog.HandlerOptions) *Handler {
	return New(opts, WithDestinationWriter(os.Stdout), WithColor(), WithOutputEmptyAttrs())
}

// Option is a function that configures a Handler.
type Option func(h *Handler)

// WithDestinationWriter sets the writer where log output will be written.
// If writer is nil, log output will be discarded.
func WithDestinationWriter(writer io.Writer) Option {
	return func(h *Handler) {
		h.writer = writer
	}
}

// WithColor enables ANSI color codes in the log output for better readability.
func WithColor() Option {
	return func(h *Handler) {
		h.colorize = true
	}
}

// WithOutputEmptyAttrs configures the handler to output empty attribute objects
// as {} even when no attributes are present in the log record.
func WithOutputEmptyAttrs() Option {
	return func(h *Handler) {
		h.outputEmptyAttrs = true
	}
}

// WithEncoder sets the encoding format for log attributes.
// Supported formats are JSON and YAML.
func WithEncoder(e encoder) Option {
	return func(h *Handler) {
		h.encoder = e
	}
}
