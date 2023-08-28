package handlers

import (
	"context"
	"io"
	"os"
	"sync"

	"go.innotegrity.dev/generic"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/formatters"
	"go.innotegrity.dev/slogx/internal/buffer"
	"go.innotegrity.dev/slogx/internal/utils"
	"golang.org/x/exp/slog"
)

// JSONHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type JSONHandlerOptionsContext struct{}

// JSONHandlerOptions holds the options for the JSON handler.
type JSONHandlerOptions struct {
	// Level is the minimum log level to write to the handler.
	//
	// By default, the level will be set to slog.LevelInfo if not supplied.
	Level slog.Leveler

	// RecordFormatter specifies the formatter to use to format the record before writing it to the writer.
	//
	// If no formatter is supplied, formatters.DefaultJSONFormatter is used to format the output.
	RecordFormatter formatters.BufferFormatter

	// Writer is where to write the output to.
	//
	// By default, messages are written to os.Stdout if not supplied.
	Writer io.Writer
}

// jsonHandler is a log handler that writes records to an io.Writer using standard JSON formatting.
type jsonHandler struct {
	activeGroup string
	attrs       []slog.Attr
	groups      []string
	options     JSONHandlerOptions
	writeLock   *sync.Mutex
}

// NewJSONHandler creates a new handler object.
func NewJSONHandler(opts JSONHandlerOptions) *jsonHandler {
	// set default options
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	// create the handler
	return &jsonHandler{
		attrs:     []slog.Attr{},
		groups:    []string{},
		options:   opts,
		writeLock: &sync.Mutex{},
	}
}

// Enabled determines whether or not the given level is enabled in this handler.
func (h jsonHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

// Handle actually handles writing the record to the output writer.
//
// Any attributes duplicated between the handler and record, including within groups, are automaticlaly removed.
// If a duplicate is encountered, the last value found will be used for the attribute's value.
func (h *jsonHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := context.WithValue(ctx, JSONHandlerOptionsContext{}, &h.options)
	attrs := utils.ConsolidateAttrs(h.attrs, h.activeGroup, r)

	// format the output into a buffer
	var buf *buffer.Buffer
	var err error
	if h.options.RecordFormatter != nil {
		buf, err = h.options.RecordFormatter.FormatRecord(handlerCtx, r.Time, slogx.Level(r.Level), r.PC, r.Message,
			attrs)
	} else {
		f := formatters.DefaultJSONFormatter()
		buf, err = f.FormatRecord(handlerCtx, r.Time, slogx.Level(r.Level), r.PC, r.Message, attrs)
	}
	if err != nil {
		return err
	}

	// write the buffer to the output
	h.writeLock.Lock()
	defer h.writeLock.Unlock()
	_, err = h.options.Writer.Write(buf.Bytes())
	return err
}

// Shutdown is responsible for cleaning up resources used by the handler.
func (h jsonHandler) Shutdown(continueOnError bool) error {
	if w, ok := h.options.Writer.(io.WriteCloser); ok {
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h jsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &jsonHandler{
		attrs:     h.attrs,
		groups:    h.groups,
		options:   h.options,
		writeLock: h.writeLock,
	}
	if h.activeGroup == "" {
		newHandler.attrs = append(newHandler.attrs, attrs...)
	} else {
		newHandler.attrs = append(newHandler.attrs, slog.Group(h.activeGroup, generic.AnySlice(attrs)...))
		newHandler.activeGroup = h.activeGroup
	}
	return newHandler
}

// WithGroup creates a new handler from the existing one adding the given group to it.
func (h *jsonHandler) WithGroup(name string) slog.Handler {
	newHandler := &jsonHandler{
		attrs:     h.attrs,
		groups:    h.groups,
		options:   h.options,
		writeLock: h.writeLock,
	}
	if name != "" {
		newHandler.groups = append(newHandler.groups, name)
		newHandler.activeGroup = name
	}
	return newHandler
}
