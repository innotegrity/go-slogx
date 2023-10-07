package handler

import (
	"context"

	"log/slog"

	"go.innotegrity.dev/slogx"
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

// GetPipeHandlerOptionsFromContext retrieves the options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func GetPipeHandlerOptionsFromContext(ctx context.Context) *PipeHandlerOptions {
	o := ctx.Value(pipeHandlerOptionsContext{})
	if o != nil {
		if opts, ok := o.(*PipeHandlerOptions); ok {
			return opts
		}
	}
	opts := DefaultPipeHandlerOptions()
	return &opts
}

// AddToContext adds the options to the given context and returns the new context.
func (o *PipeHandlerOptions) AddToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, pipeHandlerOptionsContext{}, o)
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
	handlerCtx := h.options.AddToContext(ctx)
	if h.next == nil {
		return false
	}
	return h.next.Enabled(handlerCtx, l)
}

// Handle runs the record through all of the pipe functions and then sends it on to the next handler.
func (h *pipeHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := h.options.AddToContext(ctx)
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

// Level returns the current logging level that is in use by the handler.
//
// In the case of a pipe handler, the level returned is that of next handler in the pipe if it implements the
// slogx.DynamicLevelHandler interface or slogx.LevelPanic if not.
func (h pipeHandler) Level() slogx.Level {
	l := slogx.LevelPanic
	if h.next != nil {
		if levelHandler, ok := h.next.(slogx.DynamicLevelHandler); ok {
			if levelHandler.Level() < l {
				l = levelHandler.Level()
			}
		}
	}
	return l
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

// WithLevel returns a new handler with the given logging level set.
//
// In the case of a pipe handler, if the next handler implements the slogx.DynamicLevelHandler interface, a new
// handler with the level set accordingly will be set as the next handler.
//
// If there is no next handler or the next handler does not implement the slogx.DynamicLevelHandler interface, the
// existing object is returned instead.
func (h pipeHandler) WithLevel(level slogx.Level) slogx.DynamicLevelHandler {
	if h.next != nil {
		if levelHandler, ok := h.next.(slogx.DynamicLevelHandler); ok {
			return NewPipeHandler(h.options, levelHandler.WithLevel(level))
		}
	}
	return &h
}
