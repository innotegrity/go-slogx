package handler

import (
	"context"
	"log/slog"
)

// nilHandler simply discards all messages.
type nilHandler struct{}

// NewNilHandler creates a new handler object.
func NewNilHandler() *nilHandler {
	return &nilHandler{}
}

// Enabled determines whether or not the given level is enabled for the handler.
//
// This function always returns false as no logging is ever done regardless of the level.
func (h nilHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return false
}

// Handle is responsible for writing the record to each and every handler.
//
// This function simply discards the record.
func (h *nilHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
//
// This function always just returns a new, default NilHandler object.
func (h nilHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewNilHandler()
}

// WithGroup creates a new handler from the existing one adding the given group to it.
//
// This function always just returns a new, default NilHandler object.
func (h nilHandler) WithGroup(name string) slog.Handler {
	return NewNilHandler()
}
