package slogx

import (
	"go.innotegrity.dev/errorx"
	"golang.org/x/exp/slog"
)

// Err returns an Attr for an error value.
func Err(key string, value error) slog.Attr {
	return slog.Attr{
		Key:   key,
		Value: slog.StringValue(value.Error()),
	}
}

// ErrX returns an Attr for an extended error value.
func ErrX(key string, value errorx.ExtendedError) slog.Attr {
	return slog.Group(
		key,
		slog.Int("code", value.Code()),
		slog.String("error", value.Error()),
		slog.String("internal_error", value.InternalError().Error()),
		slog.Any("properties", value.Properties()),
	)
}
