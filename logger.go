package slogx

import (
	"context"

	"golang.org/x/exp/slog"
)

// ShutdownableHandler describes a handler which is capable of being shutdown.
type ShutdownableHandler interface {
	slog.Handler
	Shutdown(bool) error
}

// Shutdown will cleanup any open resources or pending goroutines being run in the handler(s) attached to the logger.
func Shutdown(l *slog.Logger, continueOnError bool) error {
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
}

// Wrap simply wraps the slog.Logger in an slogx.Logger object.
func Wrap(l *slog.Logger) *Logger {
	return &Logger{
		Logger: l,
	}
}

// Fatal logs a message using FATAL level.
func (l *Logger) Fatal(msg string, args ...any) {
	l.Log(context.Background(), LevelFatal.Level(), msg, args...)
}

// FatalContext logs a message using FATAL level with context.
func (l *Logger) FatalContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelFatal.Level(), msg, args...)
}

// Notice logs a message using NOTICE level.
func (l *Logger) Notice(msg string, args ...any) {
	l.Log(context.Background(), LevelNotice.Level(), msg, args...)
}

// NoticeContext logs a message using NOTICE level with context.
func (l *Logger) NoticeContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelNotice.Level(), msg, args...)
}

// Panic logs a message using PANIC level.
func (l *Logger) Panic(msg string, args ...any) {
	l.Log(context.Background(), LevelPanic.Level(), msg, args...)
}

// PanicContext logs a message using PANIC level with context.
func (l *Logger) PanicContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelPanic.Level(), msg, args...)
}

// Trace logs a message using TRACE level.
func (l *Logger) Trace(msg string, args ...any) {
	l.Log(context.Background(), LevelTrace.Level(), msg, args...)
}

// TraceContext logs a message using TRACE level with context.
func (l *Logger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelTrace.Level(), msg, args...)
}
