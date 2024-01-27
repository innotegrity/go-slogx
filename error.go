package slogx

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"go.innotegrity.dev/errorx"
)

const (
	// default values
	defaultErrorAttrName = "error"
)

// errorOptionsContext is used to store an [ErrorOptions] object within a standard Go context object.
type errorOptionsContext struct{}

// ErrorOptions stores options for working with error records.
type ErrorOptions struct {
	// AdjustFrameCount indicates a number of frames to adjust the skip by when calling runtime.Callers. By default,
	// 3 frames are skipped when creating the record.
	AdjustFrameCount int

	// IncludeFileLine indicates whether or not to invoke runtime.Callers to get the program counter in order to retrieve
	// source file information.
	IncludeFileLine bool
}

// NewErrorOptions creates a new object with default values.
func NewErrorOptions() *ErrorOptions {
	return &ErrorOptions{
		IncludeFileLine:  false,
		AdjustFrameCount: 0,
	}
}

// FromContext retrieves the error options stored inside the context and replaces the current options with those.
//
// If no options are found in the context, the object is left unchanged. The function returns the object itself
// as a convenient method for assigning the value.
func (o *ErrorOptions) FromContext(ctx context.Context) *ErrorOptions {
	opts := ctx.Value(errorOptionsContext{})
	if opts != nil {
		if opts, ok := opts.(*ErrorOptions); ok {
			o = opts
		}
	}
	return o
}

// SaveToContext stores the options within the given context and returns the new context.
func (o *ErrorOptions) SaveToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, errorOptionsContext{}, o)
}

// ErrorRecord holds extended error information an a log record created when the error occurred.
//
// The error can be logged at a later time using [LogRecord].
type ErrorRecord struct {
	// Error holds details of the extended error that occurred.
	Error errorx.Error

	// Record is the actual record which can be logged at a later time.
	Record slog.Record
}

// NewErrorRecord creates a new ErrorRecord object.
func NewErrorRecord(ctx context.Context, level slog.Leveler, msg string, err errorx.Error) *ErrorRecord {
	opts := NewErrorOptions().FromContext(ctx)

	// include program counter, if desired
	pc := uintptr(0)
	if opts.IncludeFileLine {
		var pcs [1]uintptr
		// skip runtime.Callers, this function
		runtime.Callers(2+opts.AdjustFrameCount, pcs[:])
		pc = pcs[0]
	}

	// create the record and attach the error as an attribute
	rec := slog.NewRecord(time.Now().UTC(), level.Level(), msg, pc)
	rec.AddAttrs(ErrX("error", err))
	return &ErrorRecord{
		Error:  err,
		Record: rec,
	}
}

// errorAttrNameContext is used to store the name of the error attribute to use in log messages.
type errorAttrNameContext struct{}

// ContextWithErrorAttrName returns a new context with the given error attribute name stored.
//
// If no name is supplied, the default attribute name is used instead.
func ContextWithErrorAttrName(ctx context.Context, name string) context.Context {
	if name == "" {
		name = defaultErrorAttrName
	}
	return context.WithValue(ctx, errorAttrNameContext{}, name)
}

// ErrorAttrNameFromContext returns the attribute name to use for logging error messages.
func ErrorAttrNameFromContext(ctx context.Context) string {
	if v := ctx.Value(errorAttrNameContext{}); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return defaultErrorAttrName
}
