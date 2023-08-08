package slogx

import (
	"context"

	"golang.org/x/exp/slog"
)

type slogKey struct{}

// FromContext retrieves the logger object stored in the given context, if it exists.
//
// If the logger cannot be found, the default logger is returned instead.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(slogKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// NewContext copies the given context and returns a new context with the given logger stored in it.
func NewContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, slogKey{}, l)
}
