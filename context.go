package slogx

import (
	"context"
)

// loggerContextKey is used to store a Logger object in a standard Go context object.
type loggerContextKey struct{}

// FromContext retrieves the logger object stored in the given context, if it exists.
//
// If the logger cannot be found, the default logger is returned instead.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerContextKey{}).(*Logger); ok {
		return l
	}
	return Default()
}

// NewContext copies the given context and returns a new context with the given logger stored in it.
func NewContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, l)
}
