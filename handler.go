package slogx

import (
	"context"
	"log/slog"
)

// LevelVarHandler should be implemented by handlers that use a dynamic level via a LevelVar object.
type LevelVarHandler interface {
	slog.Handler

	// Level returns a pointer to the dynamic level variable.
	Level() *LevelVar
}

// ShutdownableHandler should be implemented by handlers which allocate resources that need cleaning up before an
// application exits.
type ShutdownableHandler interface {
	slog.Handler

	// Shutdown should handle cleaning up any resources created by the handler (eg: closing file handles,
	// DB connections, etc.).
	Shutdown(bool) error
}

// handlerContextKey is used to store a handler in a standard Go context object.
type handlerContextKey struct {
	name string
}

// ContextWithHandler copies the given context and returns a new context with the given handler stored in it with
// the given name.
func ContextWithHandler(ctx context.Context, h slog.Handler, name string) context.Context {
	return context.WithValue(ctx, handlerContextKey{name: name}, h)
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

// handlerOptionsContextKey is used to store a handler's options in a standard Go context object.
type handlerOptionsContextKey struct {
	name string
}

// ContextWithHandlerOptions copies the given context and returns a new context with the given handler options stored
// in it with the given name.
func ContextWithHandlerOptions(ctx context.Context, opts any, name string) context.Context {
	return context.WithValue(ctx, handlerOptionsContextKey{name: name}, opts)
}

// HandlerOptionsFromContext retrieves the handler options stored in the given context with the given name, if
// it exists.
//
// If the handler options cannot be found, nil is returned.
func HandlerOptionsFromContext(ctx context.Context, name string) any {
	if v := ctx.Value(handlerOptionsContextKey{name: name}); v != nil {
		return v
	}
	return nil
}
