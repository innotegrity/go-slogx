package slogx

import (
	"context"
	"runtime"
	"time"

	"log/slog"
)

const (
	// DefaultLoggerName is used if a logger is being stored in the context with no name.
	DefaultLoggerName = "__default__"
)

// activeLoggerNameContextKey is used to store the name of the active logger in a standard Go context object if
// there is more than one logger in the context.
type activeLoggerNameContextKey struct{}

// loggerContextKey is used to store a Logger object in a standard Go context object.
type loggerContextKey struct {
	name string
}

// ContextWithActiveLogger returns a new context with the given logger stored with the given name and saves
// it as the active logger.
//
// If multiple loggers are associated with a context, external libraries can use ActiveLoggerFromContext()
// to retrieve the appropriate logger to use for logging debug, error, etc. messages within their code.
//
// If no name is supplied, the default logger name is used.
func ContextWithActiveLogger(ctx context.Context, l *Logger, name string) context.Context {
	if name == "" {
		name = DefaultLoggerName
	}
	return ContextWithActiveLoggerName(ContextWithLogger(ctx, l, name), name)
}

// ContextWithActiveLoggerName returns a new context with the name of the active logger set.
//
// If multiple loggers are associated with a context, external libraries can use ActiveLoggerNameFromContext()
// to retrieve the appropriate logger name to then retrieve for logging debug, error, etc. messages within
// their code.
//
// If no name is supplied, the default logger name is used.
func ContextWithActiveLoggerName(ctx context.Context, name string) context.Context {
	if name == "" {
		name = DefaultLoggerName
	}
	return context.WithValue(ctx, activeLoggerNameContextKey{}, name)
}

// ContextWithLogger copies the given context and returns a new context with the given logger stored in it with
// the given name.
//
// If no name is supplied, the default logger name is used.
func ContextWithLogger(ctx context.Context, l *Logger, name string) context.Context {
	if name == "" {
		name = DefaultLoggerName
	}
	return context.WithValue(ctx, loggerContextKey{name: name}, l)
}

// Shutdown will cleanup any open resources or pending goroutines being run in the handler(s) attached to the logger.
func Shutdown(l *Logger, continueOnError bool) error {
	if l == nil {
		return nil
	}
	if h, ok := l.Handler().(ShutdownableHandler); ok {
		if err := h.Shutdown(continueOnError); err != nil && !continueOnError {
			return err
		}
	}
	return nil
}

// Logger is just a composition to be able to add functionality to the slog.Logger type.
type Logger struct {
	*slog.Logger

	// AdjustFrameCount indicates a number of frames to adjust the skip by when calling runtime.Callers. By default,
	// 3 frames are skipped when creating the record.
	AdjustFrameCount int

	// IncludeFileLine indicates whether or not to invoke runtime.Callers to get the program counter in order to retrieve
	// source file information.
	IncludeFileLine bool
}

// ActiveLoggerFromContext returns the active logger from the context, if it exists.
//
// If multiple loggers are associated with a context, external libraries can use this function to retrieve
// the the appropriate logger to retrieve for logging debug, error, etc. messages within their code.
//
// If no active logger is set in the context, the logger with the default logger name is returned. If no
// loggers are stored at all, the default global logger is returned.
//
// This function should *never* return a nil logger unless for some reason the default global logger has been
// erroneously set to nil.
func ActiveLoggerFromContext(ctx context.Context) *Logger {
	return LoggerFromContext(ctx, ActiveLoggerNameFromContext(ctx))
}

// ActiveLoggerNameFromContext retrieves the active logger name from the context, if it exists.
//
// If multiple loggers are associated with a context, external libraries can use this function to retrieve
// the name of the appropriate logger to retrieve for logging debug, error, etc. messages within their code.
//
// If no active logger name is set in the context, the default logger name is returned instead.
func ActiveLoggerNameFromContext(ctx context.Context) string {
	if v := ctx.Value(activeLoggerNameContextKey{}); v != nil {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return DefaultLoggerName
}

// Default returns the default logger object.
func Default() *Logger {
	return &Logger{
		Logger: slog.Default(),
	}
}

// LoggerFromContext retrieves the logger object stored in the given context with the given name, if it exists.
//
// If no name is supplied, the default logger name is used. If no matching logger can be found, the default
// global logger is returned instead.
//
// This function should *never* return a nil logger unless for some reason the default global logger has been
// erroneously set to nil.
func LoggerFromContext(ctx context.Context, name string) *Logger {
	if name == "" {
		name = DefaultLoggerName
	}
	if v := ctx.Value(loggerContextKey{name: name}); v != nil {
		if l, ok := v.(*Logger); ok {
			return l
		}
	}
	return Default()
}

// Nil returns a new "nil" logger which does not log anything, ever.
func Nil() *Logger {
	return &Logger{
		Logger: slog.New(newNilHandler()),
	}
}

// Wrap simply wraps the slog.Logger in an slogx.Logger object.
func Wrap(l *slog.Logger) *Logger {
	return &Logger{
		Logger: l,
	}
}

// Debug logs a message using DEBUG level.
func (l *Logger) Debug(msg string, args ...any) {
	l.log(context.Background(), LevelDebug, msg, args...)
}

// DebugContext logs a message using DEBUG level with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelDebug, msg, args...)
}

