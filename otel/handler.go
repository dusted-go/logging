package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// Handler is a slog.Handler that adds OpenTelemetry trace context
// (trace_id and span_id) to log records. It wraps another handler and
// ensures trace attributes are always added at the root level in an "otel" group.
type Handler struct {
	handler  slog.Handler
	preAttrs []slog.Attr // Attributes to prepend (including trace attrs)
	groups   []string    // Current group path
}

// Wrap creates a new OpenTelemetry-aware handler that wraps
// the provided handler. When a valid span context is present in the
// context passed to logging methods, it automatically adds trace_id
// and span_id attributes at the root level in an "otel" group.
func Wrap(handler slog.Handler) *Handler {
	return &Handler{
		handler:  handler,
		preAttrs: nil,
		groups:   nil,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle processes the Record by adding trace context if present,
// then delegates to the wrapped handler.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	// Check for span context
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() && len(h.preAttrs) == 0 {
		// No trace context and no pre-attrs, just pass through
		return h.handler.Handle(ctx, r)
	}

	// We need to inject attributes at the root level
	// Create a new record with our pre-attrs first
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	// Add trace attributes if present
	if span.SpanContext().IsValid() {
		newRecord.AddAttrs(
			slog.Group("otel",
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			),
		)
	}

	// Add any pre-attrs
	newRecord.AddAttrs(h.preAttrs...)

	// Now we need to handle groups properly
	// We'll rebuild the structure with groups
	if len(h.groups) > 0 {
		// We need to wrap the remaining attrs in the group structure
		var groupedAttrs []slog.Attr
		r.Attrs(func(a slog.Attr) bool {
			groupedAttrs = append(groupedAttrs, a)
			return true
		})

		// Build nested groups from inside out
		// Convert attrs to any slice
		anyAttrs := make([]any, len(groupedAttrs))
		for i, a := range groupedAttrs {
			anyAttrs[i] = a
		}

		current := slog.Group(h.groups[len(h.groups)-1], anyAttrs...)
		for i := len(h.groups) - 2; i >= 0; i-- {
			current = slog.Group(h.groups[i], current)
		}

		newRecord.AddAttrs(current)
	} else {
		// No groups, just add the remaining attributes
		r.Attrs(func(a slog.Attr) bool {
			newRecord.AddAttrs(a)
			return true
		})
	}

	// Use the base handler (not the grouped one)
	return h.handler.Handle(ctx, newRecord)
}

// WithAttrs returns a new Handler that includes the given attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	if len(h.groups) == 0 {
		// At root level, add to preAttrs
		newPreAttrs := make([]slog.Attr, len(h.preAttrs)+len(attrs))
		copy(newPreAttrs, h.preAttrs)
		copy(newPreAttrs[len(h.preAttrs):], attrs)

		return &Handler{
			handler:  h.handler,
			preAttrs: newPreAttrs,
			groups:   h.groups,
		}
	}

	// In a group, need to use wrapped handler
	return &Handler{
		handler:  h.handler.WithAttrs(attrs),
		preAttrs: h.preAttrs,
		groups:   h.groups,
	}
}

// WithGroup returns a new Handler that starts a group.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	// Don't call WithGroup on the wrapped handler when we have groups
	// We'll handle the grouping ourselves in Handle
	var newHandler slog.Handler
	if len(h.groups) == 0 {
		// First group, keep the base handler
		newHandler = h.handler
	} else {
		// Already have groups, propagate
		newHandler = h.handler.WithGroup(name)
	}

	return &Handler{
		handler:  newHandler,
		preAttrs: h.preAttrs,
		groups:   newGroups,
	}
}
