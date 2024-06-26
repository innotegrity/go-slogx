package slogx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"log/slog"
)

// Extended log levels in addition to the standard ones.
const (
	LevelUnknown  = Level(-2147483648)
	LevelMin      = Level(-2147483647)
	LevelTrace    = Level(-8)
	LevelDebug    = Level(slog.LevelDebug)
	LevelInfo     = Level(slog.LevelInfo)
	LevelNotice   = Level(2)
	LevelWarn     = Level(slog.LevelWarn)
	LevelError    = Level(slog.LevelError)
	LevelFatal    = Level(12)
	LevelPanic    = Level(16)
	LevelMax      = Level(2147483646)
	LevelDisabled = Level(2147483647)
)

// Level is just an integer
type Level int

// ParseLevel parses a level string into an actual value.
func ParseLevel(l string) (Level, error) {
	var level Level
	if err := level.UnmarshalText([]byte(l)); err != nil {
		return LevelUnknown, err
	}
	return level, nil
}

// Level returns the level itself in order to implement the `Leveler` interface.
func (l Level) Level() slog.Level {
	return slog.Level(l)
}

// MarshalJSON marshals the level into a JSON string.
func (l Level) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, l.String()), nil
}

// MarshalText marshals the level into a text string.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// ShortString returns a 3-character name for the level.
//
// If the level has a name, then that name in uppercase is returned. If the level is between named values, then
// an integer is appended to the uppercased name.

// Examples:
//
//	LevelWarn.String() => "WRN"
//	(LevelInfo+2).String() => "INF+2"
func (l Level) ShortString() string {
	str := func(base string, val Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l == LevelMin:
		return str("MIN", 0)
	case l == LevelMax:
		return str("MAX", 0)
	case l <= LevelTrace:
		return str("TRC", l-LevelTrace)
	case l <= LevelDebug:
		return str("DBG", l-LevelDebug)
	case l <= LevelInfo:
		return str("INF", l-LevelInfo)
	case l <= LevelNotice:
		return str("NOT", l-LevelNotice)
	case l <= LevelWarn:
		return str("WRN", l-LevelWarn)
	case l <= LevelError:
		return str("ERR", l-LevelError)
	case l <= LevelFatal:
		return str("FTL", l-LevelFatal)
	default:
		return str("PAN", l-LevelPanic)
	}
}

// String returns a name for the level.
//
// If the level has a name, then that name in uppercase is returned. If the level is between named values, then
// an integer is appended to the uppercased name.
//
// Examples:
//
//	LevelWarn.String() => "WARN"
//	(LevelInfo+2).String() => "INFO+2"
func (l Level) String() string {
	str := func(base string, val Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l == LevelMin:
		return str("MIN", 0)
	case l == LevelMax:
		return str("MAX", 0)
	case l <= LevelTrace:
		return str("TRACE", l-LevelTrace)
	case l <= LevelDebug:
		return str("DEBUG", l-LevelDebug)
	case l <= LevelInfo:
		return str("INFO", l-LevelInfo)
	case l <= LevelNotice:
		return str("NOTICE", l-LevelNotice)
	case l <= LevelWarn:
		return str("WARN", l-LevelWarn)
	case l <= LevelError:
		return str("ERROR", l-LevelError)
	case l <= LevelFatal:
		return str("FATAL", l-LevelFatal)
	default:
		return str("PANIC", l-LevelPanic)
	}
}

// UnmarshalJSON parses the given JSON string into the current level object.
func (l *Level) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return l.parse(str)
}

// UnmarshalText parses the given text string into the current level object.
func (l *Level) UnmarshalText(data []byte) error {
	return l.parse(string(data))
}

// parse handles parsing the given string into an appropriate level and storing it in the current object.
func (l *Level) parse(s string) error {
	var err error
	name := s
	offset := 0
	if i := strings.IndexAny(s, "+-"); i >= 0 {
		name = s[:i]
		offset, err = strconv.Atoi(s[i:])
		if err != nil {
			return fmt.Errorf("%s: failed to parse level: %s", s, err.Error())
		}
	}

	switch strings.ToUpper(name) {
	case "MIN", "UNK", "UNKNOWN":
		*l = LevelMin
	case "MAX", "DIS", "DISABLED":
		*l = LevelMax
	case "TRC", "TRACE":
		*l = LevelTrace
	case "DBG", "DEBUG":
		*l = LevelDebug
	case "INF", "INFO":
		*l = LevelInfo
	case "NOT", "NOTICE":
		*l = LevelNotice
	case "WRN", "WARN", "WARNING":
		*l = LevelWarn
	case "ERR", "ERROR":
		*l = LevelError
	case "FTL", "FATAL":
		*l = LevelFatal
	case "PAN", "PANIC":
		*l = LevelPanic
	default:
		return fmt.Errorf("%s: unknown level", name)
	}
	*l += Level(offset)
	return nil
}

// A LevelVar is a Level variable, to allow a Handler level to change dynamically.
//
// It implements Leveler as well as a Set method, and it is safe for use by multiple goroutines. The zero LevelVar
// corresponds to LevelInfo.
type LevelVar struct {
	val atomic.Int64
}

// NewLevelVar returns a new object with the given level set.
func NewLevelVar(l Level) *LevelVar {
	lv := &LevelVar{}
	lv.val.Store(int64(l))
	return lv
}

// Level returns v's level.
func (v *LevelVar) Level() Level {
	return Level(int(v.val.Load()))
}

// Set sets v's level to l.
func (v *LevelVar) Set(l Level) {
	v.val.Store(int64(l))
}

func (v *LevelVar) String() string {
	return fmt.Sprintf("LevelVar(%s)", v.Level())
}

// MarshalText implements [encoding.TextMarshaler]
// by calling [Level.MarshalText].
func (v *LevelVar) MarshalText() ([]byte, error) {
	return v.Level().MarshalText()
}

// UnmarshalText implements [encoding.TextUnmarshaler]
// by calling [Level.UnmarshalText].
func (v *LevelVar) UnmarshalText(data []byte) error {
	var l Level
	if err := l.UnmarshalText(data); err != nil {
		return err
	}
	v.Set(l)
	return nil
}
