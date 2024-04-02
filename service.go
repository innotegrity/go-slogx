package slogx

import (
	"context"
)

const (
	// DefaultLoggingServiceName is used if a logging service is being stored in the context with no name.
	DefaultLoggingServiceName = "__default__"
)

// LoggingService defines the interface a generic logging service should provide.
//
// While it's ultimately up to the library calling these functions how to format the arg values of a function,
// it is highly recommended to use standard log/slog library attribute functions such as slog.String() or slog.Any()
// to construct key/value pairs or slog.Group() to create groups of key/value pairs.
//
// Libraries which do not adhere to using log/slog library attribute functions, either as a sender or as a receiver,
// should clearly document how the arguments are expected.
type LoggingService interface {
	// Debug logs a message using DEBUG level.
	Debug(msg string, args ...any)

	// DebugContext logs a message using DEBUG level with context.
	DebugContext(ctx context.Context, msg string, args ...any)

	// Error logs a message using ERROR level.
	Error(msg string, args ...any)

	// ErrorContext logs a message using ERROR level with context.
	ErrorContext(ctx context.Context, msg string, args ...any)

	// Fatal logs a message using FATAL level.
	Fatal(msg string, args ...any)

	// FatalContext logs a message using FATAL level with context.
	FatalContext(ctx context.Context, msg string, args ...any)

	// Info logs a message using INFO level.
	Info(msg string, args ...any)

	// InfoContext logs a message using INFO level with context.
	InfoContext(ctx context.Context, msg string, args ...any)

	// Notice logs a message using NOTICE level.
	Notice(msg string, args ...any)

	// NoticeContext logs a message using NOTICE level with context.
	NoticeContext(ctx context.Context, msg string, args ...any)

	// Panic logs a message using PANIC level.
	Panic(msg string, args ...any)

	// PanicContext logs a message using PANIC level with context.
	PanicContext(ctx context.Context, msg string, args ...any)

	// Trace logs a message using TRACE level.
	Trace(msg string, args ...any)

	// TraceContext logs a message using TRACE level with context.
	TraceContext(ctx context.Context, msg string, args ...any)

	// Warn logs a message using WARN level.
	Warn(msg string, args ...any)

	// WarnContext logs a message using WARN level with context.
	WarnContext(ctx context.Context, msg string, args ...any)
}

// ActiveLoggingServiceFromContext returns the active logging service from the context, if it exists.
//
// If multiple logging services are associated with a context, external libraries can use this function to retrieve
// the name of the appropriate logging service to retrieve for logging debug, error, etc. messages within their code.
//
// If no active logging service is set in the context, the logging service with the default name is returned. If no
// logging services are stored at all, nil is returned.
func ActiveLoggingServiceFromContext(ctx context.Context) LoggingService {
	return LoggingServiceFromContext(ctx, ActiveLoggingServiceNameFromContext(ctx))
}

// ActiveLogingServiceNameFromContext retrieves the active logging service name from the context, if it exists.
//
// If multiple logging services are associated with a context, external libraries can use this function to retrieve
// the name of the appropriate logging service to retrieve for logging debug, error, etc. messages within their code.
//
// If no active logging service name is set in the context, the default name is returned instead.
func ActiveLoggingServiceNameFromContext(ctx context.Context) string {
	if v := ctx.Value(activeLoggingServiceNameContextKey{}); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return DefaultLoggingServiceName
}

// ContextWithActiveLoggingService returns a new context with the given logging service stored with the given
// name and marks it as active.
//
// If multiple logging services are associated with a context, external libraries can use
// ActiveLoggingServiceFromContext() to retrieve the appropriate logging service to use for logging debug,
// error, etc. messages within their code.
//
// If no name is supplied, the default logging service name is used.
func ContextWithActiveLoggingService(ctx context.Context, s LoggingService, name string) context.Context {
	if name == "" {
		name = DefaultLoggingServiceName
	}
	return ContextWithActiveLoggingServiceName(ContextWithLoggingService(ctx, s, name), name)
}

// activeLoggingServiceNameContextKey is used to store the name of the active logging service in a standard Go
// context object if there is more than one logging service in the context.
type activeLoggingServiceNameContextKey struct{}

// ContextWithActiveLoggingServiceName returns a new context with the name of the active logging service set.
//
// If multiple logging services are associated with a context, external libraries can use
// ActiveLoggingServiceNameFromContext() to retrieve the appropriate logger name to then retrieve for logging
// debug, error, etc. messages within their code.
//
// If no name is supplied, the default logging service name is used.
func ContextWithActiveLoggingServiceName(ctx context.Context, name string) context.Context {
	if name == "" {
		name = DefaultLoggingServiceName
	}
	return context.WithValue(ctx, activeLoggingServiceNameContextKey{}, name)
}

// loggingServiceContextKey is used to store a LoggingService in a standard Go context object.
type loggingServiceContextKey struct {
	name string
}

// ContextWithLoggingService copies the given context and returns a new context with the given logging service
// stored in it with the given name.
//
// If no name is supplied, the default logging service name is used.
func ContextWithLoggingService(ctx context.Context, s LoggingService, name string) context.Context {
	if name == "" {
		name = DefaultLoggingServiceName
	}
	return context.WithValue(ctx, loggingServiceContextKey{name: name}, s)
}

// LoggingServiceFromContext retrieves the logging service object stored in the given context with the given name,
// if it exists.
//
// If no name is supplied, the default logging service name is used. If no matching logging service can be found,
// nil is returned instead.
func LoggingServiceFromContext(ctx context.Context, name string) LoggingService {
	if name == "" {
		name = DefaultLoggingServiceName
	}
	if v := ctx.Value(loggingServiceContextKey{name: name}); v != nil {
		if s, ok := v.(LoggingService); ok {
			return s
		}
	}
	return nil
}
