package slogx

import (
	"context"
	"runtime"
	"time"

	"log/slog"

	"go.innotegrity.dev/errorx"
)

// errorOptionsContext can be used to retrieve error options the context.
type errorOptionsContext struct{}

type ErrorOptions struct {
	// AdjectPCFramesBy adjusts the default frame count of 2 by the given amount when retrieving the program counter.
	AdjustPCFramesBy int

	// IgnorePC determines whether or not to retrieve the program counter (for determining file+line) when creating
	// the error roecord.
	IgnorePC bool
}

// DefaultErrorOptions returns a default set of error options.
func DefaultErrorOptions() ErrorOptions {
	return ErrorOptions{
		IgnorePC:         false,
		AdjustPCFramesBy: 0,
	}
}

// GetErrorOptionsFromContext retrieves the options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func GetErrorOptionsFromContext(ctx context.Context) *ErrorOptions {
	o := ctx.Value(errorOptionsContext{})
	if o != nil {
		if opts, ok := o.(*ErrorOptions); ok {
			return opts
		}
	}
	opts := DefaultErrorOptions()
	return &opts
}

// AddToContext adds the options to the given context and returns the new context.
func (o *ErrorOptions) AddToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, errorOptionsContext{}, o)
}

// ErrorRecord holds extended error information an a log record created when the error occurred.
//
// The error can be logged at a later time using slogx.LogRecord().
type ErrorRecord struct {
	// Error holds details of the extended error that occurred.
	Error errorx.Error

	// Record is the actual record which can be logged at a later time.
	Record slog.Record
}

// NewErrorRecord creates a new ErrorRecord object.
func NewErrorRecord(ctx context.Context, level slog.Leveler, msg string, err errorx.Error) *ErrorRecord {
	opts := GetErrorOptionsFromContext(ctx)

	// include program counter, if desired
	pc := uintptr(0)
	if !opts.IgnorePC {
		var pcs [1]uintptr
		// skip runtime.Callers, this function
		runtime.Callers(2+opts.AdjustPCFramesBy, pcs[:])
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
