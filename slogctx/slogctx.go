package slogctx

import (
	"context"
	"log/slog"
)

type contextKey int

const loggerKey contextKey = 0

// WithLogger adds a *slog.Logger to the current context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	ctx = context.WithValue(ctx, loggerKey, logger)
	return ctx
}

// GetLogger gets a *slog.Logger from context or returns slog.Default().
func GetLogger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
