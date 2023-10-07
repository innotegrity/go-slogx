package slogx

import "log/slog"

// DynamicLevelHandler is implemented by handlers which can have their level updated without being recreated.
type DynamicLevelHandler interface {
	slog.Handler

	// Level should return the current logging level that is in use by the handler.
	//
	// For handlers with multiple sub-handlers, the levels should all typically be identical.
	// If the levels are not identical, it is up to the handler's implementation on what level
	// to return.
	Level() Level

	// WithLevel returns a new handler with the given logging level set.
	WithLevel(l Level) DynamicLevelHandler
}

// ShutdownableHandler should be implemented by handlers which allocate resources that need cleaning up before an
// application exits.
type ShutdownableHandler interface {
	slog.Handler

	// Shutdown should handle cleaning up any resources created by the handler (eg: closing file handles,
	// DB connections, etc.).
	Shutdown(bool) error
}
