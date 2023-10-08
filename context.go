package slogx

import (
	"context"
	"log/slog"
)

// loggerContextKey is used to store a Logger object in a standard Go context object.
type loggerContextKey struct {
	name string
}

// handlerContextKey is used to store a Handler in a standard Go context object.
type handlerContextKey struct {
	name string
}

// ContextWithHandler copies the given context and returns a new context with the given handler stored in it with
// the given name.
func ContextWithHandler(ctx context.Context, h slog.Handler, name string) context.Context {
	return context.WithValue(ctx, handlerContextKey{name: name}, h)
}

// ContextWithLogger copies the given context and returns a new context with the given logger stored in it with
// the given name.
func ContextWithLogger(ctx context.Context, l *Logger, name string) context.Context {
	return context.WithValue(ctx, loggerContextKey{name: name}, l)
}

// HandlerFromContext retrieves the handler stored in the given context with the given name, if it exists.
//
// If the handler cannot be found, nil is returned.
func HandlerFromContext(ctx context.Context, name string) slog.Handler {
	if v := ctx.Value(handlerContextKey{name: name}); v != nil {
		if l, ok := v.(slog.Handler); ok {
			return l
		}
	}
	return nil
}

// LoggerFromContext retrieves the logger object stored in the given context with the given name, if it exists.
//
// If the logger cannot be found, the default logger is returned instead.
func LoggerFromContext(ctx context.Context, name string) *Logger {
	if v := ctx.Value(loggerContextKey{name: name}); v != nil {
		if l, ok := v.(*Logger); ok {
			return l
		}
	}
	return Default()
}
