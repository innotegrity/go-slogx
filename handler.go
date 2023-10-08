package slogx

import "log/slog"

// ShutdownableHandler should be implemented by handlers which allocate resources that need cleaning up before an
// application exits.
type ShutdownableHandler interface {
	slog.Handler

	// Shutdown should handle cleaning up any resources created by the handler (eg: closing file handles,
	// DB connections, etc.).
	Shutdown(bool) error
}
