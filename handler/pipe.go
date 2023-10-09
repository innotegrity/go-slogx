package handler

import (
	"context"

	"log/slog"
)

// pipeHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type pipeHandlerOptionsContext struct{}

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

// DefaultPipeHandlerOptions returns a default set of options for the handler.
func DefaultPipeHandlerOptions() PipeHandlerOptions {
	return PipeHandlerOptions{
		PipeFns: []PipeHandlerFn{},
	}
}

// PipeHandlerOptionsFromContext retrieves the options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func PipeHandlerOptionsFromContext(ctx context.Context) *PipeHandlerOptions {
	o := ctx.Value(pipeHandlerOptionsContext{})
	if o != nil {
		if opts, ok := o.(*PipeHandlerOptions); ok {
			return opts
		}
	}
	opts := DefaultPipeHandlerOptions()
	return &opts
}

// ContextWithPipeHandlerOptions adds the options to the given context and returns the new context.
func ContextWithPipeHandlerOptions(ctx context.Context, opts PipeHandlerOptions) context.Context {
	return context.WithValue(ctx, pipeHandlerOptionsContext{}, &opts)
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

// Enabled returns whether or not the next handler would log this message.
func (h pipeHandler) Enabled(ctx context.Context, l slog.Level) bool {
	handlerCtx := ContextWithPipeHandlerOptions(ctx, h.options)
	if h.next == nil {
		return false
	}
	return h.next.Enabled(handlerCtx, l)
}

// Handle runs the record through all of the pipe functions and then sends it on to the next handler.
func (h *pipeHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := ContextWithPipeHandlerOptions(ctx, h.options)
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
//
// If there is no next handler, the existing object is returned instead.
func (h pipeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.next != nil {
		return NewPipeHandler(h.options, h.next.WithAttrs(attrs))
	}
	return &h
}

// WithGroup creates a new handler from the existing one adding the given group to it.
//
// If there is no next handler, the existing object is returned instead.
func (h pipeHandler) WithGroup(name string) slog.Handler {
	if h.next != nil {
		return NewPipeHandler(h.options, h.next.WithGroup(name))
	}
	return &h
}
