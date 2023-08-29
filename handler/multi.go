package handler

import (
	"context"

	"go.innotegrity.dev/async"
	"go.innotegrity.dev/slogx"
	"golang.org/x/exp/slog"
)

// MultiHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type MultiHandlerOptionsContext struct{}

// MultiHandlerOptions holds the options available when creating the multiHandler object.
type MultiHandlerOptions struct {
	// ContinueOnError determines whether or not to continue logging to handlers to if an error occurs while writing to
	// a particular handler or running any middleware.
	ContinueOnError bool

	// EnableAsync will execute the Handle() function in a separate goroutine in case there are slow handlers being
	// used for writing the record.
	//
	// Typically, a specific handler should implement its own async writing if it is slow, but this is a fallback in
	// case it does not.
	//
	// When async is enabled, you should be sure to call the Shutdown() function or use the slogx.Shutdown()
	// function to ensure all goroutines are finished and any pending records have been written.
	EnableAsync bool
}

// multiHandler sends the log message to multiple handlers.
type multiHandler struct {
	// unexported variables
	futures  []async.Future
	handlers []slog.Handler
	options  MultiHandlerOptions
}

// NewMultiHandler creates a new handler object.
func NewMultiHandler(opts MultiHandlerOptions, handler ...slog.Handler) *multiHandler {
	return &multiHandler{
		futures:  []async.Future{},
		handlers: handler,
		options:  opts,
	}
}

// Enabled determines whether or not the given level is enabled for any handler.
func (h multiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	handlerCtx := context.WithValue(ctx, MultiHandlerOptionsContext{}, &h.options)
	for _, handler := range h.handlers {
		if handler.Enabled(handlerCtx, l) {
			return true
		}
	}
	return false
}

// Handle is responsible for writing the record to each and every handler.
func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := context.WithValue(ctx, MultiHandlerOptionsContext{}, &h.options)
	if !h.options.EnableAsync {
		return h.handle(handlerCtx, r)
	}

	future := async.Exec(func() any {
		return h.handle(handlerCtx, r)
	})
	h.futures = append(h.futures, future)
	return nil
}

// Shutdown is responsible for cleaning up resources used by the handler.
func (h multiHandler) Shutdown(continueOnError bool) error {
	for _, handler := range h.handlers {
		if sh, ok := handler.(slogx.ShutdownableHandler); ok {
			if err := sh.Shutdown(continueOnError); err != nil && !continueOnError {
				return err
			}
		}
	}
	for _, f := range h.futures {
		if f != nil {
			f.Await()
		}
	}
	return nil
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := []slog.Handler{}
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	handler := NewMultiHandler(h.options, handlers...)
	handler.futures = h.futures
	return handler
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h multiHandler) WithGroup(name string) slog.Handler {
	handlers := []slog.Handler{}
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	handler := NewMultiHandler(h.options, handlers...)
	handler.futures = h.futures
	return handler
}

// handle is responsible for actually writing the record to the appropriate handler(s).
func (h multiHandler) handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil && !h.options.ContinueOnError {
			return err
		}
	}
	return nil
}