// Error logs a message using ERROR level.
func (l *Logger) Error(msg string, args ...any) {
	l.log(context.Background(), LevelError, msg, args...)
}

// ErrorContext logs a message using ERROR level with context.
func (l *Logger) ErrorlContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelError, msg, args...)
}

// Fatal logs a message using FATAL level.
func (l *Logger) Fatal(msg string, args ...any) {
	l.log(context.Background(), LevelFatal, msg, args...)
}

// FatalContext logs a message using FATAL level with context.
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelFatal, msg, args...)
}

// Info logs a message using INFO level.
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), LevelInfo, msg, args...)
}

// InfoContext logs a message using INFO level with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelInfo, msg, args...)
}

// Log simply logs a message at the given level.
func (l *Logger) Log(ctx context.Context, level Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

// LogAttrs is a more efficient way to log a message at any level while adding attributes.
func (l *Logger) LogAttrs(ctx context.Context, level Level, msg string, attrs ...slog.Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

// LogRecord simply logs the given pre-created record.
func (l *Logger) LogRecord(ctx context.Context, r slog.Record) {
	_ = l.Handler().Handle(ctx, r)
}

// Notice logs a message using NOTICE level.
func (l *Logger) Notice(msg string, args ...any) {
	l.log(context.Background(), LevelNotice, msg, args...)
}

// NoticeContext logs a message using NOTICE level with context.
func (l *Logger) NoticeContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelNotice, msg, args...)
}

// Panic logs a message using PANIC level.
func (l *Logger) Panic(msg string, args ...any) {
	l.log(context.Background(), LevelPanic, msg, args...)
}

// PanicContext logs a message using PANIC level with context.
func (l *Logger) PanicContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelPanic, msg, args...)
}

// Shutdown will cleanup any open resources or pending goroutines being run in the handler(s) attached to the logger.
func (l *Logger) Shutdown(continueOnError bool) error {
	if h, ok := l.Handler().(ShutdownableHandler); ok {
		if err := h.Shutdown(continueOnError); err != nil && !continueOnError {
			return err
		}
	}
	return nil
}

// Trace logs a message using TRACE level.
func (l *Logger) Trace(msg string, args ...any) {
	l.log(context.Background(), LevelTrace, msg, args...)
}

// TraceContext logs a message using TRACE level with context.
func (l *Logger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelTrace, msg, args...)
}

// Warn logs a message using WARN level.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), LevelWarn, msg, args...)
}

// With returns a new logger with the given attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger:           l.Logger.With(args...),
		AdjustFrameCount: l.AdjustFrameCount,
		IncludeFileLine:  l.IncludeFileLine,
	}
}

// WarnContext logs a message using WARN level with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelWarn, msg, args...)
}

// log is the low-level logging method for methods that take ...any.
func (l *Logger) log(ctx context.Context, level Level, msg string, args ...any) {
	if !l.Enabled(ctx, slog.Level(level)) {
		return
	}
	var pc uintptr
	if l.IncludeFileLine {
		var pcs [1]uintptr
		// skip runtime.Callers, this function, calling function
		runtime.Callers(3+l.AdjustFrameCount, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), slog.Level(level), msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
}

// logAttrs is like [Logger.log], but for methods that take ...Attr.
func (l *Logger) logAttrs(ctx context.Context, level Level, msg string, attrs ...slog.Attr) {
	if !l.Enabled(ctx, slog.Level(level)) {
		return
	}
	var pc uintptr
	if l.IncludeFileLine {
		var pcs [1]uintptr
		// skip runtime.Callers, this function, calling function
		runtime.Callers(3+l.AdjustFrameCount, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), slog.Level(level), msg, pc)
	r.AddAttrs(attrs...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Handler().Handle(ctx, r)
}
