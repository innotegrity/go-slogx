package formatter

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
// The group name will be an empty string for attributes not nested within a group. Otherwise, the group will
// be populated with the name of the group to which the attribute belongs. If an attribue is nested in multiple
// groups, each group is separated by a single period (.) character (eg: group1.group2.group3).
//
// The function should return the formatted key and value. The value passed to this function should be resolved by
// the handler prior to being passed. The value returned by the function will not be resolved again, so if it needs
// resolving, it should be done prior to returning it.
type FormatAttrFn func(context.Context, slog.Leveler, string, string, slog.Value) (string, slog.Value, error)

// FormatLevelValueDefault is a default level formatter which simply shortens the level to 3 letters.
//
// Standard slogx levels are formatted using their ShortString() function while any other level is shortened to the
// first three characters and capitalized.
func FormatLevelValueDefault(ctx context.Context, level slog.Leveler) (string, error) {
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
		levelStr = slogx.Level(level.Level()).ShortString()
	default:
		levelStr = strings.ToUpper(level.Level().String())
		if len(levelStr) > 3 {
			levelStr = levelStr[0:3]
		}
	}
	return levelStr, nil
}

// FormatLevelValueFn is used to format the record's level.
type FormatLevelValueFn func(context.Context, slog.Leveler) (string, error)

// FormatMessageValueFn is used to format the message.
type FormatMessageValueFn func(context.Context, slog.Leveler, string) (string, error)

// FormatSourceValueDefault is a default source code location formatter which returns the location as
// filename:line.
func FormatSourceValueDefault(ctx context.Context, level slog.Leveler, pc uintptr) (string, error) {
	frame := runtimex.FrameFromPC(pc)
	return fmt.Sprintf("%s", frame), nil
}

// FormatSourceValueFn is used to format the source code location where the record was created.
type FormatSourceValueFn func(context.Context, slog.Leveler, uintptr) (string, error)

// FormatTimeValueDefault is a default source code location formatter which returns the time as
// the UTC time in RFC3339 format.
func FormatTimeValueDefault(ctx context.Context, level slog.Leveler, t time.Time) (string, error) {
	return t.UTC().Format(time.RFC3339), nil
}

// FormatTimeValueFn is used to format the time the record was created.
type FormatTimeValueFn func(context.Context, slog.Leveler, time.Time) (string, error)
