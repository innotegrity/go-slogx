package formatter

import (
	"context"
	"time"

	"log/slog"

	"go.innotegrity.dev/slogx"
)

// BufferFormatter describes the interface a formatter which outputs a record to a buffer must implement.
type BufferFormatter interface {
	// FormatRecord should take the data from the record and format it as needed, storing it in the returned buffer.
	FormatRecord(context.Context, time.Time, slogx.Level, uintptr, string, []slog.Attr) (*slogx.Buffer, error)
}

// ColorBufferFormatter describes the interface a formatter which supports colorized text and outputs a record to
// a buffer must implement.
type ColorBufferFormatter interface {
	BufferFormatter

	// IsColorized should return whether or not the formatter uses colorized output.
	IsColorized() bool
}
