package multi

import (
	"context"

	"go.innotegrity.dev/slogx"
	"golang.org/x/exp/slog"
)

// RoundRobinHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type RoundRobinHandlerOptionsContext struct{}

// RoundRobinHandlerOptions holds the options available when creating the roundRobinHandler object.
type RoundRobinHandlerOptions struct {
	// ContinueOnError determines whether or not to continue if an error occurs running middleware.
	ContinueOnError bool
}

// roundRobinHandler simply sends the log message to the next available handler after whichever handler was last used
// successfully.
type roundRobinHandler struct {
	// unexported variables
	handlers  []slog.Handler
	lastIndex int
	options   RoundRobinHandlerOptions
}

// NewRoundRobinHandler creates a new handler object.
func NewRoundRobinHandler(opts RoundRobinHandlerOptions, handlers ...slog.Handler) *roundRobinHandler {
	return &roundRobinHandler{
		handlers:  handlers,
		lastIndex: 0,
		options:   opts,
	}
}

// Enabled determines whether or not the given level is enabled for any handler.
func (h roundRobinHandler) Enabled(ctx context.Context, l slog.Level) bool {
	handlerCtx := context.WithValue(ctx, RoundRobinHandlerOptionsContext{}, &h.options)
	for _, handler := range h.handlers {
		if handler.Enabled(handlerCtx, l) {
			return true
		}
	}
	return false
}

// Handle is responsible for finding the next available handler after the last previously used handler to write the
// ßrecord.
func (h *roundRobinHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	handlers := append(h.handlers[h.lastIndex:], h.handlers[:h.lastIndex]...)
	handlerCtx := context.WithValue(ctx, RoundRobinHandlerOptionsContext{}, &h.options)

	for i, handler := range handlers {
		if handler.Enabled(handlerCtx, r.Level) {
			if err = handler.Handle(handlerCtx, r); err == nil {
				h.lastIndex = i
				return nil
			}
		}
	}

	// exhausted all handlers - return the error from the last one tried
	return err
}

// Shutdown is responsible for cleaning up resources used by the handler.
func (h roundRobinHandler) Shutdown(continueOnError bool) error {
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
func (h roundRobinHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := []slog.Handler{}
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	handler := NewRoundRobinHandler(h.options, handlers...)
	return handler
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h roundRobinHandler) WithGroup(name string) slog.Handler {
	handlers := []slog.Handler{}
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	handler := NewRoundRobinHandler(h.options, handlers...)
	return handler
}
