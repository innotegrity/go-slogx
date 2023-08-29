package handler

import (
	"context"

	"go.innotegrity.dev/slogx"
	"golang.org/x/exp/slog"
)

// FailoverHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type FailoverHandlerOptionsContext struct{}

// FailoverHandlerOptions holds the options available when creating the failoverHandler object.
type FailoverHandlerOptions struct {
	// ContinueOnError determines whether or not to continue if an error occurs running middleware.
	ContinueOnError bool
}

// failoverHandler simply sends the log message to the first available handler.
type failoverHandler struct {
	// unexported variables
	handlers []slog.Handler
	options  FailoverHandlerOptions
}

// NewFailoverHandler creates a new handler object.
func NewFailoverHandler(opts FailoverHandlerOptions, handlers ...slog.Handler) *failoverHandler {
	return &failoverHandler{
		handlers: handlers,
		options:  opts,
	}
}

// Enabled determines whether or not the given level is enabled for any handler.
func (h failoverHandler) Enabled(ctx context.Context, l slog.Level) bool {
	handlerCtx := context.WithValue(ctx, FailoverHandlerOptionsContext{}, &h.options)
	for _, handler := range h.handlers {
		if handler.Enabled(handlerCtx, l) {
			return true
		}
	}
	return false
}

// Handle is responsible for finding the first available handler to write the record.
func (h *failoverHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	handlerCtx := context.WithValue(ctx, FailoverHandlerOptionsContext{}, &h.options)

	for _, handler := range h.handlers {
		if handler.Enabled(handlerCtx, r.Level) {
			if err = handler.Handle(handlerCtx, r); err == nil {
				return nil
			}
		}
	}

	// exhausted all handlers - return the error from the last one tried
	return err
}

// Shutdown is responsible for cleaning up resources used by the handler.
func (h failoverHandler) Shutdown(continueOnError bool) error {
	for _, handler := range h.handlers {
		if sh, ok := handler.(slogx.ShutdownableHandler); ok {
			if err := sh.Shutdown(continueOnError); err != nil && !continueOnError {
				return err
			}
		}
	}
	return nil
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h failoverHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := []slog.Handler{}
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	handler := NewFailoverHandler(h.options, handlers...)
	return handler
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h failoverHandler) WithGroup(name string) slog.Handler {
	handlers := []slog.Handler{}
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	handler := NewFailoverHandler(h.options, handlers...)
	return handler
}
