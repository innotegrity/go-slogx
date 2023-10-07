package slogx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"log/slog"
)

// Extended log levels in addition to the standard ones.
const (
	LevelNone   = Level(-2147483648)
	LevelTrace  = Level(-8)
	LevelDebug  = Level(slog.LevelDebug)
	LevelInfo   = Level(slog.LevelInfo)
	LevelNotice = Level(2)
	LevelWarn   = Level(slog.LevelWarn)
	LevelError  = Level(slog.LevelError)
	LevelFatal  = Level(12)
	LevelPanic  = Level(16)
)

// Level is just like slog.Level - an integer
type Level int

// Level returns the slog.Level equivalent level.
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

// ShortString returns the level as a 3 character string.
func (l Level) ShortString() string {
	str := func(base string, val Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l == LevelNone:
		return str("NON", 0)
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

// String returns the level as a string.
func (l Level) String() string {
	str := func(base string, val Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l == LevelNone:
		return str("NONE", 0)
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
	case "NONE":
		*l = LevelNone
	case "TRACE":
		*l = LevelTrace
	case "DEBUG":
		*l = LevelDebug
	case "INFO":
		*l = LevelInfo
	case "NOTICE":
		*l = LevelNotice
	case "WARN", "WARNING":
		*l = LevelWarn
	case "ERR", "ERROR":
		*l = LevelError
	case "FATAL":
		*l = LevelFatal
	case "PANIC":
		*l = LevelPanic
	default:
		return fmt.Errorf("%s: unknown level", name)
	}
	*l += Level(offset)
	return nil
}
