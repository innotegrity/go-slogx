package slogx

import "golang.org/x/exp/slog"

// Typical log levels.
const (
	LevelTrace  = slog.Level(-8)
	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
	LevelPanic  = slog.Level(16)
)
