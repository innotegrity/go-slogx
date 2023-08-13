package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.innotegrity.dev/runtimex"
	"go.innotegrity.dev/slogx"
	"golang.org/x/exp/slog"
)

// FormatAttrFn is used to format the key and value for a particular attribute in the record.
//
// Attribute keys for groups are flattened and will start with "GROUP." where GROUP is the actual group name.
// Sub-groups are flattened in the same way. Essentially a period (.) indicates a group boundary.
//
// The value sent to the function is resolved before it is sent. If a LogValuer object is returned by this function,
// it is automatically resolved as well.
type FormatAttrFn func(ctx context.Context, attrKey string, attrValue slog.Value) (string, slog.Value, error)

// FormatLevelValueDefault is a default level formatter which simply shortens the level to 3 letters.
//
// Standard slogx levels are formatted using their ShortString() function while any other level is shortened to the
// first three characters and capitalized.
func FormatLevelValueDefault(ctx context.Context, level slogx.Level) (slog.Value, error) {
	var levelStr string
	switch level {
	case slogx.LevelTrace,
		slogx.LevelDebug,
		slogx.LevelInfo,
		slogx.LevelNotice,
		slogx.LevelWarn,
		slogx.LevelError,
		slogx.LevelFatal,
		slogx.LevelPanic:
		levelStr = level.ShortString()
	default:
		levelStr = strings.ToUpper(level.String())
		if len(levelStr) > 3 {
			levelStr = levelStr[0:3]
		}
	}
	return slog.StringValue(levelStr), nil
}

// FormatLevelValueFn is used to format the record's level.
//
// The value returned by this function is automatically resolved if a LogValuer object is returned.
type FormatLevelValueFn func(ctx context.Context, level slogx.Level) (slog.Value, error)

// FormatMessageValueFn is used to format the message.
//
// The value returned by this function is automatically resolved if a LogValuer object is returned.
type FormatMessageValueFn func(ctx context.Context, msg string) (slog.Value, error)

// FormatSourceValueDefault is a default source code location formatter which returns the location as
// filename:line.
func FormatSourceValueDefault(ctx context.Context, pc uintptr) (slog.Value, error) {
	frame := runtimex.FrameFromPC(pc)
	return slog.StringValue(fmt.Sprintf("%s", frame)), nil
}

// FormatSourceValueFn is used to format the source code location where the record was created.
//
// The value returned by this function is automatically resolved if a LogValuer object is returned.
type FormatSourceValueFn func(ctx context.Context, pc uintptr) (slog.Value, error)

// FormatTimeValueDefault is a default source code location formatter which returns the time as
// the UTC time in RFC3339 format.
func FormatTimeValueDefault(ctx context.Context, t time.Time) (slog.Value, error) {
	return slog.StringValue(t.UTC().Format(time.RFC3339)), nil
}

// FormatTimeValueFn is used to format the time the record was created.
//
// The value returned by this function is automatically resolved if a LogValuer object is returned.
type FormatTimeValueFn func(ctx context.Context, t time.Time) (slog.Value, error)
