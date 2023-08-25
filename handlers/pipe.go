package handlers

import (
	"context"

	"golang.org/x/exp/slog"
)

// PipeHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type PipeHandlerOptionsContext struct{}

// PipeHandlerFn should take the clone of the given record, modify it as needed and return the modified version.
type PipeHandlerFn func(context.Context, slog.Record) (slog.Record, error)

// PipeHandlerOptions holds the options for the pipe handler.
type PipeHandlerOptions struct {
	// ContinueOnError determines whether or not to continue logging to handlers to if an error occurs while running any
	// of the pipe functions.
	ContinueOnError bool

	// PipeFns defines the list of functions to pipe the record through before passing it onto the next handler.
	PipeFns []PipeHandlerFn
}

// pipeHandler is a handler which sends the record through one or more functions before passing it onto the next handler.
type pipeHandler struct {
	// unexported variables
	next    slog.Handler
	options PipeHandlerOptions
}

// NewPipeHandler creates a new object.
func NewPipeHandler(opts PipeHandlerOptions, next slog.Handler) *pipeHandler {
	return &pipeHandler{
		options: opts,
		next:    next,
	}
}

// Enabled determines whether or not the given level is enabled for any handler.
func (h pipeHandler) Enabled(ctx context.Context, l slog.Level) bool {
	handlerCtx := context.WithValue(ctx, PipeHandlerOptionsContext{}, &h.options)
	if h.next == nil {
		return false
	}
	return h.next.Enabled(handlerCtx, l)
}

// Handle runs the record through all of the pipe functions and then sends it on to the next handler.
func (h *pipeHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := context.WithValue(ctx, PipeHandlerOptionsContext{}, &h.options)
	if h.next == nil {
		return nil
	}

	// run the pipe functions
	var err error
	record := r.Clone()
	for _, fn := range h.options.PipeFns {
		record, err = fn(handlerCtx, record)
		if err != nil && !h.options.ContinueOnError {
			return err
		}
	}

	// send the record to the next handler
	return h.next.Handle(handlerCtx, r)
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h pipeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.next != nil {
		h.next = h.next.WithAttrs(attrs)
	}
	return NewPipeHandler(h.options, h.next)
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h pipeHandler) WithGroup(name string) slog.Handler {
	if h.next != nil {
		h.next = h.next.WithGroup(name)
	}
	return NewPipeHandler(h.options, h.next)
}
