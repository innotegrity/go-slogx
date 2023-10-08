package handler

import (
	"context"

	"log/slog"

	"go.innotegrity.dev/async"
	"go.innotegrity.dev/slogx"
)

// ConditionMatchesFn is called to determine whether or not the given record should be logged.
type ConditionMatchesFn func(ctx context.Context, r slog.Record) bool

// Condition defines the condition(s) which must all be true in order to log a message to the given handler.
//
// If no conditions are specified, the handler will always be used to log the messages.
type condition struct {
	// unexported variables
	handler    slog.Handler
	matcherFns []ConditionMatchesFn
}

// NewCondition defines one or more functions to call to determine whether or not to log to the given handler.
//
// If no conditions are specified, the handler will always be used to log the messages.
func NewCondition(handler slog.Handler, matcher ...ConditionMatchesFn) *condition {
	return &condition{
		matcherFns: matcher,
		handler:    handler,
	}
}

// Add simply adds one or more additional conditions to the existing condition and returns the updated object.
func (c *condition) Add(matcher ...ConditionMatchesFn) *condition {
	c.matcherFns = append(c.matcherFns, matcher...)
	return c
}

// WithCondition adds one or more additional conditions to the existing condition and returns a new condition.
func (c condition) WithCondition(matcher ...ConditionMatchesFn) *condition {
	return &condition{
		matcherFns: append(c.matcherFns, matcher...),
		handler:    c.handler,
	}
}

// WithHandler updates the underlying handler and returns a new condition.
func (c condition) WithHandler(handler slog.Handler) *condition {
	return &condition{
		matcherFns: c.matcherFns,
		handler:    handler,
	}
}

// conditionalHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type conditionalHandlerOptionsContext struct{}

// ConditionalHandlerOptions holds the options available when creating the conditionalHandler object.
type ConditionalHandlerOptions struct {
	// ContinueOnError determines whether or not to continue looking for handlers to log to if an error occurs while
	// logging with a matching handler or running any middleware.
	ContinueOnError bool

	// EnableAsync will execute the Handle() function in a separate goroutine in case there are time-consuming
	// conditions which must be evaluated before determining which handler(s) to use for writing the record.
	//
	// When async is enabled, you should be sure to call the Shutdown() function or use the slogx.Shutdown()
	// function to ensure all goroutines are finished and any pending records have been written.
	EnableAsync bool
}

// ConditionalHandlerOptionsFromContext retrieves the options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func ConditionalHandlerOptionsFromContext(ctx context.Context) *ConditionalHandlerOptions {
	o := ctx.Value(conditionalHandlerOptionsContext{})
	if o != nil {
		if opts, ok := o.(*ConditionalHandlerOptions); ok {
			return opts
		}
	}
	opts := DefaultConditionalHandlerOptions()
	return &opts
}

// ContextWithConditionalHandlerOptions adds the options to the given context and returns the new context.
func ContextWithConditionalHandlerOptions(ctx context.Context, opts ConditionalHandlerOptions) context.Context {
	return context.WithValue(ctx, conditionalHandlerOptionsContext{}, &opts)
}

// DefaultConditionalHandlerOptions returns a default set of options for the handler.
func DefaultConditionalHandlerOptions() ConditionalHandlerOptions {
	return ConditionalHandlerOptions{}
}

// conditionalHandler sends the log message to any handler for which the record being logged matches one or more
// conditions.
//
// If multiple handlers have a matching condition, the message will be sent to multiple handlers. If no handler has
// a matching condition, the message is not logged.
type conditionalHandler struct {
	// unexported variables
	conditions []*condition
	futures    []async.Future
	options    ConditionalHandlerOptions
}

// NewConditionalHandler creates a new handler object.
func NewConditionalHandler(opts ConditionalHandlerOptions, cond ...*condition) *conditionalHandler {
	return &conditionalHandler{
		conditions: cond,
		futures:    []async.Future{},
		options:    opts,
	}
}

// Enabled always returns true for this handler as there is no way for it to test whether or not a handler
// would be enabled without passing in a record.
func (h conditionalHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return true
}

// Handle is responsible for finding one or more matching handlers to write the record to.
func (h *conditionalHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := ContextWithConditionalHandlerOptions(ctx, h.options)
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
func (h conditionalHandler) Shutdown(continueOnError bool) error {
	for _, c := range h.conditions {
		if sh, ok := c.handler.(slogx.ShutdownableHandler); ok {
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
func (h conditionalHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	conditions := []*condition{}
	for _, c := range h.conditions {
		conditions = append(conditions, NewCondition(c.handler.WithAttrs(attrs), c.matcherFns...))
	}
	handler := NewConditionalHandler(h.options, conditions...)
	handler.futures = h.futures
	return handler
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h conditionalHandler) WithGroup(name string) slog.Handler {
	conditions := []*condition{}
	for _, c := range h.conditions {
		conditions = append(conditions, NewCondition(c.handler.WithGroup(name), c.matcherFns...))
	}
	handler := NewConditionalHandler(h.options, conditions...)
	handler.futures = h.futures
	return handler
}

// handle is responsible for actually writing the record to the appropriate handler(s).
func (h conditionalHandler) handle(ctx context.Context, r slog.Record) error {
	for _, c := range h.conditions {
		if h.matchesAll(ctx, r, c.matcherFns) {
			if err := c.handler.Handle(ctx, r); err != nil && !h.options.ContinueOnError {
				return err
			}
		}
	}
	return nil
}

// matchesAll determines whether or not the record given matches all of the given conditions.
func (h conditionalHandler) matchesAll(ctx context.Context, r slog.Record, m []ConditionMatchesFn) bool {
	for _, fn := range m {
		if fn != nil && !fn(ctx, r) {
			return false
		}
	}
	return true
}
