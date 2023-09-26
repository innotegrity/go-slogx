package handler

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/mattn/go-colorable"
	"go.innotegrity.dev/generic"
	"go.innotegrity.dev/slogx"
	"go.innotegrity.dev/slogx/formatter"
	"golang.org/x/exp/slog"
)

// consoleHandlerOptionsContext can be used to retrieve the options used by the handler from the context.
type consoleHandlerOptionsContext struct{}

// ConsoleHandlerOptions holds the options for the console handler.
type ConsoleHandlerOptions struct {
	// Level is the minimum log level to write to the handler.
	//
	// By default, the level will be set to slog.LevelInfo if not supplied.
	Level slog.Leveler

	// RecordFormatter specifies the formatter to use to format the record before writing it to the writer.
	//
	// If no formatter is supplied, a colorized formatter.DefaultConsoleFormatter is used to format the output.
	RecordFormatter formatter.ColorBufferFormatter

	// Writer is where to write the output to.
	//
	// By default, messages are written to os.Stdout if not supplied.
	Writer io.Writer
}

// DefaultConsoleHandlerOptions returns a default set of options for the handler.
func DefaultConsoleHandlerOptions() ConsoleHandlerOptions {
	return ConsoleHandlerOptions{
		Level:           slog.LevelInfo,
		RecordFormatter: formatter.DefaultConsoleFormatter(true),
		Writer:          os.Stdout,
	}
}

// GetConsoleHandlerOptionsFromContext retrieves the options from the context.
//
// If the options are not set in the context, a set of default options is returned instead.
func GetConsoleHandlerOptionsFromContext(ctx context.Context) *ConsoleHandlerOptions {
	o := ctx.Value(consoleHandlerOptionsContext{})
	if o != nil {
		if opts, ok := o.(*ConsoleHandlerOptions); ok {
			return opts
		}
	}
	opts := DefaultConsoleHandlerOptions()
	return &opts
}

// AddToContext adds the options to the given context and returns the new context.
func (o *ConsoleHandlerOptions) AddToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, consoleHandlerOptionsContext{}, o)
}

// consoleHandler is a log handler that writes records to an io.Writer, typically a console in a specified format.
type consoleHandler struct {
	activeGroup string
	attrs       []slog.Attr
	groups      []string
	options     ConsoleHandlerOptions
	writeLock   *sync.Mutex
}

// NewConsoleHandler creates a new handler object.
func NewConsoleHandler(opts ConsoleHandlerOptions) *consoleHandler {
	// set default options
	if opts.Level == nil {
		opts.Level = slogx.LevelInfo
	}
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if (opts.Writer == os.Stdout || opts.Writer == os.Stderr) && opts.RecordFormatter != nil &&
		opts.RecordFormatter.IsColorized() {
		opts.Writer = colorable.NewColorable(opts.Writer.(*os.File))
	}

	// create the handler
	return &consoleHandler{
		attrs:     []slog.Attr{},
		groups:    []string{},
		options:   opts,
		writeLock: &sync.Mutex{},
	}
}

// Enabled determines whether or not the given level is enabled in this handler.
func (h consoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

// Handle actually handles writing the record to the output writer.
//
// Any attributes duplicated between the handler and record, including within groups, are automaticlaly removed.
// If a duplicate is encountered, the last value found will be used for the attribute's value.
func (h *consoleHandler) Handle(ctx context.Context, r slog.Record) error {
	handlerCtx := h.options.AddToContext(ctx)
	attrs := slogx.ConsolidateAttrs(h.attrs, h.activeGroup, r)

	// format the output into a buffer
	var buf *slogx.Buffer
	var err error
	if h.options.RecordFormatter != nil {
		buf, err = h.options.RecordFormatter.FormatRecord(handlerCtx, r.Time, slogx.Level(r.Level), r.PC, r.Message,
			attrs)
	} else {
		f := formatter.DefaultConsoleFormatter(true)
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
func (h consoleHandler) Shutdown(continueOnError bool) error {
	if w, ok := h.options.Writer.(io.WriteCloser); ok {
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}

// WithAttrs creates a new handler from the existing one adding the given attributes to it.
func (h consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &consoleHandler{
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
func (h *consoleHandler) WithGroup(name string) slog.Handler {
	newHandler := &consoleHandler{
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
